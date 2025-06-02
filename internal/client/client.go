package client

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Builds and returns a Kubernetes clientset from provided ConfigFlags.
func GetClientSet(flags *genericclioptions.ConfigFlags) (*kubernetes.Clientset, error) {
	restConfig, err := flags.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}
