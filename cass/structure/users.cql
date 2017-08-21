CREATE TABLE IF NOT EXISTS bkn.users (
  id uuid,
  email varchar,
  created_at timestamp,
  updated_at timestamp,
  provider_id tinyint,
  given_name varchar,
  family_name varchar,
  public_picture_url varchar,
  PRIMARY KEY(id)
);

CREATE MATERIALIZED VIEW IF NOT EXISTS bkn.users_by_email
AS SELECT *
FROM bkn.users
WHERE id IS NOT NULL AND email IS NOT NULL
PRIMARY KEY ((email), id);