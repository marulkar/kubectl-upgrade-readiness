package addons

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/marulkar/kubectl-upgrade_readiness/internal/matrix"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var tagRegex = regexp.MustCompile(`^v?([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)

func CheckAddonCompatibility(cs *kubernetes.Clientset, target semver.Version, raw string, verbose bool) {
	key := fmt.Sprintf("v%d.%d", target.Major, target.Minor)
	compatMatrix, ok := matrix.VersionedCompatMatrix[key]
	if !ok {
		fmt.Printf("\n[!] No compatibility data for %s\n", key)
		return
	}

	pods, err := cs.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing kube-system pods: %v\n", err)
		return
	}

	group := map[string]map[string][]string{}

	for _, p := range pods.Items {
		for _, c := range p.Spec.Containers {
			addon, ver := parseImage(c.Image)
			if addon == "" || ver == "" {
				continue
			}
			if _, tracked := compatMatrix[addon]; !tracked {
				continue
			}
			if _, ok := group[addon]; !ok {
				group[addon] = map[string][]string{}
			}
			group[addon][ver] = append(group[addon][ver], p.Name)
		}
	}

	if len(group) == 0 {
		fmt.Println("\nControl Plane Addon Compatibility: <none found>")
		return
	}

	fmt.Println("\nControl Plane Addon Compatibility:")

	addons := make([]string, 0, len(group))
	for a := range group {
		addons = append(addons, a)
	}
	sort.Strings(addons)

	for _, a := range addons {
		versions := make([]string, 0, len(group[a]))
		for v := range group[a] {
			versions = append(versions, v)
		}
		sort.Strings(versions)

		expected := compatMatrix[a]
		for _, v := range versions {
			pods := group[a][v]
			compliant := contains(expected, v)
			icon := "✅"
			if !compliant {
				icon = "❌"
			}
			fmt.Printf("  %s %s: %s (%d pods", icon, a, v, len(pods))
			if !compliant {
				fmt.Printf(", expected: %v", expected)
			}
			fmt.Println(")")

			if verbose {
				sort.Strings(pods)
				for _, n := range pods {
					fmt.Printf("    • %s\n", n)
				}
			}
		}
	}
}

func parseImage(ref string) (addon, ver string) {
	if i := strings.IndexRune(ref, '@'); i != -1 {
		ref = ref[:i]
	}
	tag := ""
	if i := strings.LastIndex(ref, ":"); i != -1 {
		tag = ref[i+1:]
		ref = ref[:i]
	}

	base := path.Base(ref)
	lower := strings.ToLower(base)
	switch {
	case strings.Contains(lower, "kube-proxy"):
		addon = "kube-proxy"
	case strings.Contains(lower, "coredns"):
		addon = "coredns"
	case strings.Contains(lower, "metrics-server"):
		addon = "metrics-server"
	case strings.Contains(lower, "etcd"):
		addon = "etcd"
	default:
		return "", ""
	}

	if tag == "" || !strings.Contains(tag, ".") {
		if idx := strings.LastIndex(base, "_"); idx != -1 {
			tag = strings.ReplaceAll(base[idx+1:], "_", ".")
		}
	}
	if tag == "" {
		return addon, ""
	}

	m := tagRegex.FindStringSubmatch(tag)
	if len(m) < 2 {
		return addon, ""
	}
	ver = m[1]
	if strings.Count(ver, ".") == 1 {
		ver += ".0"
	}
	return
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
