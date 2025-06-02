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
	verbose       bool
	configFlags   *genericclioptions.ConfigFlags
)

// Parses CLI flags (plugin only) and runs upgrade‑readiness checks.
func Execute() {
	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(pflag.CommandLine)

	pluginFlags := pflag.NewFlagSet("upgrade_readiness", pflag.ExitOnError)
	pluginFlags.StringVarP(&targetVersion, "target-version", "t", "v1.31", "Target Kubernetes version")
	pluginFlags.BoolVar(&verbose, "verbose", false, "Show full list of nodes per version")

	pflag.CommandLine.AddFlagSet(pluginFlags)

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "kubectl upgrade_readiness validates kubelet skew and core‑addon compatibility before a Kubernetes version upgrade.\n\nUsage:\n")
		pluginFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n(Advanced kubeconfig flags such as --kubeconfig and --context are accepted but hidden)\n")
	}

	pluginFlags.Parse(os.Args[1:])
	pflag.Parse()

	fmt.Printf("kubectl-upgrade-readiness: MVP (target: %s)\n", targetVersion)

	v, err := semver.Parse(normalize(targetVersion))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cs, err := client.GetClientSet(configFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	kubelet.CheckKubeletVersions(cs, v, targetVersion, verbose)
	addons.CheckAddonCompatibility(cs, v, targetVersion, verbose)
}

func normalize(v string) string {
	s := strings.TrimPrefix(v, "v")
	if len(strings.Split(s, ".")) == 2 {
		s += ".0"
	}
	return s
}
