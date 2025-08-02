package k8s

import (
	"context"
	"fmt"
	podv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type Client struct {
	clientSet *kubernetes.Clientset
}
type ClientI interface {
	GetSecrets(namespace string) (*podv1.SecretList, error)
	GetNamespaces() (*podv1.NamespaceList, error)
}

func NewK8sClient() (*Client, error) {
	var config *rest.Config
	var err error
	kubeconfig := os.Getenv("KUBECONFIG_PATH")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	return &Client{clientSet: clientset}, nil
}

func (c Client) GetSecrets(namespace string) (*podv1.SecretList, error) {
	return c.clientSet.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
}

func (c Client) GetNamespaces() (*podv1.NamespaceList, error) {
	return c.clientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
}
