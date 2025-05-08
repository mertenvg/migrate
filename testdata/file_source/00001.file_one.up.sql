CREATE TABLE IF NOT EXISTS "public"."migrations" (
    "name" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "down" TEXT NULL,
    CONSTRAINT "migrations_pkey" PRIMARY KEY ("name")
);