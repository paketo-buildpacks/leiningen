package lein

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
)

type Detect struct{}

func (Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	file := filepath.Join(context.Application.Path, "project.clj")
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return libcnb.DetectResult{Pass: false}, nil
	} else if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to determine if %s exists\n%w", file, err)
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "jvm-application"},
					{Name: "lein"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "jdk"},
					{Name: "lein"},
				},
			},
		},
	}, nil
}
