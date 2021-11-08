// Copyright © 2021 The Knative Authors
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

package config

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"knative.dev/pkg/tracing/config"
)

func Load(ctx context.Context, client kubernetes.Interface) (*config.Config, error) {
	cm, err := client.CoreV1().ConfigMaps("knative-eventing").Get(ctx, "config-tracing", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return config.NewTracingConfigFromConfigMap(cm)
}

// Validate the given configuration is compatible with kn trace
func Validate(cfg *config.Config) error {
	if cfg.Backend != "zipkin" {
		return errors.New("incompatible tracing configuration. Only zipkin backend is currently supported")
	}

	if cfg.ZipkinEndpoint == "" {
		return errors.New("missing Zipkin endpoint")
	}

	// TODO: otel support.
	// ie.  zipkin-endpoint: http://otel-collector.observability:9411/api/v2/spans

	return nil
}
