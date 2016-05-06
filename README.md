# cf-plex

Runs `cf` commands against multiple Cloud Foundry instances.

[![Build Status](https://travis-ci.org/EngineerBetter/cf-plex.svg?branch=master)](https://travis-ci.org/EngineerBetter/cf-plex)

## Using

```
cf-plex add-api https://api.some.com username password
cf-plex add-api https://api.another.com username password
# Then use regular CF commands:
cf-plex create-org new-org
cf-plex list-apis
cf-plex remove-api https://api.another.com
```

`cf-plex` will fail fast if the `cf` CLI returns a non-zero exit code against any API. To override this behaviour (ignore the error and continue running the command) specify `--force`:

```
# Will continue even if it fails against one API
cf-plex delete org might-not-exist --force
```

## Testing

```
CF_USERNAME=testing@engineerbetter.com \
CF_PASSWORD=lookitup \
go test ./...
```