package addons

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var versionedCompatMatrix = map[string]map[string][]string{
	"v1.28": {
		"etcd":           {"3.5.6", "3.5.7", "3.5.8", "3.5.9"},
		"coredns":        {"1.10.0", "1.10.1"},
		"metrics-server": {"0.6.0", "0.6.1"},
		"kube-proxy":     {"1.25.0", "1.26.0", "1.27.0", "1.28.0"},
	},
	"v1.29": {
		"etcd":           {"3.5.7", "3.5.8", "3.5.9", "3.5.10"},
		"coredns":        {"1.10.1", "1.11.0"},
		"metrics-server": {"0.6.1", "0.6.2"},
		"kube-proxy":     {"1.26.0", "1.27.0", "1.28.0", "1.29.0"},
	},
	"v1.30": {
		"etcd":           {"3.5.9", "3.5.10", "3.5.11"},
		"coredns":        {"1.11.0", "1.11.1"},
		"metrics-server": {"0.6.2", "0.7.0"},
		"kube-proxy":     {"1.27.0", "1.28.0", "1.29.0", "1.30.0"},
	},
	"v1.31": {
		"etcd":           {"3.5.10", "3.5.11", "3.5.12"},
		"coredns":        {"1.11.1", "1.12.0"},
		"metrics-server": {"0.7.0"},
		"kube-proxy":     {"1.31.0", "1.30.0", "1.29.0", "1.28.0", "1.27.0"},
	},
	"v1.32": {
		"etcd":           {"3.5.11", "3.5.12", "3.5.13"},
		"coredns":        {"1.12.0"},
		"metrics-server": {"0.7.0", "0.7.1"},
		"kube-proxy":     {"1.32.0", "1.31.0", "1.30.0", "1.29.0", "1.28.0"},
	},
	"v1.33": {
		"etcd":           {"3.5.13", "3.5.14"},
		"coredns":        {"1.12.0", "1.13.0"},
		"metrics-server": {"0.7.1", "0.7.2"},
		"kube-proxy":     {"1.33.0", "1.32.0", "1.31.0", "1.30.0", "1.29.0"},
	},
}

// Extracts and validates addon image versions for known control plane components
func CheckAddonCompatibility(client *kubernetes.Clientset, target semver.Version, targetVersionRaw string) {
	versionKey := fmt.Sprintf("v%d.%d", target.Major, target.Minor)
	validVersions, ok := versionedCompatMatrix[versionKey]
	if !ok {
		fmt.Printf("\n[!] No compatibility data found for %s\n", versionKey)
		return
	}

	fmt.Println("\nControl Plane Addon Compatibility:")
	pods, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing kube-system pods: %v\n", err)
		return
	}

	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			img := c.Image
			parts := strings.Split(img, ":")
			name := path.Base(parts[0])
			if len(parts) < 2 {
				continue
			}
			tag := strings.TrimPrefix(parts[1], "v")
			if strings.Count(tag, ".") == 1 {
				tag += ".0"
			}
			cleanedTag := strings.SplitN(tag, "-", 2)[0]

			expected, known := validVersions[name]
			if !known {
				continue
			}

			match := false
			for _, v := range expected {
				if v == cleanedTag {
					match = true
					break
				}
			}
			if !match {
				fmt.Printf("  ❌ %s: %s (Expected: %v)\n", name, cleanedTag, expected)
			} else {
				fmt.Printf("  ✅ %s: %s\n", name, cleanedTag)
			}
		}
	}
}
