# `gcr.io/paketo-buildpacks/executable-jar`
The Paketo Executable JAR Buildpack is a Cloud Native Buildpack that contributes a Process Type for executable JARs.

## Behavior
This buildpack will participate if all the following conditions are met:

* `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` contains a `Main-Class` entry

When building a JVM application the buildpack will do the following:
* Requests that a JRE be installed
* Contributes `<APPLICATION_ROOT>` to build and runtime `$CLASSPATH`
* If `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` `Class-Path` exists
  * Contributes entries to build and runtime `$CLASSPATH`
* Contributes `executable-jar`, `task`, and `web` process types

When participating in the build of a native image application the buildpack will:
* Contributes `<APPLICATION_ROOT>` to build-time `$CLASSPATH`
* If `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` `Class-Path` exists
  * Contributes entries to build-time `$CLASSPATH`

When `$BP_LIVE_RELOAD_ENABLE` is true:
* Requests that `watchexec` be installed
* Contributes `reload` process type

## Configuration
| Environment Variable      | Description                                       |
| ------------------------- | ------------------------------------------------- |
| `$BP_LIVE_RELOAD_ENABLED` | Enable live process reloading. Defaults to false. |

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0

