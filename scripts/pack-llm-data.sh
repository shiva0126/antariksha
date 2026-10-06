#!/usr/bin/env bash
# Builds the Astrisk model training data and packs it for a free GPU
# notebook (Google Colab or Kaggle): training/astrisk_finetune.ipynb.
# Output: .runtime/llm-data/astrisk-train.zip
set -euo pipefail
cd "$(dirname "$0")/.."
GO=${GO:-$HOME/.local/toolchains/go/bin/go}
export CGO_ENABLED=1
$GO run ./cmd/corpus build >/dev/null
$GO run ./cmd/llmdata books
$GO run ./cmd/llmdata bookqa
$GO run ./cmd/llmdata gen -n 4000 -out .runtime/llm-data/train.jsonl
$GO run ./cmd/llmdata compact -n 3000
$GO run ./cmd/bookbench -n 2000 >/dev/null
cd .runtime/llm-data
python3 -c "import zipfile; z = zipfile.ZipFile('astrisk-train.zip', 'w', zipfile.ZIP_DEFLATED); [z.write(f) for f in ['books.jsonl', 'book-qa.jsonl', 'train.jsonl', 'compact.jsonl', 'book-bench.jsonl']]"
ls -la astrisk-train.zip
