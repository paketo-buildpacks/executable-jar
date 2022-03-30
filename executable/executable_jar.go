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
	"github.com/paketo-buildpacks/libjvm"
)

type ExecutableJAR struct {
	MainClass   string
	Path        string
	Properties  *properties.Properties
	Executable  bool
	ExplodedJAR bool
}

func LoadExecutableJAR(appPath string) (ExecutableJAR, error) {
	_, err := os.Stat(filepath.Join(appPath, "META-INF", "MANIFEST.MF"))
	if err != nil && !os.IsNotExist(err) {
		return ExecutableJAR{}, fmt.Errorf("unable to read manifest.mf\n%w", err)
	}

	explodedJAR := !(err != nil && os.IsNotExist(err))
	var props *properties.Properties
	var jarPath = appPath

	if explodedJAR {
		props, err = libjvm.NewManifest(appPath)
		if err != nil {
			return ExecutableJAR{}, fmt.Errorf("unable to parse manifest\n%w", err)
		}
	} else {
		jarPath, props, err = findExecutableJAR(appPath)
		if err != nil {
			return ExecutableJAR{}, fmt.Errorf("unable to parse manifest\n%w", err)
		}
	}

	if mc, ok := props.Get("Main-Class"); ok {
		return ExecutableJAR{
			MainClass:   mc,
			Properties:  props,
			Path:        jarPath,
			Executable:  ok,
			ExplodedJAR: appPath == jarPath,
		}, nil
	}

	return ExecutableJAR{}, nil
}

func findExecutableJAR(appPath string) (string, *properties.Properties, error) {
	props := &properties.Properties{}
	jarPath := ""
	stopWalk := errors.New("stop walking")

	err := filepath.Walk(appPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
		props, err = libjvm.NewManifestFromJAR(path)
		if err != nil {
			return fmt.Errorf("unable to load manifest\n%w", err)
		}

		// we take it if it has a Main-Class
		if _, ok := props.Get("Main-Class"); ok {
			jarPath = path
			return stopWalk
		}

		return nil
	})

	if err != nil && !errors.Is(err, stopWalk) {
		return "", nil, err
	}

	return jarPath, props, nil
}
