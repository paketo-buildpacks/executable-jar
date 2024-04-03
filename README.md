# `gcr.io/paketo-buildpacks/executable-jar`

The Paketo Buildpack for Executable JAR is a Cloud Native Buildpack that contributes a Process Type for executable JARs.

## Behavior

This buildpack will participate if any the following conditions are met:

* `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` contains a `Main-Class` entry
* `<APPLICATION_ROOT>/**/*.jar` exists and that JAR has a `/META-INF/MANIFEST.MF` file which contains a `Main-Class` entry

When building a JVM application the buildpack will do the following:

* Requests that a JRE be installed
* If `<APPLICATION_ROOT>` contains an exploded JAR:
  * It contributes `<APPLICATION_ROOT>` to build and runtime `$CLASSPATH`
  * If `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` `Class-Path` exists
    * Contributes entries to build and runtime `$CLASSPATH`
* Contributes `executable-jar`, `task`, and `web` process types

When participating in the build of a native image application the buildpack will:

* If `<APPLICATION_ROOT>` contains an exploded JAR:
  * Contributes `<APPLICATION_ROOT>` to build-time `$CLASSPATH`
  * If `<APPLICATION_ROOT>/META-INF/MANIFEST.MF` `Class-Path` exists
    * Contributes entries to build-time `$CLASSPATH`

When `$BP_LIVE_RELOAD_ENABLE` is true:

* Requests that `watchexec` be installed
* Contributes `reload` process type

## Configuration

| Environment Variable          | Description                                                                                                                                                               |
|-------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `$BP_LIVE_RELOAD_ENABLED`     | Enable live process reloading. Defaults to false.                                                                                                                         |
| `$BP_EXECUTABLE_JAR_LOCATION` | An optional glob to specify the JAR used as an entrypoint. Defaults to "", which causes the buildpack to do a breadth-first search for the first executable JAR it finds. |
## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0

