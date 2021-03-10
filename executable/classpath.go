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
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type ClassPath struct {
	ClassPath        []string
	Launch           bool
	LayerContributor libpak.LayerContributor
	Logger           bard.Logger
}

func NewClassPath(classpath []string, launch bool) ClassPath {
	return ClassPath{
		ClassPath: classpath,
		Launch:    launch,
	}
}

func (c ClassPath) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	contributor := libpak.NewLayerContributor(
		"Class Path",
		map[string]interface{}{
			"classpath": c.ClassPath,
			"launch":    c.Launch,
		},
		libcnb.LayerTypes{
			Build:  true,
			Launch: c.Launch,
		},
	)
	contributor.Logger = c.Logger

	return contributor.Contribute(layer, func() (libcnb.Layer, error) {
		var env libcnb.Environment
		if c.Launch {
			env = layer.SharedEnvironment
		} else {
			env = layer.BuildEnvironment
		}
		env.Prepend("CLASSPATH", string(os.PathListSeparator), strings.Join(c.ClassPath, string(filepath.ListSeparator)))

		return layer, nil
	})
}

func (ClassPath) Name() string {
	return "classpath"
}
