CREATE TABLE IF NOT EXISTS "comments" (
    "id" INTEGER PRIMARY KEY,
    "date" INTEGER NOT NULL,
    "user_id"  INTEGER NOT NULL,
    "post_id"  INTEGER NOT NULL,
    "content"  TEXT NOT NULL,
    "image" TEXT);