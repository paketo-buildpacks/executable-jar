/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sbom"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger      bard.Logger
	SBOMScanner sbom.SBOMScanner
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	result := libcnb.NewBuildResult()

	// Check if there is a top-level META-INF/MANIFEST.MF
	manifest := filepath.Join(context.Application.Path, "META-INF", "MANIFEST.MF")
	manifestExists := true
	if _, err := os.Stat(manifest); errors.Is(err, os.ErrNotExist) {
		manifestExists = false
	}

	mainJar := ""
	m := properties.NewProperties()

	if manifestExists {
		var err error
		m, err = libjvm.NewManifest(context.Application.Path)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to read manifest in %s\n%w", context.Application.Path, err)
		}
	} else {
		// walk through directories, find the JAR file with a Main-Class
		err := filepath.Walk(context.Application.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// opt-out if we already found a JAR file
			if mainJar != "" {
				return nil
			}

			// make sure it is a file
			if info.IsDir() {
				return nil
			}

			// make sure it is a JAR file
			if !strings.HasSuffix(path, ".jar") {
				return nil
			}

			// get the MANIFEST of the JAR file
			manifest, err := libjvm.NewManifestFromJAR(path)
			if err != nil {
				return err
			}

			// we take it if it has a Main-Class
			if _, ok := manifest.Get("Main-Class"); ok {
				mainJar = path
			}

			return nil
		})

		if err != nil {
			return libcnb.BuildResult{}, err
		}
	}

	mainClass := ""
	if mainJar == "" {
		var ok bool
		mainClass, ok = m.Get("Main-Class")
		if !ok {
			for _, entry := range context.Plan.Entries {
				result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
			}
			return result, nil
		}
	}

	b.Logger.Title(context.Buildpack)

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, nil)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	launch := true
	if n, ok, err := pr.Resolve(PlanEntryJVMApplication); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jvm-application plan entry\n%w", err)
	} else if ok {
		if nativeImage, ok := n.Metadata["native-image"].(bool); ok && nativeImage {
			// this executable JAR is an input to a native image build and will not be included in the final image
			launch = false
		}
	}

	if launch {
		arguments := []string{}

		if mainClass != "" {
			arguments = append(arguments, mainClass)
		} else {
			arguments = append(arguments, "-jar", mainJar)
		}

		command := "java"
		result.Processes = append(result.Processes,
			libcnb.Process{
				Type:      "executable-jar",
				Command:   command,
				Arguments: arguments,
				Direct:    true,
			},
			libcnb.Process{
				Type:      "task",
				Command:   command,
				Arguments: arguments,
				Direct:    true,
			},
			libcnb.Process{
				Type:      "web",
				Command:   command,
				Arguments: arguments,
				Direct:    true,
				Default:   true,
			},
		)

		if cr.ResolveBool("BP_LIVE_RELOAD_ENABLED") {
			for i := 0; i < len(result.Processes); i++ {
				result.Processes[i].Default = false
			}

			result.Processes = append(result.Processes,
				libcnb.Process{
					Type:      "reload",
					Command:   "watchexec",
					Arguments: append([]string{"-r", command}, arguments...),
					Direct:    true,
					Default:   true,
				},
			)
		}

		if b.SBOMScanner == nil {
			b.SBOMScanner = sbom.NewSyftCLISBOMScanner(context.Layers, effect.NewExecutor(), b.Logger)
		}
		if err := b.SBOMScanner.ScanLaunch(context.Application.Path, libcnb.SyftJSON, libcnb.CycloneDXJSON); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create Build SBoM \n%w", err)
		}
	}

	if mainJar == "" {
		cp := []string{context.Application.Path}
		if s, ok := m.Get("Class-Path"); ok {
			cp = append(cp, strings.Split(s, " ")...)
		}

		classpathLayer := NewClassPath(cp, launch)
		classpathLayer.Logger = b.Logger
		result.Layers = append(result.Layers, classpathLayer)
	}

	return result, nil
}
