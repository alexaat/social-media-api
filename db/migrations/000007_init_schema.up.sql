CREATE TABLE IF NOT EXISTS "groups" (
    "id" INTEGER PRIMARY KEY,
    "creator_id" INTEGER NOT NULL,
    "date" INTEGER NOT NULL,
    "title" TEXT NOT NULL UNIQUE,
    "description" TEXT NOT NULL,
    "members" TEXT);