package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/marulkar/kubectl-upgrade_readiness/internal/addons"
	"github.com/marulkar/kubectl-upgrade_readiness/internal/client"
	"github.com/marulkar/kubectl-upgrade_readiness/internal/kubelet"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	targetVersion string
	configFlags   *genericclioptions.ConfigFlags
)

// Parses CLI flags (including kubeconfig/context) and runs upgrade readiness checks.
func Execute() {
	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(pflag.CommandLine)

	pflag.StringVar(&targetVersion, "target-version", "v1.31", "Target Kubernetes version")
	pflag.BoolVar(&kubelet.Verbose, "verbose", false, "Show full list of nodes per version")
	pflag.Parse()

	fmt.Printf("kubectl-upgrade-readiness: MVP (target: %s)\n", targetVersion)

	v, err := semver.Parse(normalize(targetVersion))
	if err != nil {
		panic(err)
	}

	cs, err := client.GetClientSet(configFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	kubelet.CheckKubeletVersions(cs, v, targetVersion)
	addons.CheckAddonCompatibility(cs, v, targetVersion)
}

func normalize(v string) string {
	s := strings.TrimPrefix(v, "v")
	if len(strings.Split(s, ".")) == 2 {
		s += ".0"
	}
	return s
}
