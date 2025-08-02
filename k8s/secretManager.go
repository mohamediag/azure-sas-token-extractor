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
	K8sClient ClientI
}

type AzureAksSecret struct {
	Namespace      string
	SecretName     string
	SecretKey      string
	SecretValue    string // masked for security
	ExpirationDate time.Time
	RemainingDays  int
}

func NewSecretManager() (*SecretManager, error) {
	k8sClient, err := NewK8sClient()
	if err != nil {
		return nil, err
	}

	return &SecretManager{
		K8sClient: k8sClient,
	}, nil
}

var excludedNamespaces = map[string]bool{
	"kube-system": true,
}

func (s SecretManager) RetrieveAzureAksSecret() ([]AzureAksSecret, error) {

	namespacesList, err := s.K8sClient.GetNamespaces()

	if err != nil {
		return nil, err
	}

	var azureAksSecretsList []AzureAksSecret

	for _, namespace := range namespacesList.Items {

		if _, ok := excludedNamespaces[namespace.Name]; ok {
			log.Debugf("Skipping namespace %s", namespace.Name)
			continue
		}
		log.Infof("Extracting secrets from namespace %s", namespace.Name)
		secrets, err := s.K8sClient.GetSecrets(namespace.Name)
		if err != nil {
			return nil, err
		}
		log.Infof("Namespace %s has %d secrets", namespace.Name, len(secrets.Items))
		for _, secret := range secrets.Items {
			azureAksSecret, err := TryToExtractAzureAksSasTokenFromK8sSecret(secret, namespace.Name)
			if err != nil {
				log.Errorf("Error while extracting Azure AKS secret from k8s secret %s - %s", secret.Name, err)
				continue
			}
			azureAksSecretsList = append(azureAksSecretsList, azureAksSecret...)
		}
	}
	printAllSecretInAWellFormatedTable(azureAksSecretsList)
	return azureAksSecretsList, nil
}

func TryToExtractAzureAksSasTokenFromK8sSecret(secret v1.Secret, namespace string) ([]AzureAksSecret, error) {
	log.Infof("Analysing secrets %s from namespace %s", secret.Name, namespace)
	var azureAksSecrets []AzureAksSecret
	for key, value := range secret.Data {
		if !IsSasToken(string(value)) {
			continue
		}
		log.Infof("key : %s", key)
		//log.Infof("value : %s", value)

		expirationDate, err := ExtractExpirationDate(string(value))
		if err != nil {
			log.Errorf("Error while extracting expiration date from secret %s with key %s", secret.Name, key)
			continue
		}
		azureAksSecret := AzureAksSecret{
			Namespace:      namespace,
			SecretName:     secret.Name,
			SecretKey:      key,
			SecretValue:    "***MASKED***",
			ExpirationDate: expirationDate,
			RemainingDays:  int(time.Until(expirationDate).Hours() / 24),
		}
		azureAksSecrets = append(azureAksSecrets, azureAksSecret)

	}
	return azureAksSecrets, nil
}

func ExtractExpirationDate(token string) (time.Time, error) {
	parts := strings.SplitAfter(token, "se=")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid SAS token format: missing 'se=' parameter")
	}

	tokenSplit := parts[1]
	if len(tokenSplit) < 10 {
		return time.Time{}, fmt.Errorf("invalid SAS token format: date part too short")
	}

	// Try different date formats
	dateStr := tokenSplit[0:10]
	formats := []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05"}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid SAS token date format: %s", dateStr)
}

func IsSasToken(token string) bool {
	return len(token) > 120 && strings.Contains(token, "se=")
}

func printAllSecretInAWellFormatedTable(azureAksSecrets []AzureAksSecret) {
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
