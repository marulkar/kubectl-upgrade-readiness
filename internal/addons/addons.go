package addons

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// versionedCompatMatrix unchanged – shortened for brevity (keep current content)
var versionedCompatMatrix = map[string]map[string][]string{
	"v1.28": {"etcd": {"3.5.6", "3.5.7", "3.5.8", "3.5.9"}, "coredns": {"1.10.0", "1.10.1"}, "metrics-server": {"0.6.0", "0.6.1"}, "kube-proxy": {"1.25.0", "1.26.0", "1.27.0", "1.28.0"}},
	"v1.29": {"etcd": {"3.5.7", "3.5.8", "3.5.9", "3.5.10"}, "coredns": {"1.10.1", "1.11.0"}, "metrics-server": {"0.6.1", "0.6.2"}, "kube-proxy": {"1.26.0", "1.27.0", "1.28.0", "1.29.0"}},
	"v1.30": {"etcd": {"3.5.9", "3.5.10", "3.5.11"}, "coredns": {"1.11.0", "1.11.1"}, "metrics-server": {"0.6.2", "0.7.0"}, "kube-proxy": {"1.27.0", "1.28.0", "1.29.0", "1.30.0"}},
	"v1.31": {"etcd": {"3.5.10", "3.5.11", "3.5.12"}, "coredns": {"1.11.1", "1.12.0"}, "metrics-server": {"0.7.0"}, "kube-proxy": {"1.31.0", "1.30.0", "1.29.0", "1.28.0", "1.27.0"}},
	"v1.32": {"etcd": {"3.5.11", "3.5.12", "3.5.13"}, "coredns": {"1.12.0"}, "metrics-server": {"0.7.0", "0.7.1"}, "kube-proxy": {"1.32.0", "1.31.0", "1.30.0", "1.29.0", "1.28.0"}},
	"v1.33": {"etcd": {"3.5.13", "3.5.14"}, "coredns": {"1.12.0", "1.13.0"}, "metrics-server": {"0.7.1", "0.7.2"}, "kube-proxy": {"1.33.0", "1.32.0", "1.31.0", "1.30.0", "1.29.0"}},
}

var tagSuffixRE = regexp.MustCompile(`^v?([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)

func CheckAddonCompatibility(cs *kubernetes.Clientset, target semver.Version, raw string) {
	key := fmt.Sprintf("v%d.%d", target.Major, target.Minor)
	matrix, ok := versionedCompatMatrix[key]
	if !ok {
		fmt.Printf("\n[!] No compatibility data for %s\n", key)
		return
	}

	fmt.Println("\nControl Plane Addon Compatibility:")
	pods, err := cs.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing kube-system pods: %v\n", err)
		return
	}

	for _, p := range pods.Items {
		for _, c := range p.Spec.Containers {
			img := c.Image
			name, version := parseImage(img)
			if name == "" || version == "" {
				continue
			}
			expect, tracked := matrix[name]
			if !tracked {
				continue
			}
			compliant := contains(expect, version)
			if compliant {
				fmt.Printf("  ✅ %s: %s\n", name, version)
			} else {
				fmt.Printf("  ❌ %s: %s (Expected: %v)\n", name, version, expect)
			}
		}
	}
}

func parseImage(ref string) (addon, ver string) {
	// strip digest if present
	before := ref
	if idx := strings.IndexRune(ref, '@'); idx != -1 {
		before = ref[:idx]
	}

	// tag after ':' if present
	tag := ""
	if idx := strings.LastIndex(before, ":"); idx != -1 {
		tag = before[idx+1:]
		before = before[:idx]
	}

	base := path.Base(before)
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

	// If tag missing or non‑semver, try to pull version from basename suffix like kube-proxy_1_25_16
	if tag == "" || !strings.Contains(tag, ".") {
		if idx := strings.LastIndex(base, "_"); idx != -1 {
			tag = strings.ReplaceAll(base[idx+1:], "_", ".")
		}
	}

	if tag == "" {
		return addon, ""
	}

	// normalise tag to x.y.z and strip suffixes
	m := tagSuffixRE.FindStringSubmatch(tag)
	if len(m) < 2 {
		return addon, ""
	}
	v := m[1]
	if strings.Count(v, ".") == 1 {
		v += ".0"
	}
	return addon, v
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
