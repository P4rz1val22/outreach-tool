-- 000002_add_user_scoping.up.sql
ALTER TABLE contacts ADD COLUMN user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE;
CREATE INDEX idx_contacts_user_id ON contacts(user_id);

ALTER TABLE tags DROP CONSTRAINT tags_name_key;
ALTER TABLE tags ADD COLUMN user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE;
ALTER TABLE tags ADD CONSTRAINT tags_user_id_name_key UNIQUE (user_id, name);
CREATE INDEX idx_tags_user_id ON tags(user_id);