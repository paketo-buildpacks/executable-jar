# `paketo-buildpacks/executable-jar`
The Paketo Executable JAR Buildpack is a Cloud Native Buildpack that contributes a Process Type for executable JARs.

## Behavior
This buildpack will participate if all of the following conditions are met

* `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` contains a `Main-Class` entry

The buildpack will do the following:

* Requests that a JRE be installed
* Contributes `<APPLICATION_ROOT>` to `$CLASSPATH`
* If `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` `Class-Path` exists
  * Contributes entries to `$CLASSPATH`
* Contributes `executable-jar`, `task`, and `web` process types

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
