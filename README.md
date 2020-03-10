# `executable-jar`
The Paketo Executable JAR Buildpack is a Cloud Native Buildpack that contributes a Process Type for executable JARs.

This buildpack is designed to work in collaboration with other buildpacks.

## Detection
The detection phase passes if

* `jvm-application` exists in the build plan
  * Contributes `jre` to the build plan
  * Contributes `jre.metadata.launch = true` to the build plan

## Build
If the build plan contains

* `jvm-application`
  * `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` with `Main-Class` key declared
    * Contributes `executable-jar` process type
    * Contributes `task` process type
    * Contributes `web` process type
   * `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` with `Class-Path` key declared
    * Contributes entries to the `$CLASSPATH` environment variable

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
