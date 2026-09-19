// Copyright 2018-2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	"github.com/paketo-buildpacks/libdependency/github"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	id       = "lein"
	name     = "Leiningen"
	purlName = "leiningen"

	org  = "technomancy"
	repo = "leiningen"
)

func main() {
	retrieve.NewMetadata(id, github.GetAllVersions(os.Getenv("GITHUB_TOKEN"), org, repo), generateMetadata)
}

func generateMetadata(version versionology.VersionFetcher) ([]versionology.Dependency, error) {
	versionString := version.Version().String()
	tag := version.Version().Original()

	uri := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/bin/lein", org, repo, tag)
	checksum, err := upstream.GetSHA256OfRemoteFile(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", uri, err)
	}

	source := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", org, repo, tag)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:leiningen:leiningen:%s:*:*:*:*:*:*:*", versionString),
		ID:       id,
		Licenses: []interface{}{
			map[string]string{
				"type": "EPL-1.0",
				"uri":  "https://github.com/technomancy/leiningen/blob/stable/COPYING",
			},
		},
		Name:           name,
		PURL:           retrieve.GeneratePURL(purlName, versionString, checksum, source),
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		URI:            uri,
		Version:        versionString,
	}

	return versionology.NewDependencyArray(dependency, "")
}
