module knative.dev/kn-plugin-trace

go 1.15

require (
	github.com/fatih/color v1.7.0
	github.com/openzipkin/zipkin-go v0.3.0
	github.com/spf13/cobra v1.2.1
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/kubectl v0.22.3
	knative.dev/client v0.27.1-0.20211104101401-4fb6bdb95a9c
	knative.dev/hack v0.0.0-20211104075903-0f69979bbb7d
	knative.dev/pkg v0.0.0-20211104101302-51b9e7f161b4
)
