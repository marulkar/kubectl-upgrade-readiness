package kubelet

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var verbose = flag.Bool("verbose", false, "Show full list of nodes per version")

// Validates kubelet versions across all nodes based on N-4 skew policy
func CheckKubeletVersions(cs *kubernetes.Clientset, target semver.Version, raw string) {
	flag.CommandLine.Parse(os.Args[1:])

	nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing nodes: %v\n", err)
		return
	}

	type nodeGroup struct {
		compliant bool
		nodeNames []string
	}

	skew := 4
	nodeGroups := make(map[string]*nodeGroup)

	for _, n := range nodes.Items {
		verStr := n.Status.NodeInfo.KubeletVersion
		v, err := semver.ParseTolerant(verStr)
		if err != nil {
			fmt.Printf("  [!] Unable to parse version for %s: %v\n", n.Name, err)
			continue
		}

		minorDelta := int(target.Minor) - int(v.Minor)
		compliant := v.Major == target.Major && minorDelta >= 0 && minorDelta <= skew

		if _, ok := nodeGroups[verStr]; !ok {
			nodeGroups[verStr] = &nodeGroup{compliant: compliant, nodeNames: []string{}}
		}
		nodeGroups[verStr].nodeNames = append(nodeGroups[verStr].nodeNames, n.Name)
	}

	fmt.Printf("\nKubelet Version Skew Check (target: %s):\n\n", raw)
	compliantVersions := []string{}
	nonCompliantVersions := []string{}

	for version, group := range nodeGroups {
		if group.compliant {
			compliantVersions = append(compliantVersions, version)
		} else {
			nonCompliantVersions = append(nonCompliantVersions, version)
		}
	}

	if len(nonCompliantVersions) > 0 {
		fmt.Println("❌ Non-compliant versions:")
		sort.Strings(nonCompliantVersions)
		for _, v := range nonCompliantVersions {
			group := nodeGroups[v]
			fmt.Printf("  - %s (%d nodes)\n", v, len(group.nodeNames))
			if *verbose {
				for _, name := range group.nodeNames {
					fmt.Printf("    • %s\n", name)
				}
			} else {
				fmt.Printf("    Examples:\n")
				sample := group.nodeNames
				if len(sample) > 3 {
					sample = sample[:3]
				}
				for _, name := range sample {
					fmt.Printf("      • %s\n", name)
				}
				fmt.Println("    ... (use --verbose to see full list)")
			}
		}
		fmt.Println()
	} else {
		fmt.Println("✅ All nodes are compliant with kubelet skew policy.")
	}

	if len(compliantVersions) > 0 {
		fmt.Println("✅ Compliant versions:")
		sort.Strings(compliantVersions)
		for _, v := range compliantVersions {
			group := nodeGroups[v]
			fmt.Printf("  - %s (%d nodes)\n", v, len(group.nodeNames))
		}
		fmt.Println()
	}
}
