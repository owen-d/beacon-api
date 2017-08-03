  /*
  Note: we no longer use static column @ partition, but rather just user userid/bname as primary key. This way, we can enforce 1 deployment per beacon.
  Updates will automatically invalidate older deployments for the beacon.
  The materialized view will give us eventual consistency for efficiently fetching beacons related
  to a deployment. 
  remember to use `cqlsh --cqlversion="3.4.4"`
*/



CREATE TABLE IF NOT EXISTS bkn.beacons (
  user_id uuid,
  deploy_name varchar,
  name blob,
  created_at timestamp,
  updated_at timestamp,
  manu_id blob,
  manu_key smallint,
  tags map<varchar, varchar>,
  PRIMARY KEY ((user_id), name)
);

CREATE MATERIALIZED VIEW IF NOT EXISTS bkn.beacon_deployments
AS SELECT user_id, deploy_name, name
FROM bkn.beacons
WHERE user_id IS NOT NULL AND deploy_name IS NOT NULL AND name IS NOT NULL
PRIMARY KEY ((user_id, deploy_name), name);