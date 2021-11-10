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

package setup

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"knative.dev/client/pkg/kn/commands"
)

const (
	KnToolsNamespace = "kntools"
)

// Zipkin sets up Zipkin and change tracing configuration accordingly
func Zipkin(ctx context.Context, p *commands.KnParams) error {
	cfg, err := p.RestConfig()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// TODO: alternate eventing installation namespace
	cm, err := client.CoreV1().ConfigMaps("knative-eventing").Get(ctx, "config-tracing", metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		// knative eventing hasn't been installed properly.
		fmt.Println("⚠️ missing config-tracing in the knative-eventing namespace which is an indicator that Knative Eventing hasn't been properly installed. Recovering.")
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "config-tracing",
			},
			Data: map[string]string{},
		}
	}

	updated := false

	backend, ok := cm.Data["backend"]
	if !ok {
		cm.Data["backend"] = "zipkin"
		updated = true
	} else if backend != "zipkin" {
		return fmt.Errorf("incompatible tracing configuration: unsupported %s backend", backend)
	}

	endpoint, ok := cm.Data["zipkin-endpoint"]
	if !ok {
		endpoint, err = installZipkin(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to install Zipkin: %w", err)
		}

		cm.Data["zipkin-endpoint"] = endpoint
		updated = true
	} else {
		// TODO: Check endpoint is a real zipkin
		// opentelemetry support receiving zipkin span but does not support queries

	}

	debug, ok := cm.Data["debug"]
	if !ok || debug != "true" {
		cm.Data["debug"] = "true"
		updated = true
	}

	if updated {
		_, err = client.CoreV1().ConfigMaps("knative-eventing").Update(ctx, cm, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		fmt.Println("tracing configuration successfully created")
	} else {
		fmt.Println("tracing configuration unchanged")
	}

	return nil
}

func installZipkin(ctx context.Context, client kubernetes.Interface) (string, error) {
	// Check if already installed

	// First: check kntools namespaces
	_, err := client.CoreV1().Namespaces().Get(ctx, KnToolsNamespace, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}
		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: KnToolsNamespace,
			},
		}
		_, err := client.CoreV1().Namespaces().Create(ctx, &ns, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
	}

	// Then: check zipkin is there and ready
	_, err = client.CoreV1().Endpoints(KnToolsNamespace).Get(ctx, "zipkin", metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}

		// Something missing so installing...
		labels := map[string]string{
			"app": "zipkin",
		}

		d := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "zipkin",
				Labels: labels,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labels},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "zipkin", Image: "openzipkin/zipkin"},
						},
					},
				},
			},
		}

		_, err = client.AppsV1().Deployments(KnToolsNamespace).Create(ctx, &d, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}

		s := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "zipkin",
				Labels: labels,
			},

			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: labels,
				Ports: []corev1.ServicePort{
					{Protocol: "TCP", Port: 9411, TargetPort: intstr.FromInt(9411)},
				},
			},
		}

		_, err = client.CoreV1().Services(KnToolsNamespace).Create(ctx, &s, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}

		// TODO: wait for endpoint to be ready
	}

	return "http://zipkin." + KnToolsNamespace + ".svc.cluster.local:9411/api/v2/spans", nil
}
