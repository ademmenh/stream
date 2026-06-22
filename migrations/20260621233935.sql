-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL,
  "name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "phone" character varying NULL,
  "password_hash" character varying NOT NULL,
  "role" character varying NOT NULL DEFAULT 'client',
  "banned" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "public"."users" ("email");
