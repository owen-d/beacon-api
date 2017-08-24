Running beacon-api requires a few things, both in a dir, whose path is passes as the `CONFIGS_DIR` env var:
- google service acct credentials: `gcp-credentials.json`
- json config file: `config.json`


For registering beacons, hash a provider key, then use that as a prefix for provider key + provider id (assuming id is coercible to hex). i.e.
`prefix = echo -n 'ibks105' | shasum`
`first_x_chars(prefix) + concat(provider_id)` so that length = 16
