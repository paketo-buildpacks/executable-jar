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
	"archive/zip"
	"errors"
	"fmt"
	"github.com/paketo-buildpacks/libpak/sherpa"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magiconair/properties"
	"github.com/paketo-buildpacks/libjvm"

	"github.com/paketo-buildpacks/executable-jar/v6/internal/fsutil"
)

type ExecutableJAR struct {
	MainClass   string
	Path        string
	Properties  *properties.Properties
	Executable  bool
	ExplodedJAR bool
}

func LoadExecutableJAR(appPath string, executableJarGlob string) (ExecutableJAR, error) {
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
		jarPath, props, err = findExecutableJAR(appPath, executableJarGlob)
		if err != nil {
			return ExecutableJAR{}, fmt.Errorf("unable to parse manifest\n%w", err)
		}
		// let's explode the executable jar we found, and remove all other jars
		tempExplodedJar := os.TempDir() + "/" + fmt.Sprint(time.Now().UnixMilli()) + "/"
		Unzip(jarPath, tempExplodedJar)
		os.RemoveAll(appPath)
		sherpa.CopyDir(tempExplodedJar, appPath)
		jarPath = appPath
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

func findExecutableJAR(appPath, executableJarGlob string) (string, *properties.Properties, error) {
	props := &properties.Properties{}
	jarPath := ""
	stopWalk := errors.New("stop walking")

	w := func(root string, fn filepath.WalkFunc) error {
		if executableJarGlob != "" {
			files, _ := filepath.Glob(filepath.Join(appPath, executableJarGlob))
			for _, f := range files {
				fi, err := os.Lstat(f)
				err = fn(f, fi, err)
				if err != nil {
					return err
				}
			}
		}
		return fsutil.Walk(root, fn)
	}

	err := w(appPath, func(path string, info os.FileInfo, err error) error {
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

func Unzip(src, dest string) error {
	dest = filepath.Clean(dest) + "/"

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer CloseOrPanic(r)()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		path := filepath.Join(dest, f.Name)
		// Check for ZipSlip: https://snyk.io/research/zip-slip-vulnerability
		if !strings.HasPrefix(path, dest) {
			return fmt.Errorf("%s: illegal file path", path)
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer CloseOrPanic(rc)()

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer CloseOrPanic(f)()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func CloseOrPanic(f io.Closer) func() {
	return func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}
