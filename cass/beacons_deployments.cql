This way, we can enforce (via static column @ partition) 1 deployment per beacon.
Updates will automatically invalidate older deployments for the beacon.
The materialized view will give us eventual consistency for efficiently fetching beacons related
to a deployment. 

remember to use `cqlsh --cqlversion="3.4.4"`

----start

CREATE KEYSPACE IF NOT EXISTS bkn
  WITH REPLICATION = { 
   'class' : 'SimpleStrategy', 
   'replication_factor' : 2 
  };

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

INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a1', null);
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a2', 'dep1');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a3', 'dep1');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a4', 'dep1');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a5', 'dep1');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a6', 'dep1');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a7', 'dep2');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a8', 'dep2');
INSERT INTO beacons (user_id, beacon_name, deploy_name) VALUES (0, 'a9', 'dep2');