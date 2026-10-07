-- Machine-translated answers keep the English original alongside, so the
-- member can switch back to it.
ALTER TABLE chat_messages ADD COLUMN original TEXT NOT NULL DEFAULT '', ADD COLUMN lang TEXT NOT NULL DEFAULT '';
