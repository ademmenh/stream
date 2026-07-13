ALTER TABLE "public"."videos"
  ADD COLUMN "raw_path" character varying NOT NULL DEFAULT '',
  ADD COLUMN "photo_path" character varying NOT NULL DEFAULT '';
