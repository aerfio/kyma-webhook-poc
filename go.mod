module github.com/aerfio/kyma-webhook-poc

go 1.13

require (
	github.com/go-logr/logr v0.3.0
	github.com/kyma-project/kyma/common v0.0.0-20210218112830-67920ad22635
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.16.0
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.2
)

replace github.com/kyma-project/kyma/common => /Users/i354746/go/src/github.com/kyma-project/kyma/common
