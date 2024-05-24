CREATE TABLE IF NOT EXISTS "notifications" (
    "id" INTEGER PRIMARY KEY,
    "date" INTEGER NOT NULL,
    "type" TEXT NOT NULL,
    "content" TEXT NOT NULL,
    "sender_id" INTEGER NOT NULL,
    "recipient_id" INTEGER NOT NULL,
    "group_id" INTEGER,
    "event_id" INTEGER,
    "is_read" BOOLEAN NOT NULL
)
