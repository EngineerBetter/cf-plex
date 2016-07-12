# cf-plex

![Build Status](http://ci.engineerbetter.com/api/v1/pipelines/cf-plex/jobs/test/badge)

Runs `cf` commands against multiple Cloud Foundry instances.

Cloud Foundry instances can be specified in three ways:

1. Ad-hoc: add APIs to one default list, and run commands against all of them
1. Groups: add APIs to named groups, and run commands against all APIs in a specified group
1. Batch: specify APIs in environment variables, and run commands only against those in the env vars

### Usage

```
  cf-plex [-g <group>] <cf cli command> [--force]
  cf-plex add-api [-g <group>] <apiUrl> [<username> <password>]
  cf-plex list-apis
  cf-plex remove-api [-g <group>] <apiUrl>
```

## Installation

### Go developers

```
go get github.com/EngineerBetter/cf-plex
```

### Everyone else

[Download the latest release](https://github.com/EngineerBetter/cf-plex/releases/latest) for your OS, save it to `PATH`, rename it `cf-plex` and make sure it is executable (`chmod +x cf-plex`).

## Detailed Usage

### Ad Hoc Mode

Add and remove APIs in **one global list**.

* `cf-plex add-api https://api.some.com username password` Add an API to be used
* `cf-plex add-api https://api.some.com` Add an API to be used, and prompt for credentials
* `cf-plex list-apis` Show APIs that are active
* `cf-plex remove-api https://api.some.com` Remove an API

`cf-plex` manages a set of `CF_HOME` directories, one for each Cloud Foundry instance you ask it to manage. These are stored in `CF_PLEX_HOME`.

### Group Mode

Manage APIs in **named groups**. Use cases include operating on all non-production instances at once.

* `cf-plex add-api -g nonprod https://api.nonprod.example.com username password` Add an API to the 'nonprod' group
* `cf-plex list-apis` Show all groups and APIs
* `cf-plex -g nonprod delete-user admin` Delete the admin user only on non-prod
* `cf-plex remove-api -g nonprod https://api.nonprod.example.com` Remove an API from a group

You may not add a group called "default" or "batch", as these are used internally.

Once any group has been added, APIs not assigned to a group can only be referenced by using `-g default`. This is to prevent slips whereby a user may forget to specify a group and accidentally run commands against whatever they defined previously.

Groups are implicitly managed: once there are no APIs in a group, it will cease to exist.

`CF_HOME` directories for APIs in a group are stored in `$CF_PLEX_HOME/groups/`, which is deleted automatically when the last group is removed.

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

`cf-plex` stores the `CF_HOME` directories for APIs used in batch mode in `$CF_PLEX_HOME/groups/batch`. These are left on disk, to prevent unecessary authentication on successive invocations.

### Ignoring Errors

`cf-plex` will fail fast if the `cf` CLI returns a non-zero exit code against any API. To override this behaviour (ignore the error and continue running the command) specify `--force`:

```bash
# Will continue even if it fails against one API
cf-plex delete org might-not-exist --force
```

### Plugins

CF CLI plugins are managed with an orthogonal home directory of `CF_PLUGIN_HOME`. `cf-plex` doesn't do anything with this, so all your usual plugins will be available. If you have a use case that requires plugin isolation, please raise an issue.

## Testing

Currently depends on having an account on Pivotal Web Services and BlueMix.

```bash
CF_USERNAME=testing@engineerbetter.com \
CF_PASSWORD=lookitup \
go test -v ./...
```

## Project

* CI: http://ci.engineerbetter.com/pipelines/cf-plex
* Tracker: https://www.pivotaltracker.com/n/projects/1579861

## Acknowledgements

In order to prove that plugin behaviour is unaffected by `cf-plex`, [Simon Leung's `cli-plugin-echo`](https://github.com/simonleung8/cli-plugin-echo) is vendored as a test fixture. 