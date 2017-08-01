CREATE TABLE IF NOT EXISTS bkn.deployments_metadata (
  user_id uuid,
  deploy_name varchar,
  message_name varchar,
  created_at timestamp,
  updated_at timestamp,
  PRIMARY KEY ((user_id), deploy_name)
);


