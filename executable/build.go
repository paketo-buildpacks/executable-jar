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
	m, err := libjvm.NewManifest(context.Application.Path)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to read manifest in %s\n%w", context.Application.Path, err)
	}

	mc, ok := m.Get("Main-Class")
	if !ok {
		return libcnb.BuildResult{}, nil
	}

	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	ni := false
	if n, ok, err := pr.Resolve("jvm-application"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jvm-application plan entry\n%w", err)
	} else if ok {
		if v, ok := n.Metadata["native-image"].(bool); ok {
			ni = v
		}
	}

	if !ni {
		command := fmt.Sprintf(`java -cp "${CLASSPATH}" ${JAVA_OPTS} %s`, mc)
		result.Processes = append(result.Processes,
			libcnb.Process{Type: "executable-jar", Command: command},
			libcnb.Process{Type: "task", Command: command},
			libcnb.Process{Type: "web", Command: command},
		)

		cp := []string{context.Application.Path}
		if s, ok := m.Get("Class-Path"); ok {
			cp = append(cp, strings.Split(s, " ")...)
		}

		result.Layers = append(result.Layers, NewClassPath(cp))
	}

	return result, nil
}
