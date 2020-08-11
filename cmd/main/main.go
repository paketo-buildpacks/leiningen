package main

import (
	"os"

	"github.com/eddumelendez/lein-paketo-buildpack/lein"
	"github.com/paketo-buildpacks/libbs"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

func main() {
	libpak.Main(
		lein.Detect{},
		lein.Build{
			Logger:             bard.NewLogger(os.Stdout),
			ApplicationFactory: libbs.NewApplicationFactory(),
		},
	)
}
