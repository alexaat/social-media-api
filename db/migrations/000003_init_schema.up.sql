CREATE TABLE IF NOT EXISTS "posts" (
    "id" INTEGER PRIMARY KEY,
    "date" INTEGER NOT NULL,
    "user_id" INTEGER NOT NULL,
    "content" TEXT NOT NULL,
    "privacy" TEXT NOT NULL,
    "specific_friends" TEXT,
    "image" TEXT,
    "group_id" INTEGER
);