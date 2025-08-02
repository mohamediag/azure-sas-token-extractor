# Azure SAS Token Extractor

A command-line tool to extract Azure SAS (Shared Access Signature) tokens from Kubernetes secrets and provide information about their expiration dates.

## Introduction

Azure SAS (Shared Access Signature) tokens provide secure delegated access to resources in Azure storage accounts. These tokens have expiration dates, and it's crucial to monitor and manage them to prevent application downtime caused by expired tokens.

Azure SAS Token Extractor helps Kubernetes cluster administrators and developers to:
- Discover all Azure SAS tokens stored in Kubernetes secrets across namespaces
- Extract and display expiration dates for each token
- Identify expired tokens or those that are about to expire
- Manage token lifecycles effectively

## Installation

### Prerequisites
- Go 1.22.0 or later
- Access to a Kubernetes cluster
- Proper Kubernetes RBAC permissions to read secrets across namespaces

### Install from source
```bash
# Clone the repository
git clone https://github.com/mohamediag/azure-sas-token-extractor.git

# Navigate to the project directory
cd azure-sas-token-extractor

# Build the project
go build -o azure-sas-token-extractor

# Move to a directory in your PATH (optional)
sudo mv azure-sas-token-extractor /usr/local/bin/
```

## Configuration

### Environment Variables

- `KUBECONFIG_PATH`: Path to your kubeconfig file (optional). If not set, the tool will attempt to use the in-cluster configuration.
  ```bash
  export KUBECONFIG_PATH=~/.kube/config
  ```

## Usage

### Basic Usage
```bash
azure-sas-token-extractor k8s get-azure-sas-tokens
```

This command will:
1. Connect to your Kubernetes cluster
2. Scan all namespaces (except excluded ones like kube-system)
3. Find all secrets containing Azure SAS tokens
4. Extract expiration dates from these tokens
5. Display the information in a formatted table

## Output Explanation

The command output displays a table with the following columns:

| Column          | Description                                   |
|-----------------|-----------------------------------------------|
| Namespace       | Kubernetes namespace containing the secret    |
| Secret Name     | Name of the Kubernetes secret                 |
| Secret Key      | Key within the secret that contains the token |
| Expiration Date | When the SAS token expires (YYYY-MM-DD)       |
| Remaining Days  | Number of days until expiration               |
| Status          | Current status of the token                   |

### Status Types

- **Expired** (shown in red): Token has already expired
- **Expiring soon** (shown in yellow): Token will expire within 30 days
- **Valid**: Token is valid and not expiring soon
- **Valid for more than a year**: Token has a long expiration period

## How It Works

Azure SAS Token Extractor identifies SAS tokens in Kubernetes secrets by:
1. Looking for strings longer than 120 characters
2. Checking if the string contains the pattern "se=" (which is used in SAS tokens to indicate the expiration time)
3. Extracting the expiration date from the "se=" parameter using multiple date format parsers

The tool calculates the number of days remaining until expiration and categorizes the token status accordingly.

**Security Note**: For security purposes, actual secret values are masked in the output to prevent accidental exposure of sensitive data.

## Troubleshooting

### Common Issues

- **Permission Errors**: Ensure your Kubernetes context has sufficient permissions to list and read secrets across namespaces.
  ```bash
  # Check your current permissions
  kubectl auth can-i get secrets --all-namespaces
  ```

- **No Tokens Found**: If no tokens are found, verify that your secrets actually contain Azure SAS tokens with the expected format.

- **Connection Issues**: Ensure your KUBECONFIG is correctly set and you can connect to the cluster. The tool will first try to use the `KUBECONFIG_PATH` environment variable, then fall back to in-cluster configuration.
  ```bash
  # Test your connection
  kubectl get nodes
  
  # Set KUBECONFIG_PATH if needed
  export KUBECONFIG_PATH=~/.kube/config
  ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.