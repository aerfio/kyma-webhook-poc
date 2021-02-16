/*


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

package main

import (
	"os"

	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/aerfio/kyma-webhook-poc/pkg"
	// +kubebuilder:scaffold:imports

)

type Config struct {
	DebugMode             bool   `envconfig:"default=false"`
	Port                  int    `envconfig:"default=8443"`
	MetricsAddress        string `envconfig:"default=:8080"`
	CertDir               string `envconfig:"default=/var/run/webhook"`
	ValidatingWebhookPath string `envconfig:"default=/pod-validating"`
}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func main() {
	var cfg Config
	if err := envconfig.InitWithPrefix(&cfg, "APP"); err != nil {
		setupLog.Error(err, "unable to parse config")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(cfg.DebugMode)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		MetricsBindAddress: cfg.MetricsAddress,
		Port:               cfg.Port,
		CertDir:            cfg.CertDir,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	decoder, err := admission.NewDecoder(scheme)
	if err != nil {
		setupLog.Error(err, "unable to create scheme decoder")
		os.Exit(1)
	}
	hookServer := mgr.GetWebhookServer()
	setupLog.Info("registering webhooks to the webhook server")
	hookServer.Register(cfg.ValidatingWebhookPath, &webhook.Admission{Handler: &pkg.PodValidator{Client: mgr.GetClient(), Decoder: decoder}})

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
