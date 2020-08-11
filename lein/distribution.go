package lein

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/pexec"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type Distribution struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewDistribution(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, plan *libcnb.BuildpackPlan) Distribution {
	return Distribution{
		LayerContributor: libpak.NewDependencyLayerContributor(dependency, cache, plan)}
}

func (d Distribution) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	d.LayerContributor.Logger = d.Logger

	return d.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		d.Logger.Bodyf("Copying lein to %s", layer.Path)
		file := filepath.Join(layer.Path, "lein")
		err := sherpa.CopyFile(artifact, file)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy lein\n%w", err)
		}
		if err := os.Chmod(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}
		err = pexec.NewExecutable(file).Execute(pexec.Execution{})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to execute lein\n%w", err)
		}

		layer.Cache = true
		return layer, nil
	})
}

func (d Distribution) Name() string {
	return d.LayerContributor.LayerName()
}
