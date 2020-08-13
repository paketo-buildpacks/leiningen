# Paketo Buildpack for Leiningen

The Paketo Leiningen Buildpack is a Cloud Native Buildpack that builds Leiningen-based applications from source.

## Behavior
This buildpack will participate all the following conditions are met

* `<APPLICATION_ROOT>/project.clj` exists

The buildpack will do the following:

* Requests that a JDK be installed
* Links the `~/.lein` to a layer for caching
* If `<APPLICATION_ROOT>/lein` exists
  * Runs `<APPLICATION_ROOT>/lein uberjar` to build the application
* If `<APPLICATION_ROOT>/lein` does not exist
  * Contributes Lein to a layer with all commands on `$PATH`
  * Runs `<LEIN_ROOT>/bin/lein uberjar` to build the application
* Removes the source code in `<APPLICATION_ROOT>`
* Expands `<APPLICATION_ROOT>/target/*-standalone.jar` to `<APPLICATION_ROOT>`

## Configuration
| Environment Variable | Description
| -------------------- | -----------
| `$BP_LEIN_BUILD_ARGUMENTS` | Configure the arguments to pass to build system.  Defaults to `uberjar`.
| `$BP_LEIN_BUILT_MODULE` | Configure the module to find application artifact in.  Defaults to the root module (empty).
| `$BP_LEIN_BUILT_ARTIFACT` | Configure the built application artifact explicitly.  Supersedes `$BP_LEIN_BUILT_MODULE`  Defaults to `target/*-standalone.jar`.
