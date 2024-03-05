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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sbom"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger      bard.Logger
	SBOMScanner sbom.SBOMScanner
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	result := libcnb.NewBuildResult()

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, nil)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	jarGlob, _ := cr.Resolve("BP_EXECUTABLE_JAR_LOCATION")
	execJar, err := LoadExecutableJAR(context.Application.Path, jarGlob)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to load executable JAR\n%w", err)
	}

	if !execJar.Executable {
		for _, entry := range context.Plan.Entries {
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
		return result, nil
	}

	b.Logger.Title(context.Buildpack)

	cr, err = libpak.NewConfigurationResolver(context.Buildpack, nil)
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
		command := "java"
		arguments := []string{}
		appCds := cr.ResolveBool("BP_JVM_CDS_ENABLED")
		if appCds {
			arguments = append(arguments, "-jar", "run-app.jar")
		} else if execJar.ExplodedJAR {
			arguments = append(arguments, execJar.MainClass)
		} else {
			arguments = append(arguments, "-jar", execJar.Path)
		}

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
			if !execJar.ExplodedJAR {
				b.Logger.Body("WARNING: Live Reload has been enabled, however, your command is set to run from a JAR file. This may fail or be a worse experience as the entire JAR file must change to trigger a reload.")
			}

			for i := 0; i < len(result.Processes); i++ {
				result.Processes[i].Default = false
			}

			result.Processes = append(result.Processes,
				libcnb.Process{
					Type:      "reload",
					Command:   "watchexec",
					Arguments: append([]string{"-r", "--shell=none", "--", command}, arguments...),
					Direct:    true,
					Default:   true,
				},
			)

			err = filepath.Walk(context.Application.Path, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if path == context.Application.Path {
					return nil
				}

				return os.Chmod(path, info.Mode()|0060)
			})
			if err != nil {
				return libcnb.BuildResult{}, fmt.Errorf("unable to mark files as group read-write for live reload\n%w", err)
			}
		}

		if b.SBOMScanner == nil {
			b.SBOMScanner = sbom.NewSyftCLISBOMScanner(context.Layers, effect.NewExecutor(), b.Logger)
		}
		if err := b.SBOMScanner.ScanLaunch(context.Application.Path, libcnb.SyftJSON, libcnb.CycloneDXJSON); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create Build SBoM \n%w", err)
		}
	}

	if execJar.ExplodedJAR {
		cp := []string{context.Application.Path}
		if s, ok := execJar.Properties.Get("Class-Path"); ok {
			cp = append(cp, strings.Split(s, " ")...)
		}

		classpathLayer := NewClassPath(cp, launch)
		classpathLayer.Logger = b.Logger
		result.Layers = append(result.Layers, classpathLayer)
	}

	return result, nil
}
