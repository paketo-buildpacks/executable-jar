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

package executable_test

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/executable-jar/v6/executable"
	"github.com/sclevine/spec"
)

func testManifest(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect  = NewWithT(t).Expect
		appPath string
	)

	it.Before(func() {
		var err error

		appPath, err = ioutil.TempDir("", "manifest")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(appPath)).To(Succeed())
	})

	context("exploded JAR", func() {
		it("fail if not executable", func() {
			Expect(os.MkdirAll(filepath.Join(appPath, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(appPath, "META-INF", "MANIFEST.MF"),
				[]byte(`foo: bar`),
				0644,
			)).To(Succeed())

			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.Executable).To(BeFalse())
			Expect(ej.ExplodedJAR).To(BeFalse())
			Expect(ej.MainClass).To(BeEmpty())
		})

		it("fail if file not found", func() {
			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.Executable).To(BeFalse())
			Expect(ej.ExplodedJAR).To(BeFalse())
			Expect(ej.MainClass).To(BeEmpty())
		})

		it("loads executable JAR properties", func() {
			Expect(os.MkdirAll(filepath.Join(appPath, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(appPath, "META-INF", "MANIFEST.MF"),
				[]byte(`Main-Class: Foo`),
				0644,
			)).To(Succeed())

			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.MainClass).To(Equal("Foo"))
			Expect(ej.Path).To(Equal(appPath))
			Expect(ej.ExplodedJAR).To(BeTrue())
			Expect(ej.Executable).To(BeTrue())
			Expect(ej.Properties.Map()).To(HaveKeyWithValue("Main-Class", "Foo"))
		})
	})

	context("executable JAR", func() {
		it("fails if it can't find an executable JAR", func() {
			Expect(CreateJAR(filepath.Join(appPath, "test-1.jar"), map[string]string{"foo": "bar"})).To(Succeed())

			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.Executable).To(BeFalse())
			Expect(ej.ExplodedJAR).To(BeFalse())
			Expect(ej.MainClass).To(BeEmpty())
		})

		it("loads props from an executable JAR", func() {
			Expect(CreateJAR(filepath.Join(appPath, "test-1.jar"), map[string]string{"Main-Class": "Foo"})).To(Succeed())

			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.MainClass).To(Equal("Foo"))
			Expect(ej.Path).To(Equal(filepath.Join(appPath, "test-1.jar")))
			Expect(ej.ExplodedJAR).To(BeFalse())
			Expect(ej.Executable).To(BeTrue())
			Expect(ej.Properties.Map()).To(HaveKeyWithValue("Main-Class", "Foo"))
		})

		it("loads props from first executable JAR found", func() {
			Expect(os.MkdirAll(filepath.Join(appPath, "lib"), 0755))
			Expect(CreateJAR(filepath.Join(appPath, "lib", "a.jar"), map[string]string{"Main-Class": "Lib1"})).To(Succeed())
			Expect(CreateJAR(filepath.Join(appPath, "test-1.jar"), map[string]string{"Main-Class": "Foo1"})).To(Succeed())
			Expect(CreateJAR(filepath.Join(appPath, "test-2.jar"), map[string]string{"Main-Class": "Foo2"})).To(Succeed())

			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.MainClass).To(Equal("Foo1"))
			Expect(ej.Path).To(Equal(filepath.Join(appPath, "test-1.jar")))
			Expect(ej.ExplodedJAR).To(BeFalse())
			Expect(ej.Executable).To(BeTrue())
			Expect(ej.Properties.Map()).To(HaveKeyWithValue("Main-Class", "Foo1"))
		})

		it("skips non-executable JARs", func() {
			Expect(CreateJAR(filepath.Join(appPath, "test-1.jar"), map[string]string{"foo": "bar"})).To(Succeed())
			Expect(CreateJAR(filepath.Join(appPath, "test-2.jar"), map[string]string{"Main-Class": "Foo2"})).To(Succeed())

			ej, err := executable.LoadExecutableJAR(appPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(ej.MainClass).To(Equal("Foo2"))
			Expect(ej.Path).To(Equal(filepath.Join(appPath, "test-2.jar")))
			Expect(ej.ExplodedJAR).To(BeFalse())
			Expect(ej.Executable).To(BeTrue())
			Expect(ej.Properties.Map()).To(HaveKeyWithValue("Main-Class", "Foo2"))
		})
	})
}

func CreateJAR(fileName string, props map[string]string) error {
	archive, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("unable to create zip\n%w", err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	if props != nil {
		manifestWriter, err := zipWriter.Create("META-INF/MANIFEST.MF")
		if err != nil {
			return fmt.Errorf("unable to create file in zip\n%w", err)
		}

		for k, v := range props {
			_, err = manifestWriter.Write([]byte(fmt.Sprintf("%s: %s", k, v)))
			if err != nil {
				return fmt.Errorf("unable to write file in zip\n%w", err)
			}
		}
	}

	return zipWriter.Close()
}
