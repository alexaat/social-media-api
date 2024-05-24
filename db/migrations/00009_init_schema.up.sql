CREATE TABLE IF NOT EXISTS "join_group_requests" (
    "id" INTEGER PRIMARY KEY,
    "date" INTEGER NOT NULL,
    "group_id"  INTEGER NOT NULL,
    "member_id"  INTEGER NOT NULL,
    CONSTRAINT unq UNIQUE (group_id, member_id));