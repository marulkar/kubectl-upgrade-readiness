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

// Extracts and validates addon image versions for known control plane components
func CheckAddonCompatibility(cs *kubernetes.Clientset, target semver.Version, raw string) {
	fmt.Println("\nControl Plane Addon Compatibility:")
	pods, err := cs.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error fetching pods: %v\n", err)
		return
	}

	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			image := c.Image
			fmt.Printf("  %s => %s\n", pod.Name, image)

			parts := strings.Split(image, ":")
			if len(parts) < 2 {
				continue
			}
			tag := strings.TrimPrefix(parts[1], "v")
			if strings.Count(tag, ".") == 1 {
				tag += ".0"
			}
			v, err := semver.Parse(tag)
			if err != nil {
				continue
			}
			switch path.Base(parts[0]) {
			case "coredns":
				if v.Major == 1 && v.Minor < 10 {
					fmt.Printf("    [!] upgrade CoreDNS to >=v1.10 for v%s\n", raw)
				}
			case "metrics-server":
				if v.Major == 0 && v.Minor == 5 {
					fmt.Printf("    [!] upgrade metrics-server to >=v0.6 for v%s\n", raw)
				}
			case "kube-proxy":
				if v.Major != target.Major || v.Minor > target.Minor || target.Minor-v.Minor > 4 {
					fmt.Printf("    [!] kube-proxy %s not within skew of v%s\n", tag, raw)
				}
			}
		}
	}
}
