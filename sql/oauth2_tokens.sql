CREATE TABLE IF NOT EXISTS oauth2_tokens (
	id         BIGSERIAL   NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	code       TEXT        NOT NULL,
	access     TEXT        NOT NULL,
	refresh    TEXT        NOT NULL,
	data       JSONB       NOT NULL,
	CONSTRAINT oauth2_tokens_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_expires_at ON oauth2_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_code ON oauth2_tokens (code);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_access ON oauth2_tokens (access);
CREATE INDEX IF NOT EXISTS idx_oauth2_tokens_refresh ON oauth2_tokens (refresh);
