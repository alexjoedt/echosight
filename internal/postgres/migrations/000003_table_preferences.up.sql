CREATE TABLE IF NOT EXISTS preferences (
  id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
  lookup_version bigint NOT NULL DEFAULT 1,
  name varchar UNIQUE NOT NULL,
  value varchar NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00'
)