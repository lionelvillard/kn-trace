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

package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"k8s.io/client-go/rest"

	kubeproxy "k8s.io/kubectl/pkg/proxy"
)

type Proxy struct {
	handler http.Handler
}

// New creates a Kubernetes service proxy
func New(cfg *rest.Config) (Proxy, error) {
	handler, err := kubeproxy.NewProxyHandler("/", nil, cfg, 0)
	if err != nil {
		return Proxy{}, err
	}

	return Proxy{handler: handler}, nil
}

func (p Proxy) Get(name, namespace, path string) (string, error) {
	target := makeURL(name, namespace, path)
	req := httptest.NewRequest("GET", target, nil)
	responseRecorder := httptest.NewRecorder()

	p.handler.ServeHTTP(responseRecorder, req)
	body := responseRecorder.Body.String()

	if responseRecorder.Code != http.StatusOK {
		return "", errors.New(body)
	}

	return body, nil

}

func makeURL(name, namespace, path string) string {
	// http://kubernetes_master_address/api/v1/namespaces/namespace_name/services/[https:]service_name[:port_name]/proxy

	sep := ""
	if len(path) > 0 && !strings.HasPrefix(path, "/") {
		sep = "/"
	}
	return fmt.Sprintf("/api/v1/namespaces/%s/services/%s/proxy%s%s", namespace, name, sep, path)
}
