/*
  This way, we can enforce (via static column @ partition) 1 deployment per beacon.
  Updates will automatically invalidate older deployments for the beacon.
  The materialized view will give us eventual consistency for efficiently fetching beacons related
  to a deployment. 
  remember to use `cqlsh --cqlversion="3.4.4"`
*/



CREATE TABLE IF NOT EXISTS bkn.beacons (
  user_id uuid,
  deploy_name varchar,
  beacon_name varchar,
  PRIMARY KEY ((user_id, beacon_name))
);

CREATE MATERIALIZED VIEW IF NOT EXISTS bkn.deployments
AS SELECT user_id, deploy_name, beacon_name
FROM bkn.beacons
WHERE user_id IS NOT NULL AND deploy_name IS NOT NULL AND beacon_name IS NOT NULL
PRIMARY KEY ((user_id, deploy_name), beacon_name);


// populate some data

INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a2', null);
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a3', null);
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a4', 'dep1');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a5', 'dep1');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a6', 'dep1');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a7', 'dep1');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a8', 'dep2');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a9', 'dep2');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a10', 'dep2');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a11', 'dep2');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a12', 'dep3');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a13', 'dep3');
INSERT INTO bkn.beacons (user_id, beacon_name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, 'a14', 'dep3');