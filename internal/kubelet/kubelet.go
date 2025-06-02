package kubelet

import (
	"context"
	"fmt"
	"sort"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CheckKubeletVersions groups nodes by version and reports N‑4 skew compliance.
func CheckKubeletVersions(cs *kubernetes.Clientset, target semver.Version, raw string, verbose bool) {
	nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing nodes: %v\n", err)
		return
	}

	type nodeGroup struct {
		compliant bool
		nodes     []string
	}

	const skew = 4
	groups := map[string]*nodeGroup{}

	for _, n := range nodes.Items {
		vstr := n.Status.NodeInfo.KubeletVersion
		v, err := semver.ParseTolerant(vstr)
		if err != nil {
			fmt.Printf("  [!] Parse error for %s: %v\n", n.Name, err)
			continue
		}
		delta := int(target.Minor) - int(v.Minor)
		ok := v.Major == target.Major && delta >= 0 && delta <= skew

		g, exists := groups[vstr]
		if !exists {
			g = &nodeGroup{compliant: ok}
			groups[vstr] = g
		}
		g.nodes = append(g.nodes, n.Name)
	}

	fmt.Printf("\nKubelet Version Skew Check (target: %s):\n\n", raw)

	var good, bad []string
	for v, g := range groups {
		if g.compliant {
			good = append(good, v)
		} else {
			bad = append(bad, v)
		}
	}
	sort.Strings(bad)
	sort.Strings(good)

	if len(bad) > 0 {
		fmt.Println("❌ Non‑compliant versions:")
		for _, v := range bad {
			g := groups[v]
			fmt.Printf("  - %s (%d nodes)\n", v, len(g.nodes))
			printExamples(g.nodes, verbose)
		}
		fmt.Println()
	} else {
		fmt.Println("✅ All nodes are compliant.")
	}

	if len(good) > 0 {
		fmt.Println("✅ Compliant versions:")
		for _, v := range good {
			g := groups[v]
			fmt.Printf("  - %s (%d nodes)\n", v, len(g.nodes))
			if verbose {
				for _, n := range g.nodes {
					fmt.Printf("    • %s\n", n)
				}
			}
		}
		fmt.Println()
	}
}

func printExamples(nodes []string, verbose bool) {
	if verbose {
		for _, n := range nodes {
			fmt.Printf("    • %s\n", n)
		}
		return
	}
	sample := nodes
	if len(sample) > 3 {
		sample = sample[:3]
	}
	fmt.Println("    Examples:")
	for _, n := range sample {
		fmt.Printf("      • %s\n", n)
	}
	if len(nodes) > 3 {
		fmt.Println("    ... (use --verbose to see full list)")
	}
}
