CREATE TABLE IF NOT EXISTS "events" (
    "id" INTEGER PRIMARY KEY,
    "creator_id" INTEGER NOT NULL,
    "create_date" INTEGER NOT NULL,
    "event_date" INTEGER NOT NULL,
    "image" TEXT,
    "title" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "members" TEXT,
    "group_id" TEXT);