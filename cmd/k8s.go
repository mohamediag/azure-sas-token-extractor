package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"md.com/azure-sas-token-extractor/k8s"
)

func init() {

	k8s := &cobra.Command{
		Use:   "k8s",
		Short: "k8s",
	}

	secret := &cobra.Command{
		Use:   "get-azure-sas-tokens",
		Short: "get-azure-sas-tokens",
		Run: func(cmd *cobra.Command, args []string) {
			runk8sSecretCommand()
		},
	}

	k8s.AddCommand(secret)
	rootCmd.AddCommand(k8s)

}

func runk8sSecretCommand() {
	secretManager, err := k8s.NewSecretManager()
	if err != nil {
		logrus.Fatal("Error while creating secret manager: ", err)
	}

	azureAksSecrets, err := secretManager.RetrieveAzureAksSecret()
	if err != nil {
		logrus.Fatal("Error while retrieving Azure AKS secrets: ", err)
	}
	logrus.Infof("Azure AKS secrets retrieved successfully, total : %d", len(azureAksSecrets))
}
