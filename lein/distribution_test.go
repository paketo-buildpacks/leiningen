package lein_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	"github.com/eddumelendez/lein-paketo-buildpack/lein"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"
)

func testDistribution(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "distribution-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes distribution", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-lein",
			SHA256: "a1ad367c13990dd91c9a34ee04b878c4256509c32038cd98440feac2dd6d5d21",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		d := lein.NewDistribution(dep, dc, &libcnb.BuildpackPlan{})
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = d.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Cache).To(BeTrue())
		Expect(filepath.Join(layer.Path, "lein")).To(BeARegularFile())
	})

}
