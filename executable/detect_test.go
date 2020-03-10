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
	"github.com/paketo-buildpacks/executable-jar/executable"
	"github.com/sclevine/spec"
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

	it("passes without META-INF/MANIFEST.MF", func() {
		Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Requires: []libcnb.BuildPlanRequire{
						{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
						{Name: "jvm-application"},
					},
				},
			},
		}))
	})

	it("passes with empty META-INF/MANIFEST.MF", func() {
		Expect(os.MkdirAll(filepath.Join(path, "META-INF"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(path, "META-INF", "MANIFEST.MF"), []byte(""), 0644))

		Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Requires: []libcnb.BuildPlanRequire{
						{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
						{Name: "jvm-application"},
					},
				},
			},
		}))
	})

	it("passes with Main-Class", func() {
		Expect(os.MkdirAll(filepath.Join(path, "META-INF"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(path, "META-INF", "MANIFEST.MF"), []byte("Main-Class: test-main-class"), 0644))

		Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{
						{Name: "jvm-application"},
					},
					Requires: []libcnb.BuildPlanRequire{
						{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
						{Name: "jvm-application"},
					},
				},
			},
		}))
	})
}
