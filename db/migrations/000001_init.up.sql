-- Enums

CREATE TYPE contact_status AS ENUM ('active', 'archived');
CREATE TYPE thread_status AS ENUM ('active', 'archived');
CREATE TYPE checkin_status AS ENUM ('pending', 'completed', 'skipped', 'missed');

-- Contact: identity only, no cadence/scheduling here

CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    role VARCHAR(255),
    status contact_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tag: normalized, many-to-many with Contact

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE contact_tags (
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (contact_id, tag_id)
);

CREATE INDEX idx_contact_tags_tag_id ON contact_tags(tag_id);

-- Campaign: stretch goal, but stubbed now for schema compatibility

CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    purpose TEXT,
    start_date DATE,
    pacing_rule VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Thread: the specific reason for contact; cadence lives here, not on Contact

CREATE TABLE threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    label VARCHAR(255) NOT NULL,
    cadence_interval_days INTEGER, -- NULL = one-off thread, auto-archives after its single CheckIn resolves
    status thread_status NOT NULL DEFAULT 'active',
    campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL,
    email_enabled BOOLEAN, -- NULL = inherit global default
    push_enabled BOOLEAN,  -- NULL = inherit global default
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_threads_contact_id ON threads(contact_id);
CREATE INDEX idx_threads_campaign_id ON threads(campaign_id);

-- Thread tags: reuses the same Tag vocabulary as Contact tags

CREATE TABLE thread_tags (
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (thread_id, tag_id)
);

CREATE INDEX idx_thread_tags_tag_id ON thread_tags(tag_id);

-- CheckIn: a single scheduled occurrence; always belongs to a Thread

CREATE TABLE check_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    deadline DATE, -- NULL only possible if thread.cadence_interval_days is NULL
    status checkin_status NOT NULL DEFAULT 'pending',
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Speeds up the core "what's due/overdue" dashboard query
CREATE INDEX idx_checkins_thread_status ON check_ins(thread_id, status);
CREATE INDEX idx_checkins_deadline ON check_ins(deadline);
