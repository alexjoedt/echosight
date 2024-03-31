DROP EXTENSION "uuid-ossp";
DROP EXTENSION "citext";
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS hosts;
DROP TABLE IF EXISTS detectors;
DROP TYPE host_address_type;
DROP TABLE IF EXISTS sessions;
DROP INDEX IF EXISTS sessions_expiry_idx;
