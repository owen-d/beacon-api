// populate some data

use bkn;
INSERT INTO beacons (user_id, name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c9, textAsBlob('00000000000000000000000000000000'), 'dep1');
INSERT INTO beacons (user_id, name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c9, textAsBlob('00000000000000000000000000000001'), 'dep1');

/* live bkns */
INSERT INTO BEACONS (user_id, name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, textAsBlob('ecde1818fe845f2e386357335a4c484b'), 'live_dep');
INSERT INTO BEACONS (user_id, name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, textAsBlob('ecde1818fe845f2e3863774977413673'), 'live_dep');
INSERT INTO BEACONS (user_id, name, deploy_name) VALUES (6ba7b810-9dad-11d1-80b4-00c04fd430c8, textAsBlob('ecde1818fe845f2e38637a6e584a4a33'), 'live_dep');