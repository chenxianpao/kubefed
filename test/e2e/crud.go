/*
Copyright 2017 The Federation v2 Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"strings"

	"github.com/kubernetes-sigs/federation-v2/pkg/federatedtypes"
	"github.com/kubernetes-sigs/federation-v2/test/common"
	"github.com/kubernetes-sigs/federation-v2/test/e2e/framework"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Federated types", func() {
	f := framework.NewFederationFramework("federated-types")

	tl := framework.NewE2ELogger()

	typeConfigs, err := common.FederatedTypeConfigs()
	if err != nil {
		tl.Fatalf("Error loading type configs: %v", err)
	}

	for i, _ := range typeConfigs {
		// Bind the type config inside the loop to ensure the ginkgo
		// closure gets a different value for every loop iteration.
		//
		// Reference: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		typeConfig := typeConfigs[i]
		templateKind := typeConfig.GetTemplate().Kind

		Describe(fmt.Sprintf("%q resources", templateKind), func() {
			It("should be created, read, updated and deleted successfully", func() {
				// TODO (font): e2e tests for federated Namespace using a
				// test managed federation does not work until k8s
				// namespace controller is added.
				if framework.TestContext.TestManagedFederation &&
					federatedtypes.IsNamespaceKind(templateKind) {
					framework.Skipf("%s not supported for test managed federation.", templateKind)
				}

				// Initialize an in-memory controller if configuration requires
				f.SetUpControllerFixture(typeConfig)

				userAgent := fmt.Sprintf("test-%s-crud", strings.ToLower(templateKind))

				fedConfig := f.FedConfig()
				kubeConfig := f.KubeConfig()
				targetAPIResource := typeConfig.GetTarget()
				testClusters := f.ClusterClients(&targetAPIResource, userAgent)
				crudTester, err := common.NewFederatedTypeCrudTester(tl, typeConfig, fedConfig, kubeConfig, testClusters, framework.PollInterval, framework.SingleCallTimeout)
				if err != nil {
					tl.Fatalf("Error creating crudtester for %q: %v", templateKind, err)
				}

				clusterNames := []string{}
				for name, _ := range testClusters {
					clusterNames = append(clusterNames, name)
				}
				template, placement, override, err := common.NewTestObjects(typeConfig, f.TestNamespaceName(), clusterNames)
				if err != nil {
					tl.Fatalf("Error creating test objects: %v", err)
				}

				crudTester.CheckLifecycle(template, placement, override)
			})
		})
	}
})
