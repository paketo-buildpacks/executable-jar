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
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/executable-jar/v6/executable"
)

func testClassPath(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx         libcnb.BuildContext
		contributor executable.ClassPath
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "class-path-layers")
		Expect(err).NotTo(HaveOccurred())
		contributor = executable.ClassPath{
			ClassPath: []string{"test-value-1", "test-value-2"},
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	context("launch is true", func() {
		it.Before(func() {
			contributor.Launch = true
		})

		it("contributes for build and launch", func() {
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = contributor.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Build).To(BeTrue())
			Expect(layer.SharedEnvironment["CLASSPATH.delim"]).To(Equal(string(os.PathListSeparator)))
			Expect(layer.SharedEnvironment["CLASSPATH.prepend"]).To(Equal("test-value-1:test-value-2"))
		})
	})

	context("launch is false", func() {
		it("contributes for build only", func() {
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = contributor.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.Launch).To(BeFalse())
			Expect(layer.Build).To(BeTrue())
			Expect(layer.BuildEnvironment["CLASSPATH.delim"]).To(Equal(string(os.PathListSeparator)))
			Expect(layer.BuildEnvironment["CLASSPATH.prepend"]).To(Equal("test-value-1:test-value-2"))
		})
	})
}
