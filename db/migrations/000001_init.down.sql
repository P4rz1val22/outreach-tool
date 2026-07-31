-- Drop in reverse dependency order (children before parents)

DROP TABLE IF EXISTS check_ins;
DROP TABLE IF EXISTS thread_tags;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS contact_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS contacts;

DROP TYPE IF EXISTS checkin_status;
DROP TYPE IF EXISTS thread_status;
DROP TYPE IF EXISTS contact_status;