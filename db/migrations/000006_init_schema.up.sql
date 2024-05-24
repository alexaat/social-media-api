CREATE TABLE IF NOT EXISTS "messages" (
    "id" INTEGER PRIMARY KEY,
    "sender_id" INTEGER NOT NULL,
    "recipient_id" INTEGER,
    "chat_group_id" INTEGER,
    "date" INTEGER NOT NULL,
    "content" TEXT NOT NULL,
    "is_read" BOOLEAN NOT NULL,
    "read_by" TEXT)