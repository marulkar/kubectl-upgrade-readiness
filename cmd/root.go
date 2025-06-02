package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/marulkar/kubectl-upgrade-readiness/internal/addons"
	"github.com/marulkar/kubectl-upgrade-readiness/internal/client"
	"github.com/marulkar/kubectl-upgrade-readiness/internal/kubelet"
)

var targetVersion string

func Execute() {
	flag.StringVar(&targetVersion, "target-version", "v1.31", "Target Kubernetes version")
	flag.Parse()

	fmt.Printf("kubectl-upgrade-readiness: MVP (target: %s)\n", targetVersion)

	v, err := semver.Parse(normalize(targetVersion))
	if err != nil {
		panic(err)
	}

	cs, err := client.GetClientSet()
	if err != nil {
		panic(err)
	}

	kubelet.CheckKubeletVersions(cs, v, targetVersion)
	addons.CheckAddonCompatibility(cs, v, targetVersion)
}

func normalize(v string) string {
	s := v
	if s[0] == 'v' {
		s = s[1:]
	}
	if parts := len(splitDots(s)); parts == 2 {
		s += ".0"
	}
	return s
}

func splitDots(s string) []string {
	return strings.Split(s, ".")
}
