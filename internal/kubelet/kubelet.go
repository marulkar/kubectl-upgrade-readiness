package kubelet

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Validates kubelet versions across all nodes based on N-4 skew policy
func CheckKubeletVersions(cs *kubernetes.Clientset, target semver.Version, raw string) {
	nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing nodes: %v\n", err)
		return
	}

	fmt.Println("\nKubelet Version Skew Check:")
	allOK := true
	for _, n := range nodes.Items {
		verStr := n.Status.NodeInfo.KubeletVersion
		v, err := semver.ParseTolerant(verStr)
		fmt.Printf("  %s => %s\n", n.Name, verStr)
		if err != nil || v.Major != target.Major || v.Minor > target.Minor || target.Minor-v.Minor > 4 {
			fmt.Printf("    [!] %s not within skew of %s\n", verStr, raw)
			allOK = false
		}
	}
	if allOK {
		fmt.Println("  [OK] All kubelets compliant.")
	} else {
		fmt.Println("  [!] Some kubelets outside supported skew.")
	}
}
