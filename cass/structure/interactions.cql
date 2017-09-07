/*
  Interaction data: when a passerby clicks into a beacon message.
  3 months in seconds via `60 * 60 * 24 * 30.5 * 3`
  yields 7905600
*/

CREATE TABLE IF NOT EXISTS bkn.interactions (
  moment timestamp,
  bkn_name blob,
  bkn_user_id uuid,
  deploy_name varchar,
  PRIMARY KEY ((bkn_user_id, deploy_name), moment)
) WITH default_time_to_live = 7905600;
