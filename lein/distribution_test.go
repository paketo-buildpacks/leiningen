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

package lein_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/leiningen/v2/lein"
)

func testDistribution(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = os.MkdirTemp("", "distribution-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes distribution", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-lein",
			SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		d, _ := lein.NewDistribution(dep, dc)
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = d.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Cache).To(BeTrue())
		Expect(filepath.Join(layer.Path, "bin", "stub-lein")).To(BeARegularFile())

		fi, err := os.Stat(filepath.Join(layer.Path, "bin", "stub-lein"))
		Expect(err).NotTo(HaveOccurred())
		Expect(fi.Mode()).To(BeEquivalentTo(0755))
	})

}
