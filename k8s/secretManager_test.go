package k8s_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	podv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"md.com/azure-sas-token-extractor/k8s"
	"md.com/azure-sas-token-extractor/k8s/mock"
)

var _ = Describe("SecretManager", func() {
	var (
		secretManager k8s.SecretManager
		mockCtrl      *gomock.Controller
		mockClient    *mock_k8s.MockClientI
	)
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock_k8s.NewMockClientI(mockCtrl)
		secretManager = k8s.SecretManager{
			K8sClient: mockClient,
		}

	})

	Describe("Extract azure SAS-Token from k8s secrets", func() {
		Context("with azure aks secret available", func() {
			It("should extract aks sas token from k8s secrets", func() {
				//GIVEN
				var namespacesList *v1.NamespaceList
				_ = yaml.Unmarshal([]byte(namespaceStringYaml), &namespacesList)

				mockClient.EXPECT().GetNamespaces().Return(namespacesList, nil)

				var secretNs1List *podv1.SecretList
				err := yaml.Unmarshal([]byte(secretStringNamespace1), &secretNs1List)
				Expect(err).ToNot(HaveOccurred())
				mockClient.EXPECT().GetSecrets("namespace-1").Return(secretNs1List, nil).Times(1)

				var secretNs1Lis2 *podv1.SecretList
				err = yaml.Unmarshal([]byte(secretStringNamespace2), &secretNs1Lis2)
				Expect(err).ToNot(HaveOccurred())
				mockClient.EXPECT().GetSecrets("namespace-2").Return(secretNs1Lis2, nil).Times(1)

				var secretNs1List3 *podv1.SecretList
				err = yaml.Unmarshal([]byte(secretStringNamespace3), &secretNs1List3)
				Expect(err).ToNot(HaveOccurred())
				mockClient.EXPECT().GetSecrets("namespace-3").Return(secretNs1List3, nil).Times(1)

				//WHEN
				azureAksSecrets, errs := secretManager.RetrieveAzureAksSecret()

				//THEN
				Expect(errs).ToNot(HaveOccurred())
				Expect(azureAksSecrets).To(HaveLen(4))
				
				// Validate the first secret from namespace-1
				Expect(azureAksSecrets[0].Namespace).To(Equal("namespace-1"))
				Expect(azureAksSecrets[0].SecretName).To(Equal("my-secret"))
				Expect(azureAksSecrets[0].SecretKey).To(Equal("MY_AZURE_SAS_TOKEN"))
				Expect(azureAksSecrets[0].ExpirationDate.Format("2006-01-02")).To(Equal("2024-11-16"))
				
				// Validate the second secret from namespace-1
				Expect(azureAksSecrets[1].Namespace).To(Equal("namespace-1"))
				Expect(azureAksSecrets[1].SecretName).To(Equal("my-secret"))
				Expect(azureAksSecrets[1].SecretKey).To(Equal("MY_AZURE_SAS_TOKEN2"))
				Expect(azureAksSecrets[1].ExpirationDate.Format("2006-01-02")).To(Equal("2025-03-13"))
				
				// Validate the first secret from namespace-2
				Expect(azureAksSecrets[2].Namespace).To(Equal("namespace-2"))
				Expect(azureAksSecrets[2].SecretName).To(Equal("my-secret"))
				Expect(azureAksSecrets[2].SecretKey).To(Equal("MY_AZURE_SAS_TOKEN"))
				Expect(azureAksSecrets[2].ExpirationDate.Format("2006-01-02")).To(Equal("2024-11-16"))
				
				// Validate the second secret from namespace-2
				Expect(azureAksSecrets[3].Namespace).To(Equal("namespace-2"))
				Expect(azureAksSecrets[3].SecretName).To(Equal("my-secret"))
				Expect(azureAksSecrets[3].SecretKey).To(Equal("MY_AZURE_SAS_TOKEN2"))
				Expect(azureAksSecrets[3].ExpirationDate.Format("2006-01-02")).To(Equal("2025-03-13"))

			})
		})
	})

	Describe("extractExpirationDate", func() {
		Context("with valid SAS token", func() {
			It("should extract the expiration date correctly", func() {
				// Given a valid SAS token with a known expiration date
				token := "sp=r&st=2024-11-16T09:11:55Z&se=2024-11-16T17:11:55Z&spr=https&sv=2022-11-02&sr=c&sig=q44i29V8tYXU7YISTNcoKPPedQfbyyGh7cDxKJlx%2FEk%3D"

				// When extracting the expiration date
				expDate, err := k8s.ExtractExpirationDate(token)

				// Then no error should occur and the date should be correctly parsed
				Expect(err).ToNot(HaveOccurred())
				Expect(expDate.Format("2006-01-02")).To(Equal("2024-11-16"))
			})
		})

		Context("with invalid SAS token", func() {
			It("should return an error for malformed token", func() {
				// Given an invalid token
				token := "invalid-token"

				// When extracting the expiration date
				_, err := k8s.ExtractExpirationDate(token)

				// Then an error should occur
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("isSasToken", func() {
		Context("with valid SAS token", func() {
			It("should identify a valid SAS token", func() {
				// Given a valid SAS token
				token := "sp=r&st=2024-11-16T09:11:55Z&se=2024-11-16T17:11:55Z&spr=https&sv=2022-11-02&sr=c&sig=q44i29V8tYXU7YISTNcoKPPedQfbyyGh7cDxKJlx%2FEk%3D"

				// When checking if it's a SAS token
				result := k8s.IsSasToken(token)

				// Then it should be identified as a SAS token
				Expect(result).To(BeTrue())
			})
		})

		Context("with non-SAS token", func() {
			It("should not identify short strings as SAS tokens", func() {
				// Given a short string
				token := "short-string"

				// When checking if it's a SAS token
				result := k8s.IsSasToken(token)

				// Then it should not be identified as a SAS token
				Expect(result).To(BeFalse())
			})

			It("should not identify long strings without 'se=' as SAS tokens", func() {
				// Given a long string without 'se='
				token := "thisisalongstringwithoutsepatternbutitshouldbemorethan120characterssoletsmakeitreallyreallylonglikethisandaddmorestufftoreachrequiredlengthandstill123456789"

				// When checking if it's a SAS token
				result := k8s.IsSasToken(token)

				// Then it should not be identified as a SAS token
				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("tryToExtractAzureAksSasTokenFromK8sSecret", func() {
		Context("with secret containing SAS token", func() {
			It("should extract SAS tokens from secret data", func() {
				// Given a secret with SAS token
				var secret podv1.Secret
				_ = yaml.Unmarshal([]byte(`
apiVersion: v1
data:
  MY_AZURE_SAS_TOKEN: c3A9ciZzdD0yMDI0LTExLTE2VDA5OjExOjU1WiZzZT0yMDI0LTExLTE2VDE3OjExOjU1WiZzcHI9aHR0cHMmc3Y9MjAyMi0xMS0wMiZzcj1jJnNpZz1xNDRpMjlWOHRZWFU3WUlTS05jb0tQUGVkUWZieXlHaDdjRHhLSmx4JTJGRWslM0QK
  ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
kind: Secret
metadata:
  name: my-secret
  namespace: test-namespace
`), &secret)

				// When extracting SAS tokens
				result, err := k8s.TryToExtractAzureAksSasTokenFromK8sSecret(secret, "test-namespace")

				// Then no error should occur and one secret should be extracted
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(HaveLen(1))
				Expect(result[0].Namespace).To(Equal("test-namespace"))
				Expect(result[0].SecretName).To(Equal("my-secret"))
				Expect(result[0].SecretKey).To(Equal("MY_AZURE_SAS_TOKEN"))
				Expect(result[0].ExpirationDate.Format("2006-01-02")).To(Equal("2024-11-16"))
			})
		})

		Context("with secret not containing SAS token", func() {
			It("should not extract any SAS tokens", func() {
				// Given a secret with no SAS tokens
				var secret podv1.Secret
				_ = yaml.Unmarshal([]byte(`
apiVersion: v1
data:
  REGULAR_SECRET: UkVHVUxBUl9TRUNSRVQK
kind: Secret
metadata:
  name: regular-secret
  namespace: test-namespace
`), &secret)

				// When extracting SAS tokens
				result, err := k8s.TryToExtractAzureAksSasTokenFromK8sSecret(secret, "test-namespace")

				// Then no error should occur and no secrets should be extracted
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeEmpty())
			})
		})
	})
})

const namespaceStringYaml = `
apiVersion: v1
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: namespace-1
- apiVersion: v1
  kind: Namespace
  metadata:
    name: namespace-2
- apiVersion: v1
  kind: Namespace
  metadata:
    name: namespace-3
`
const secretStringNamespace1 = `
apiVersion: v1
items:
- apiVersion: v1
  data:
    MY_AZURE_SAS_TOKEN: c3A9ciZzdD0yMDI0LTExLTE2VDA5OjExOjU1WiZzZT0yMDI0LTExLTE2VDE3OjExOjU1WiZzcHI9aHR0cHMmc3Y9MjAyMi0xMS0wMiZzcj1jJnNpZz1xNDRpMjlWOHRZWFU3WUlTS05jb0tQUGVkUWZieXlHaDdjRHhLSmx4JTJGRWslM0QK
    ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
  kind: Secret
  metadata:
    name: my-secret
    namespace: namespace-1
- apiVersion: v1
  data:
    MY_AZURE_SAS_TOKEN2: c3A9ciZzdD0yMDI0LTExLTE2VDA5OjExOjU1WiZzZT0yMDI1LTAzLTEzVDE3OjExOjU1WiZzcHI9aHR0cHMmc3Y9MjAyMi0xMS0wMiZzcj1jJnNpZz1WSEVaQkd0cGRkZnJiNENiWkk0aTJpTXR1JTJGUmElMkZYM0dzdzRZTyUyRnBmc28wJTNECg==
    ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
  kind: Secret
  metadata:
    name: my-secret
    namespace: namespace-1
kind: List
metadata:
  resourceVersion: ""
`
const secretStringNamespace2 = `
apiVersion: v1
items:
- apiVersion: v1
  data:
    MY_AZURE_SAS_TOKEN: c3A9ciZzdD0yMDI0LTExLTE2VDA5OjExOjU1WiZzZT0yMDI0LTExLTE2VDE3OjExOjU1WiZzcHI9aHR0cHMmc3Y9MjAyMi0xMS0wMiZzcj1jJnNpZz1xNDRpMjlWOHRZWFU3WUlTS05jb0tQUGVkUWZieXlHaDdjRHhLSmx4JTJGRWslM0QK
    ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
  kind: Secret
  metadata:
    name: my-secret
    namespace: namespace-2
- apiVersion: v1
  data:
    MY_AZURE_SAS_TOKEN2: c3A9ciZzdD0yMDI0LTExLTE2VDA5OjExOjU1WiZzZT0yMDI1LTAzLTEzVDE3OjExOjU1WiZzcHI9aHR0cHMmc3Y9MjAyMi0xMS0wMiZzcj1jJnNpZz1WSEVaQkd0cGRkZnJiNENiWkk0aTJpTXR1JTJGUmElMkZYM0dzdzRZTyUyRnBmc28wJTNECg==
    ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
  kind: Secret
  metadata:
    name: my-secret
    namespace: namespace-2
kind: List
metadata:
  resourceVersion: ""
`
const secretStringNamespace3 = `
apiVersion: v1
items:
- apiVersion: v1
  data:
    ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
  kind: Secret
  metadata:
    name: my-secret
    namespace: namespace-3
- apiVersion: v1
  data:
    ANOTHER_SECRET: QU5PVEhFUl9TRUNSRVQK
  kind: Secret
  metadata:
    name: my-secret
    namespace: namespace-3
kind: List
metadata:
  resourceVersion: ""

`
