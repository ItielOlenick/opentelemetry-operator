// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package targetallocator

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/open-telemetry/opentelemetry-operator/internal/autodetect/certmanager"
	"github.com/open-telemetry/opentelemetry-operator/internal/manifests"
	"github.com/open-telemetry/opentelemetry-operator/pkg/featuregate"
)

const (
	ComponentOpenTelemetryTargetAllocator = "opentelemetry-targetallocator"
)

// Build creates the manifest for the TargetAllocator resource.
func Build(params manifests.Params) ([]client.Object, error) {
	var resourceManifests []client.Object
	if !params.OtelCol.Spec.TargetAllocator.Enabled {
		return resourceManifests, nil
	}
	resourceFactories := []manifests.K8sManifestFactory{
		manifests.Factory(ConfigMap),
		manifests.Factory(Deployment),
		manifests.FactoryWithoutError(ServiceAccount),
		manifests.FactoryWithoutError(Service),
		manifests.Factory(PodDisruptionBudget),
	}

	if params.TargetAllocator.Spec.Observability.Metrics.EnableMetrics && featuregate.PrometheusOperatorIsAvailable.IsEnabled() {
		resourceFactories = append(resourceFactories, manifests.FactoryWithoutError(ServiceMonitor))
	}

	if params.Config.CertManagerAvailability() == certmanager.Available && params.Config.EnableTargetAllocatorMTLS() {
		resourceFactories = append(resourceFactories, manifests.FactoryWithoutError(SelfSignedIssuer))
		resourceFactories = append(resourceFactories, manifests.FactoryWithoutError(CACertificate))
		resourceFactories = append(resourceFactories, manifests.FactoryWithoutError(CAIssuer))
		resourceFactories = append(resourceFactories, manifests.FactoryWithoutError(ServingCertificate))
		resourceFactories = append(resourceFactories, manifests.FactoryWithoutError(ClientCertificate))
	}

	for _, factory := range resourceFactories {
		res, err := factory(params)
		if err != nil {
			return nil, err
		} else if manifests.ObjectIsNotNil(res) {
			resourceManifests = append(resourceManifests, res)
		}
	}
	return resourceManifests, nil
}
