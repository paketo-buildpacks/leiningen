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

package lein

import (
	"fmt"
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
		d.Logger.Bodyf("Copying %s", layer.Path)

		file := filepath.Join(layer.Path, "bin", filepath.Base(artifact.Name()))
		err := sherpa.CopyFile(artifact, file)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy lein\n%w", err)
		}

		if err := os.Chmod(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		layer.Cache = true
		return layer, nil
	})
}

func (d Distribution) Name() string {
	return d.LayerContributor.LayerName()
}
