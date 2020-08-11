package lein_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	"github.com/eddumelendez/lein-paketo-buildpack/lein"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libbs"
	"github.com/sclevine/spec"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx       libcnb.BuildContext
		leinBuild lein.Build
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "build-application")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "build-layers")
		Expect(err).NotTo(HaveOccurred())
		leinBuild = lein.Build{ApplicationFactory: &FakeApplicationFactory{}}
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("does not contribute distribution if wrapper exists", func() {
		Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "lein"), []byte{}, 0644)).To(Succeed())
		ctx.StackID = "test-stack-id"

		result, err := leinBuild.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("cache"))
		Expect(result.Layers[1].Name()).To(Equal("application"))
		Expect(result.Layers[1].(libbs.Application).Command).To(Equal(filepath.Join(ctx.Application.Path, "lein")))
	})

	it("contributes distribution", func() {
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "lein",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := leinBuild.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(3))
		Expect(result.Layers[0].Name()).To(Equal("lein"))
		Expect(result.Layers[1].Name()).To(Equal("cache"))
		Expect(result.Layers[2].Name()).To(Equal("application"))
		Expect(result.Layers[2].(libbs.Application).Command).To(Equal(filepath.Join(ctx.Layers.Path, "lein", "bin", "lein")))
	})
}

type FakeApplicationFactory struct{}

func (f *FakeApplicationFactory) NewApplication(
	_ map[string]interface{},
	_ []string,
	_ libbs.ArtifactResolver,
	_ libbs.Cache,
	command string,
	_ *libcnb.BuildpackPlan,
	_ string,
) (libbs.Application, error) {
	return libbs.Application{
		Command: command,
	}, nil
}
