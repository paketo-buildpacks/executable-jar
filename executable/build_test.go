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

package executable_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/executable-jar/executable"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "build-application")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "build-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes Executable JAR with Class-Path", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"), []byte(`Main-Class: test-main-class
Class-Path: test-class-path`), 0644))

		result, err := executable.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].(executable.ClassPath).ClassPath).To(Equal([]string{ctx.Application.Path, "test-class-path"}))
		Expect(result.Processes).To(ContainElements(
			libcnb.Process{Type: "executable-jar", Command: `java -cp "${CLASSPATH}" ${JAVA_OPTS} test-main-class`},
			libcnb.Process{Type: "task", Command: `java -cp "${CLASSPATH}" ${JAVA_OPTS} test-main-class`},
			libcnb.Process{Type: "web", Command: `java -cp "${CLASSPATH}" ${JAVA_OPTS} test-main-class`},
		))
	})

	it("contributes Executable JAR without Class-Path", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"), []byte(`Main-Class: test-main-class`), 0644))

		result, err := executable.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers[0].(executable.ClassPath).ClassPath).To(Equal([]string{ctx.Application.Path}))
		Expect(result.Processes).To(ContainElements(
			libcnb.Process{Type: "executable-jar", Command: `java -cp "${CLASSPATH}" ${JAVA_OPTS} test-main-class`},
			libcnb.Process{Type: "task", Command: `java -cp "${CLASSPATH}" ${JAVA_OPTS} test-main-class`},
			libcnb.Process{Type: "web", Command: `java -cp "${CLASSPATH}" ${JAVA_OPTS} test-main-class`},
		))
	})

}
