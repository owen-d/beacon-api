CREATE TABLE IF NOT EXISTS bkn.messages (
  user_id uuid,
  name varchar,
  title varchar,
  url varchar,
  lang varchar,
  deployments set<varchar>,
  PRIMARY KEY ((user_id), name)
);
