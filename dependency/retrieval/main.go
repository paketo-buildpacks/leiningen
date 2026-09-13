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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

var httpClient = &http.Client{}

func main() {
	var buildpackTomlPath, outputPath string
	flag.StringVar(&buildpackTomlPath, "buildpack-toml-path", "", "Path to buildpack.toml")
	flag.StringVar(&outputPath, "output", "", "Path to output metadata.json")
	flag.Parse()

	if buildpackTomlPath == "" || outputPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s --buildpack-toml-path <path> --output <path>\n", os.Args[0])
		os.Exit(1)
	}

	// Load buildpack.toml
	file, err := os.Open(buildpackTomlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening buildpack.toml: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var config cargo.Config
	if _, err := toml.NewDecoder(file).Decode(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing buildpack.toml: %v\n", err)
		os.Exit(1)
	}

	// Get constraints for lein
	var constraints []cargo.ConfigMetadataDependencyConstraint
	for _, c := range config.Metadata.DependencyConstraints {
		if c.ID == "lein" {
			constraints = append(constraints, c)
		}
	}

	// Get maximum existing version
	var maxExistingVersion *semver.Version
	for _, dep := range config.Metadata.Dependencies {
		if dep.ID == "lein" {
			v, err := semver.NewVersion(dep.Version)
			if err == nil {
				if maxExistingVersion == nil || v.GreaterThan(maxExistingVersion) {
					maxExistingVersion = v
				}
			}
		}
	}

	// Fetch GitHub releases, which carry the lein release script
	releases, err := fetchReleases("technomancy", "leiningen")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching releases: %v\n", err)
		os.Exit(1)
	}

	type candidate struct {
		tag     string
		version *semver.Version
	}
	var candidates []candidate
	for _, tag := range releases {
		v, err := semver.NewVersion(strings.TrimPrefix(tag, "v"))
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", tag)
			continue
		}
		if maxExistingVersion != nil && !v.GreaterThan(maxExistingVersion) {
			fmt.Printf("Skipping %s: not newer than max existing version %s\n", v.String(), maxExistingVersion.String())
			continue
		}
		if !matchesConstraints(v, constraints) {
			continue
		}
		candidates = append(candidates, candidate{tag: tag, version: v})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].version.LessThan(candidates[j].version)
	})

	var output []OutputMetadata

	for _, c := range candidates {
		versionStr := c.version.String()

		uri := fmt.Sprintf("https://raw.githubusercontent.com/technomancy/leiningen/%s/bin/lein", c.tag)

		// Compute checksums
		fmt.Printf("Processing version %s...\n", versionStr)
		checksum, err := computeChecksum(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to checksum uri for %s: %v\n", versionStr, err)
			continue
		}

		sourceURL := fmt.Sprintf("https://github.com/technomancy/leiningen/archive/refs/tags/%s.tar.gz", c.tag)
		sourceChecksum, err := computeChecksum(sourceURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to checksum source for %s: %v\n", versionStr, err)
			continue
		}

		cpe := fmt.Sprintf("cpe:2.3:a:leiningen:leiningen:%s:*:*:*:*:*:*:*", versionStr)
		purl := fmt.Sprintf("pkg:generic/leiningen@%s", versionStr)

		licenses := []map[string]string{
			{
				"type": "EPL-1.0",
				"uri":  "https://github.com/technomancy/leiningen/blob/stable/COPYING",
			},
		}

		output = append(output, OutputMetadata{
			ID:             "lein",
			Name:           "Leiningen",
			Version:        versionStr,
			URI:            uri,
			Checksum:       "sha256:" + checksum,
			Source:         sourceURL,
			SourceChecksum: "sha256:" + sourceChecksum,
			CPE:            cpe,
			PURL:           purl,
			Licenses:       licenses,
			Stacks:         []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny", "*"},
		})
	}

	// Write output
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote %d dependency entries to %s\n", len(output), outputPath)
}

type OutputMetadata struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Version        string              `json:"version"`
	URI            string              `json:"uri"`
	Checksum       string              `json:"checksum"`
	Source         string              `json:"source,omitempty"`
	SourceChecksum string              `json:"source-checksum,omitempty"`
	CPE            string              `json:"cpe,omitempty"`
	PURL           string              `json:"purl,omitempty"`
	Licenses       []map[string]string `json:"licenses,omitempty"`
	Stacks         []string            `json:"stacks,omitempty"`
}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
}

func fetchReleases(owner, repo string) ([]string, error) {
	var tags []string
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d&per_page=100", owner, repo, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
		}

		var releases []GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, err
		}

		if len(releases) == 0 {
			break
		}

		for _, r := range releases {
			if r.Prerelease {
				continue
			}
			tags = append(tags, r.TagName)
		}

		page++
	}

	return tags, nil
}

func computeChecksum(uri string) (string, error) {
	resp, err := httpClient.Get(uri)
	if err != nil {
		return "", fmt.Errorf("unable to download %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to download %s: status %d", uri, resp.StatusCode)
	}

	h := sha256.New()
	if _, err := io.Copy(h, resp.Body); err != nil {
		return "", fmt.Errorf("unable to read %s: %w", uri, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func matchesConstraints(v *semver.Version, constraints []cargo.ConfigMetadataDependencyConstraint) bool {
	if len(constraints) == 0 {
		return true
	}
	for _, c := range constraints {
		cstr, err := semver.NewConstraint(c.Constraint)
		if err != nil {
			continue
		}
		if cstr.Check(v) {
			return true
		}
	}
	return false
}
