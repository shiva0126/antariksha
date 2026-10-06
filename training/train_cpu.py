"""Fine-tune a small Astrisk model on this machine's CPU (no GPU needed).

Two stages with LoRA on Qwen3-0.6B (Apache-2.0):
  books  read passages of the public-domain classics (.runtime/llm-data/books.jsonl)
  sft    answer like Astrisk: questions on every checked book rule
         (book-qa.jsonl) and compressed chart questions (compact.jsonl)

Then the adapter is merged and saved as a Hugging Face model for GGUF export.

  PYTHONPATH=.runtime/train-deps python3 training/train_cpu.py --stage books --minutes 60
  PYTHONPATH=.runtime/train-deps python3 training/train_cpu.py --stage sft --minutes 150
  PYTHONPATH=.runtime/train-deps python3 training/train_cpu.py --stage merge

Each stage stops after --minutes (the data is shuffled, so a time budget
trains on a representative sample) and saves its adapter; the next stage
continues from it.
"""
import argparse, json, math, os, random, time

import torch
from peft import LoraConfig, PeftModel, get_peft_model
from transformers import AutoModelForCausalLM, AutoTokenizer

BASE = ".runtime/llm-train/base"  # Qwen/Qwen3-0.6B, downloaded from Hugging Face
DATA = ".runtime/llm-data"
OUT = ".runtime/llm-train"
BOOK_SYSTEM = 'You are Astrisk. Answer in plain language and cite the classical source. Return JSON only: {"answer": string}.'


def log(msg):
    line = time.strftime("%H:%M:%S ") + msg
    print(line, flush=True)
    with open(os.path.join(OUT, "train.log"), "a") as f:
        f.write(line + "\n")


def read(name):
    with open(os.path.join(DATA, name)) as f:
        return [json.loads(l) for l in f]


def book_examples(tok, max_len):
    out = []
    for p in read("books.jsonl"):
        who = f"tr. {p['translator']}, " if p.get("translator") else ""
        text = f"{p['title']} ({who}{p['year']}) {p['ref']}\n{p['text']}{tok.eos_token}"
        ids = tok(text, add_special_tokens=False)["input_ids"][:max_len]
        out.append((ids, list(ids)))  # learn every token of the book text
    return out


def chat_examples(tok, max_len):
    # Book rules first (every one, once, shuffled), then compressed chart
    # questions: on a CPU the time budget may end before the second part.
    qa, charts = read("book-qa.jsonl"), read("compact.jsonl")
    random.shuffle(qa)
    random.shuffle(charts)
    out = []
    for r in qa + charts:
        msgs = r["messages"]
        prompt = tok.apply_chat_template(msgs[:-1], tokenize=False, add_generation_prompt=True, enable_thinking=False)
        full = prompt + msgs[-1]["content"] + "<|im_end|>\n"
        p_ids = tok(prompt, add_special_tokens=False)["input_ids"]
        ids = tok(full, add_special_tokens=False)["input_ids"]
        if len(ids) > max_len:
            continue
        labels = [-100] * len(p_ids) + ids[len(p_ids):]  # learn only the answer
        out.append((ids, labels))
    return out


def answer_loss(model, x, y):
    """Next-token loss on the labelled positions only. The vocabulary has
    ~152k entries, so logits for every position of a 700-token example would
    take ~425 MB (plus gradients); only the answer's ~200 positions need them."""
    inner = model.get_base_model()
    hidden = inner.model(input_ids=x).last_hidden_state[:, :-1]
    target = y[:, 1:]
    keep = target != -100
    logits = inner.lm_head(hidden[keep])
    return torch.nn.functional.cross_entropy(logits.float(), target[keep])


