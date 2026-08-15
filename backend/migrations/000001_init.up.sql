CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    avatar_url    TEXT,
    bio           TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'mod', 'admin')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE boards (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE threads (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id   UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id),
    title      TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    image_url  TEXT,
    is_pinned  BOOLEAN NOT NULL DEFAULT false,
    is_locked  BOOLEAN NOT NULL DEFAULT false,
    bumped_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_threads_board_bumped ON threads (board_id, is_pinned DESC, bumped_at DESC);
CREATE INDEX idx_threads_bumped ON threads (bumped_at DESC);

CREATE TABLE replies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id   UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    body        TEXT NOT NULL,
    image_url   TEXT,
    reply_to_id UUID REFERENCES replies(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_replies_thread ON replies (thread_id, created_at ASC);

CREATE TABLE reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT NOT NULL CHECK (target_type IN ('thread', 'reply')),
    target_id   UUID NOT NULL,
    reporter_id UUID NOT NULL REFERENCES users(id),
    reason      TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'dismissed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reports_status ON reports (status, created_at ASC);

CREATE TABLE bans (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id),
    ip         INET,
    reason     TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bans_user ON bans (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_bans_ip ON bans (ip) WHERE ip IS NOT NULL;

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    family_id   UUID NOT NULL,
    token_hash  TEXT NOT NULL UNIQUE,
    revoked     BOOLEAN NOT NULL DEFAULT false,
    replaced_by UUID REFERENCES refresh_tokens(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family_id);
