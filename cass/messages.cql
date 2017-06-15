CREATE TABLE IF NOT EXISTS bkn.messages (
  user_id int,
  name varchar,
  title varchar STATIC,
  url varchar STATIC,
  lang varchar STATIC,
  deployment_name varchar,
  PRIMARY KEY ((user_id, name), deployment_name)
);

// rely on client side denormalization to be able to efficiently get list of messages belonging to a user
CREATE TABLE IF NOT EXISTS bkn.users_messages (
  user_id int,
  name varchar
  PRIMARY KEY ((user_id), name)
)