/*
 * Copyright 2018-2020 the original author or authors.
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
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	result := libcnb.NewBuildResult()

	m, err := libjvm.NewManifest(context.Application.Path)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to read manifest in %s\n%w", context.Application.Path, err)
	}

	mc, ok := m.Get("Main-Class")
	if !ok {
		for _, entry := range context.Plan.Entries {
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
		return result, nil
	}

	b.Logger.Title(context.Buildpack)

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
		arguments := []string{mc}
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
	}

	cp := []string{context.Application.Path}
	if s, ok := m.Get("Class-Path"); ok {
		cp = append(cp, strings.Split(s, " ")...)
	}

	classpathLayer := NewClassPath(cp, launch)
	classpathLayer.Logger = b.Logger
	result.Layers = append(result.Layers, classpathLayer)

	return result, nil
}
