# cf-plex

Runs `cf` commands against multiple Cloud Foundry instances.

[![Build Status](https://travis-ci.org/EngineerBetter/cf-plex.svg?branch=master)](https://travis-ci.org/EngineerBetter/cf-plex)

## Example

Create `new-org` on two Cloud Foundry instances:

```bash
cf-plex add-api https://api.some.com username password
cf-plex add-api https://api.another.com username password
cf-plex create-org new-org
```

## Usage

### Interactive Mode

`cf-plex` manages a set of `CF_HOME` directories, one for each Cloud Foundry instance you ask it to manage. These are stored in `CF_PLEX_HOME`.

* `cf-plex add-api https://api.some.com username password` Add an API to be used
* `cf-plex add-api https://api.some.com` Add an API to be used, and prompt for credentials
* `cf-plex list-apis` Show APIs that are active
* `cf-plex remove-api https://api.some.com` Remove an API

### Batch Mode

Specify API details in `CF_PLEX_APIS` to avoid manual credential management:

```bash
export CF_PLEX_APIS="username^password>https://api.some.com;username^password>https://api.another.com"
cf-plex create-org new-org
```

State for batch operations is stored separately to interactive mode: that is, each API's `CF_HOME` is stored as a subdirectory `$HOME/$CF_PLEX_HOME/batch`. 

If your credentials contain the separators used in the example above, you can specify your own as environment variables:

* `CF_PLEX_SEP_TRIPLE` for the separator between the three items that identify a Cloud Foundry
* `CF_PLEX_SEP_CREDS_API` for the separator between the user/pass and the API URL
* `CF_PLEX_SEP_USER_PASS` for the separator betwen the username and the password

### Ignoring Errors

`cf-plex` will fail fast if the `cf` CLI returns a non-zero exit code against any API. To override this behaviour (ignore the error and continue running the command) specify `--force`:

```bash
# Will continue even if it fails against one API
cf-plex delete org might-not-exist --force
```

## Testing

Currently depends on having an account on Pivotal Web Services and BlueMix.

```bash
CF_USERNAME=testing@engineerbetter.com \
CF_PASSWORD=lookitup \
go test -v ./...
```