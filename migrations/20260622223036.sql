-- Create "videos" table
CREATE TABLE "public"."videos" (
  "id" uuid NOT NULL,
  "title" character varying NOT NULL,
  "description" character varying NOT NULL DEFAULT '',
  "type" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'PendingUpload',
  "qualities" jsonb NOT NULL,
  "uploaded_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
