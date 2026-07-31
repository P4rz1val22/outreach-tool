-- 000002_add_user_scoping.down.sql
ALTER TABLE tags DROP CONSTRAINT tags_user_id_name_key;
ALTER TABLE tags DROP COLUMN user_id;
ALTER TABLE tags ADD CONSTRAINT tags_name_key UNIQUE (name);

ALTER TABLE contacts DROP COLUMN user_id;