CREATE TABLE IF NOT EXISTS "chat_groups" (
    "id" INTEGER PRIMARY KEY,
    "date" INTEGER NOT NULL,
    "title" TEXT NOT NULL UNIQUE,
    "image" TEXT,
    "members" TEXT);