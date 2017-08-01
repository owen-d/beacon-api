CREATE TABLE IF NOT EXISTS bkn.users (
  id uuid,
  provider_id varchar,
  email varchar,
  created_at timestamp,
  updated_at timestamp,
  PRIMARY KEY(id)
);

CREATE MATERIALIZED VIEW IF NOT EXISTS bkn.users_by_email
AS SELECT *
FROM bkn.users
WHERE id IS NOT NULL AND email IS NOT NULL
PRIMARY KEY ((email), id);

CREATE MATERIALIZED VIEW IF NOT EXISTS bkn.users_by_oauth
AS SELECT *
FROM bkn.users
WHERE id IS NOT NULL AND provider_id IS NOT NULL
PRIMARY KEY ((provider_id), id);
