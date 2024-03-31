CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;


CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    first_name varchar,
    last_name varchar,
    email citext UNIQUE NOT NULL, -- case-insensitive text
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    role varchar NOT NULL DEFAULT 'regular',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(), 
    updated_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00',
    deleted_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00', 
    lookup_version integer NOT NULL DEFAULT 1
);

-- Insert Admin user
-- default pass: echosight
INSERT INTO users (id,first_name, email, password_hash, activated, role) 
VALUES ('00000000-0000-0000-0000-000000000001','Admin', 'admin@echo.sight','JDJhJDEyJDhmUjUyT0FuS3EvMDd4SVNJazRTaHU5b2ZmaEp0d0JFSVNUbUM0SDUwUmdRcmpzV0pmYUNh',true,'admin');

CREATE TYPE host_address_type AS ENUM ('ipv4','ipv6');

-- Session Table
CREATE TABLE sessions (
    hash bytea PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users ON DELETE CASCADE,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- Host Table
CREATE TABLE hosts (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    lookup_version bigint NOT NULL DEFAULT 1,
    name varchar UNIQUE NOT NULL,
    address_type varchar,
    address varchar,
    location varchar,
    os varchar,
    agent BOOLEAN NOT NULL DEFAULT false,
    active BOOLEAN NOT NULL DEFAULT false,
    state varchar NOT NULL,
    status_message varchar,
    last_checked_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00',
    tags TEXT[],
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(), 
    updated_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00'
);


-- Detector Table
CREATE TABLE detectors (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host_id         UUID NOT NULL,
    name            varchar NOT NULL,
    host_name       varchar NOT NULL,
    lookup_version  bigint NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT false,
    type            varchar NOT NULL,
    timeout         varchar,
    interval        varchar,
    tags            TEXT[] NOT NULL,
    state           varchar NOT NULL,
    status_message  TEXT,
    config          JSONB, -- JSONB für flexible, strukturierte Konfigurationsdaten
    last_checked_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(), 
    updated_at timestamp(0) with time zone NOT NULL DEFAULT '1900-01-01 00:00:00+00'
);

ALTER TABLE detectors ADD FOREIGN KEY (host_id) REFERENCES hosts (id) ON DELETE CASCADE;

-- TODO: permissions
-- ALTER TABLE permissions_host_user ADD FOREIGN KEY (host_id) REFERENCES hosts (id) ON DELETE CASCADE;
-- ALTER TABLE permissions_detecotr_user ADD FOREIGN KEY (detecotr_id) REFERENCES detectors (id) ON DELETE CASCADE;