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
			It("should extract aks sas token from k8s secrets)", func() {
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
				//TODO : add more assertions

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
