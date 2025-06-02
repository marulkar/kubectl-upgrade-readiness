// kubectl-upgrade-readiness: MVP in Go
// Verifies kubelet/kube-proxy skew (N-4 rule) and addon versions.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/blang/semver/v4"
)

var targetVersion string

// Handles both in-cluster and out-of-cluster config detection
func getClientSet() (*kubernetes.Clientset, error) {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(cfg)
	}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func parseSemver(v string) (semver.Version, error) {
	s := strings.TrimPrefix(v, "v")
	if strings.Count(s, ".") == 1 {
		s += ".0"
	}
	return semver.Parse(s)
}

// Validates kubelet versions across all nodes based on N-4 skew policy
func checkKubeletVersions(client *kubernetes.Clientset, target semver.Version) {
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing nodes: %v\n", err)
		return
	}

	fmt.Println("\nKubelet Version Skew Check:")
	allOK := true
	for _, node := range nodes.Items {
		verStr := node.Status.NodeInfo.KubeletVersion
		ver, err := parseSemver(verStr)
		fmt.Printf("  %s => %s\n", node.Name, verStr)
		if err != nil || ver.Major != target.Major || ver.Minor > target.Minor || target.Minor-ver.Minor > 4 {
			fmt.Printf("    [!] %s not within skew of %s\n", verStr, targetVersion)
			allOK = false
		}
	}
	if allOK {
		fmt.Println("  [OK] All kubelets compliant.")
	} else {
		fmt.Println("  [!] Some kubelets outside supported skew.")
	}
}

// Extracts and validates addon image versions for known control plane components
func checkAddonCompatibility(client *kubernetes.Clientset, target semver.Version) {
	fmt.Println("\nControl Plane Addon Compatibility:")
	pods, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing kube-system pods: %v\n", err)
		return
	}
	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			img := c.Image
			fmt.Printf("  %s => %s\n", pod.Name, img)
			parts := strings.Split(img, ":")
			name := path.Base(parts[0])
			if len(parts) < 2 {
				continue
			}
			tag := strings.TrimPrefix(parts[1], "v")
			if strings.Count(tag, ".") == 1 {
				tag += ".0"
			}
			v, err := parseSemver(tag)
			if err != nil {
				continue
			}
			switch name {
			case "coredns":
				if v.Major == 1 && v.Minor < 10 {
					fmt.Printf("    [!] upgrade CoreDNS to >=v1.10 for v%s\n", targetVersion)
				}
			case "metrics-server":
				if v.Major == 0 && v.Minor == 5 {
					fmt.Printf("    [!] upgrade metrics-server to >=v0.6 for v%s\n", targetVersion)
				}
			case "kube-proxy":
				if v.Major != target.Major || v.Minor > target.Minor || target.Minor-v.Minor > 4 {
					fmt.Printf("    [!] kube-proxy %s not within skew of v%s\n", tag, targetVersion)
				}
			}
		}
	}
}

func main() {
	flag.StringVar(&targetVersion, "target-version", "v1.31", "Target Kubernetes version")
	flag.Parse()

	target, err := parseSemver(targetVersion)
	if err != nil {
		panic(err)
	}

	client, err := getClientSet()
	if err != nil {
		panic(err)
	}

	fmt.Printf("kubectl-upgrade-readiness: MVP (target: %s)\n", targetVersion)
	checkKubeletVersions(client, target)
	checkAddonCompatibility(client, target)
}
