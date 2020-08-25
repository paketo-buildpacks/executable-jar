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
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
)

type ClassPath struct {
	ClassPath        []string
	LayerContributor libpak.LayerContributor
}

func NewClassPath(classpath []string) ClassPath {
	expected := map[string][]string{"classpath": classpath}

	return ClassPath{
		ClassPath:        classpath,
		LayerContributor: libpak.NewLayerContributor("JVM Classpath", expected),
	}
}

func (c ClassPath) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return c.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		layer.LaunchEnvironment.PrependPath("CLASSPATH", strings.Join(c.ClassPath, string(filepath.ListSeparator)))

		return layer, nil
	}, libpak.LaunchLayer)
}

func (ClassPath) Name() string {
	return "class-path"
}
