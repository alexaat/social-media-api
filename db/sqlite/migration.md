Create set:
migrate create -ext sql -dir server/db/migrations/sqlite -seq init_schema
Upgrade db:
migrate -source file://db/migrations/sqlite -database sqlite3://db/sqlite/social.db up
Downgrade db:
migrate -source file://db/migrations/sqlite -database sqlite3://db/sqlite/social.db down