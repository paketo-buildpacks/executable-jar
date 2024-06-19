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

	"github.com/paketo-buildpacks/executable-jar/v6/executable"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.DetectContext
		detect executable.Detect
		path   string
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "executable-jar")
		Expect(err).NotTo(HaveOccurred())

		ctx.Application.Path = path
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("META-INF/MANIFEST.MF not found", func() {
		it("requires jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("empty META-INF/MANIFEST.MF not found", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(path, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(path, "META-INF", "MANIFEST.MF"),
				[]byte(""),
				0644,
			)).To(Succeed())
		})

		it("requires jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("META-INF/MANIFEST.MF with Main-Class", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(path, "META-INF"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(
				filepath.Join(path, "META-INF", "MANIFEST.MF"),
				[]byte("Main-Class: test-main-class"),
				0644,
			)).To(Succeed())
		})

		it("requires and provides jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
							{Name: "jvm-application-package"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("META-INF/MANIFEST.MF not found but there is JAR with Main-Class", func() {
		it.Before(func() {
			Expect(CreateJAR(filepath.Join(ctx.Application.Path, "a.jar"), map[string]string{"Main-Class": "test.Main"})).To(Succeed())
		})

		it("requires and provides jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
							{Name: "jvm-application-package"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("$BP_LIVE_RELOAD_ENABLED is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
		})

		it("requires watchexec", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
							{Name: "watchexec"},
						},
					},
				},
			}))
		})
	})
}