def train(stage, minutes, max_len, threads, lr):
    torch.set_num_threads(threads)
    torch.manual_seed(7)
    random.seed(7)
    tok = AutoTokenizer.from_pretrained(BASE)
    dtype = torch.bfloat16 if os.environ.get("TRAIN_DTYPE") == "bf16" else torch.float32
    model = AutoModelForCausalLM.from_pretrained(BASE, dtype=dtype)
    if os.environ.get("TRAIN_CHECKPOINTING") == "1":  # saves memory, costs ~25% speed
        model.gradient_checkpointing_enable()
        model.enable_input_require_grads()
    prev = os.path.join(OUT, "adapter-books")
    if stage == "sft" and os.path.isdir(prev):
        model = PeftModel.from_pretrained(model, prev, is_trainable=True)
        log(f"continuing from {prev}")
    else:
        cfg = LoraConfig(r=16, lora_alpha=32, lora_dropout=0.0, task_type="CAUSAL_LM",
                         target_modules=["q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"])
        model = get_peft_model(model, cfg)
    model.print_trainable_parameters()
    data = book_examples(tok, max_len) if stage == "books" else chat_examples(tok, max_len)
    if stage == "books":
        random.shuffle(data)
    log(f"stage {stage}: {len(data)} examples, budget {minutes} min, {threads} threads")
    opt = torch.optim.AdamW([p for p in model.parameters() if p.requires_grad], lr=lr, weight_decay=0.0)
    accum, step, seen_tokens, loss_sum, n_loss = 8, 0, 0, 0.0, 0
    est_steps = max(1, len(data) // accum)
    start = time.time()
    model.train()
    for i, (ids, labels) in enumerate(data):
        x = torch.tensor([ids])
        y = torch.tensor([labels])
        loss = answer_loss(model, x, y) / accum
        loss.backward()
        loss_sum += loss.item() * accum
        n_loss += 1
        seen_tokens += len(ids)
        if i < 4:
            log(f"example {i}: {len(ids)} tokens, {time.time() - start:.0f}s since start")
        if (i + 1) % accum == 0:
            # Cosine decay over the steps the time budget is likely to allow.
            for g in opt.param_groups:
                g["lr"] = lr * (0.1 + 0.9 * 0.5 * (1 + math.cos(math.pi * min(1.0, step / est_steps))))
            torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
            opt.step()
            opt.zero_grad()
            step += 1
            elapsed = time.time() - start
            if step == 3:
                # Re-estimate how many steps fit in the budget from the real speed.
                est_steps = min(len(data) // accum, int(minutes * 60 / (elapsed / step)))
                log(f"speed {seen_tokens / elapsed:.0f} tokens/s; about {est_steps} steps fit the budget")
            if step % 10 == 0:
                log(f"step {step} loss {loss_sum / n_loss:.3f} tokens {seen_tokens} elapsed {elapsed / 60:.1f} min")
                loss_sum, n_loss = 0.0, 0
            if step % 50 == 0:
                model.save_pretrained(os.path.join(OUT, f"adapter-{stage}"))
            if elapsed > minutes * 60:
                break
    model.save_pretrained(os.path.join(OUT, f"adapter-{stage}"))
    log(f"stage {stage} done: {step} steps, {seen_tokens} tokens, {(time.time() - start) / 60:.1f} min")


def merge():
    tok = AutoTokenizer.from_pretrained(BASE)
    model = AutoModelForCausalLM.from_pretrained(BASE, torch_dtype=torch.float32)
    # The sft adapter continues the books adapter, so merge only the latest.
    for stage in ("sft", "books"):
        path = os.path.join(OUT, f"adapter-{stage}")
        if os.path.isdir(path):
            model = PeftModel.from_pretrained(model, path).merge_and_unload()
            log(f"merged {path}")
            break
    model = model.to(torch.bfloat16)
    model.save_pretrained(os.path.join(OUT, "merged"), safe_serialization=True)
    tok.save_pretrained(os.path.join(OUT, "merged"))
    log("saved merged model")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--stage", choices=["books", "sft", "merge"], required=True)
    ap.add_argument("--minutes", type=float, default=60)
    ap.add_argument("--max-len", type=int, default=768)
    ap.add_argument("--threads", type=int, default=3)
    ap.add_argument("--lr", type=float, default=2e-4)
    a = ap.parse_args()
    os.makedirs(OUT, exist_ok=True)
    if a.stage == "merge":
        merge()
    else:
        train(a.stage, a.minutes, a.max_len, a.threads, a.lr)
