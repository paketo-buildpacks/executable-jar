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

		Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	context("manifest with Main-Class and Class-Path", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(
				filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
				[]byte("Main-Class: test-main-class"),
				0644,
			)).To(Succeed())
		})

		it("contributes process types and classpath", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
				[]byte("Main-Class: test-main-class\nClass-Path: test-class-path"),
				0644,
			)).To(Succeed())

			result, err := executable.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers[0].(executable.ClassPath).ClassPath).To(Equal([]string{ctx.Application.Path, "test-class-path"}))
			Expect(result.Layers[0].(executable.ClassPath).Launch).To(BeTrue())
			Expect(result.Processes).To(ContainElements(
				libcnb.Process{
					Type:      "executable-jar",
					Command:   "java",
					Arguments: []string{"test-main-class"},
					Direct:    true,
				},
				libcnb.Process{
					Type:      "task",
					Command:   "java",
					Arguments: []string{"test-main-class"},
					Direct:    true,
				},
				libcnb.Process{
					Type:      "web",
					Command:   "java",
					Arguments: []string{"test-main-class"},
					Direct:    true,
				},
			))
		})

		context("manifest with Main-Class and without Class-Path", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(
					filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
					[]byte("Main-Class: test-main-class"),
					0644,
				)).To(Succeed())
			})

			it("contributes process types and classpath", func() {
				Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(
					filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
					[]byte("Main-Class: test-main-class"),
					0644,
				)).To(Succeed())

				result, err := executable.Build{}.Build(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Layers[0].(executable.ClassPath).ClassPath).To(Equal([]string{ctx.Application.Path}))
				Expect(result.Layers[0].(executable.ClassPath).Launch).To(BeTrue())
				Expect(result.Processes).To(ContainElements(
					libcnb.Process{
						Type:      "executable-jar",
						Command:   "java",
						Arguments: []string{"test-main-class"},
						Direct:    true,
					},
					libcnb.Process{
						Type:      "task",
						Command:   "java",
						Arguments: []string{"test-main-class"},
						Direct:    true,
					},
					libcnb.Process{
						Type:      "web",
						Command:   "java",
						Arguments: []string{"test-main-class"},
						Direct:    true,
					},
				))
			})
		})

		it("contributes Executable JAR without Class-Path", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
				[]byte(`Main-Class: test-main-class`),
				0644,
			)).To(Succeed())

			result, err := executable.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers[0].(executable.ClassPath).ClassPath).To(Equal([]string{ctx.Application.Path}))
			Expect(result.Processes).To(ContainElements(
				libcnb.Process{
					Type:      "executable-jar",
					Command:   "java",
					Arguments: []string{"test-main-class"},
					Direct:    true,
				},
				libcnb.Process{
					Type:      "task",
					Command:   "java",
					Arguments: []string{"test-main-class"},
					Direct:    true,
				},
				libcnb.Process{
					Type:      "web",
					Command:   "java",
					Arguments: []string{"test-main-class"},
					Direct:    true,
				},
			))
		})

		context("native image", func() {
			it("contributes classpath for build", func() {
				Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(
					filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
					[]byte("Main-Class: test-main-class\nClass-Path: test-class-path"),
					0644,
				)).To(Succeed())

				ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{
					Name:     "jvm-application",
					Metadata: map[string]interface{}{"native-image": true},
				})

				result, err := executable.Build{}.Build(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Processes).To(BeEmpty())
				Expect(result.Layers).To(HaveLen(1))
				Expect(result.Layers[0].(executable.ClassPath).ClassPath).To(Equal([]string{ctx.Application.Path, "test-class-path"}))
				Expect(result.Layers[0].(executable.ClassPath).Launch).To(BeFalse())
			})
		})
	})

	context("Manifest does not have Main-Class", func() {
		it("return unmet jvm-application plan entry", func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(ctx.Application.Path, "META-INF", "MANIFEST.MF"),
				[]byte(`Class-Path: test-class-path`),
				0644,
			)).To(Succeed())

			ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{
				Name: "jvm-application",
			})

			result, err := executable.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(BeEmpty())
			Expect(result.Processes).To(BeEmpty())
			Expect(result.Unmet).To(HaveLen(1))
			Expect(result.Unmet[0].Name).To(Equal("jvm-application"))
		})
	})
}
