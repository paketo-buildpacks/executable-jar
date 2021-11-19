module github.com/paketo-buildpacks/executable-jar

go 1.15

require (
	github.com/buildpacks/libcnb v1.25.0
	github.com/onsi/gomega v1.17.0
	github.com/paketo-buildpacks/libjvm v1.31.0
	github.com/paketo-buildpacks/libpak v1.54.0
	github.com/sclevine/spec v1.4.0
)

replace github.com/paketo-buildpacks/libpak v1.54.0 => /Users/davidos/workspace/libraries/libpak
