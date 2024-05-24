CREATE TABLE IF NOT EXISTS "session" (
    "id" INTEGER PRIMARY KEY,
    "user_id" INTEGER NOT NULL UNIQUE,
    "session_id" TEXT);