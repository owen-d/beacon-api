/*
  deployment table; mainly stores metadata about deployments
*/

CREATE TABLE IF NOT EXISTS bkn.deployments_metadata (
  user_id uuid,
  deploy_name varchar,
  message_name varchar,
  PRIMARY KEY ((user_id), deploy_name)
)