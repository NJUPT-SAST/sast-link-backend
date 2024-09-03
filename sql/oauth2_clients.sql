CREATE TABLE IF NOT EXISTS oauth2_clients (
  id     TEXT    NOT NULL,
  secret TEXT    NOT NULL,
  domain TEXT[]  NOT NULL,
  data   JSONB   NOT NULL,
  CONSTRAINT oauth2_clients_pkey PRIMARY KEY (id)
);
