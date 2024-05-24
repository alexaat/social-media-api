CREATE TABLE IF NOT EXISTS "followers" (
    "id" INTEGER PRIMARY KEY,
    "date" INTEGER NOT NULL,
    "follower" INTEGER NOT NULL,
    "followee" INTEGER NOT NULL,
    "approved" BOOLEAN NOT NULL
)
