package k8s

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

type SecretManagerOptions struct {
	PrometheusEnabled bool
}
type SecretManager struct {
	k8sClient *Client
}

type AzureAksSecret struct {
	Namespace      string
	SecretName     string
	SecretKey      string
	SecretValue    string
	ExpirationDate time.Time
	RemainingDays  int
}

func NewSecretManager() *SecretManager {
	k8sClient := NewK8sClient()

	return &SecretManager{
		k8sClient: k8sClient,
	}
}

var exludedNamespaces = map[string]bool{
	"kube-system": true,
}

func (s SecretManager) RetrieveAzureAksSecret() ([]AzureAksSecret, error) {

	namespcesList, err := s.k8sClient.GetNamespaces()

	if err != nil {
		return nil, err
	}

	var azureAksSecretsList []AzureAksSecret

	for _, namespace := range namespcesList.Items {

		if _, ok := exludedNamespaces[namespace.Name]; ok {
			log.Debugf("Skipping namespace %s", namespace.Name)
			continue
		}
		log.Infof("Extracting secrets from namespace %s", namespace.Name)
		secrets, err := s.k8sClient.GetSecrets(namespace.Name)
		if err != nil {
			return nil, err
		}
		log.Infof("Namespace %s has %d secrets", namespace.Name, len(secrets.Items))
		for _, secret := range secrets.Items {
			azureAksSecret, err := tryToExtractAzureAksSasTokenFromK8sSecret(secret, namespace.Name)
			if err != nil {
				log.Errorf("Error while extracting Azure AKS secret from k8s secret %s - %s", secret.Name, err)
				continue
			}
			azureAksSecretsList = append(azureAksSecretsList, azureAksSecret...)
		}
	}
	printAllSecretInAWellFormatedTable(azureAksSecretsList)
	return azureAksSecretsList, err
}

func tryToExtractAzureAksSasTokenFromK8sSecret(secret v1.Secret, namespace string) ([]AzureAksSecret, error) {
	log.Infof("Analysing secrets %s from namespace %s", secret.Name, namespace)
	var azureAksSecrets []AzureAksSecret
	for key, value := range secret.Data {
		if !isSasToken(string(value)) {
			continue
		}
		log.Infof("key : %s", key)
		//log.Infof("value : %s", value)

		expirationDate, err := extractExpirationDate(string(value))
		if err != nil {
			log.Errorf("Error while extracting expiration date from secret %s with key %s", secret.Name, key)
			continue
		}
		azureAksSecret := AzureAksSecret{
			Namespace:      namespace,
			SecretName:     secret.Name,
			SecretKey:      key,
			SecretValue:    string(value),
			ExpirationDate: expirationDate,
			RemainingDays:  int(expirationDate.Sub(time.Now()).Hours() / 24),
		}
		azureAksSecrets = append(azureAksSecrets, azureAksSecret)

	}
	return azureAksSecrets, nil
}

func extractExpirationDate(token string) (time.Time, error) {
	tokenSplit := strings.SplitAfter(token, "se=")[1]
	return time.Parse("2006-01-02", tokenSplit[0:10])
}

func isSasToken(token string) bool {
	return len(token) > 120 && strings.Contains(token, "se=")
}

func printAllSecretInAWellFormatedTable(azureAksSecrets []AzureAksSecret) {
	time.Sleep(1 * time.Second)
	log.Infof("Total Azure AKS secrets retrieved : %d", len(azureAksSecrets))
	fmt.Printf("%-30s %-30s %-30s %-30s %-30s %-30s\n", "Namespace", "Secret Name", "Secret Key",
		"Expiration Date", "Remaining Days", "Status")
	for _, azureAksSecret := range azureAksSecrets {
		status := "Valid"
		if azureAksSecret.RemainingDays < 0 {
			status = "Expired"
		} else if azureAksSecret.RemainingDays < 30 {
			status = "Expiring soon"
		} else if azureAksSecret.RemainingDays > 365 {
			status = "Valid for more than a year"
		}

		if azureAksSecret.RemainingDays < 0 {
			fmt.Printf("\033[31m%-30s %-30s %-30s %-30s %-30d %-30s\033[0m\n", azureAksSecret.Namespace,
				azureAksSecret.SecretName, azureAksSecret.SecretKey,
				azureAksSecret.ExpirationDate.Format("2006-01-02"), azureAksSecret.RemainingDays, status)
		} else if azureAksSecret.RemainingDays < 30 {
			fmt.Printf("\033[33m%-30s %-30s %-30s %-30s %-30d %-30s\033[0m\n", azureAksSecret.Namespace,
				azureAksSecret.SecretName, azureAksSecret.SecretKey,
				azureAksSecret.ExpirationDate.Format("2006-01-02"), azureAksSecret.RemainingDays, status)
		} else {
			fmt.Printf("%-30s %-30s %-30s %-30s %-30d %-30s\n", azureAksSecret.Namespace,
				azureAksSecret.SecretName, azureAksSecret.SecretKey,
				azureAksSecret.ExpirationDate.Format("2006-01-02"), azureAksSecret.RemainingDays, status)
		}
	}
}
