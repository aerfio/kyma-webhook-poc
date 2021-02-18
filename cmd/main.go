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
	"encoding/json"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	controller_zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/aerfio/kyma-webhook-poc/pkg"
)

type Config struct {
	Port                  int    `envconfig:"default=8443"`
	MetricsAddress        string `envconfig:"default=:8080"`
	CertDir               string `envconfig:"default=/var/run/webhook"`
	ValidatingWebhookPath string `envconfig:"default=/validating"`
	LogLevel              string `envconfig:"default=info"`
	StackTraceLevel       string `envconfig:"default=warn"`
	LogFormat             string `envconfig:"default=json"`
}

func main() {
	setupLog := ctrl.Log.WithName("setup")
	var cfg Config
	if err := envconfig.InitWithPrefix(&cfg, "APP"); err != nil {
		setupLog.Error(err, "unable to parse config")
		os.Exit(1)
	}

	setupLoggingOrDie(cfg, setupLog)

	if lg := ctrl.Log.V(1); lg.Enabled() {
		marshalledConfig, err := json.Marshal(cfg)
		if err != nil {
			setupLog.Error(err, "unable to marshal config to json")
			os.Exit(1)
		}
		lg.WithName("setup.config").Info(string(marshalledConfig))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		MetricsBindAddress: cfg.MetricsAddress,
		Port:               cfg.Port,
		CertDir:            cfg.CertDir,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	hookServer := mgr.GetWebhookServer()
	setupLog.Info("registering webhooks to the webhook server")
	hookServer.Register(cfg.ValidatingWebhookPath, &webhook.Admission{Handler: &pkg.Validator{Log: ctrl.Log.WithName("webhook")}})

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupLoggingOrDie(cfg Config, setupLog logr.Logger) {
	logLevel, err := logger.MapLevel(cfg.LogLevel)
	if err != nil {
		setupLog.Error(err, "unable to map supplied log level to interally used ones")
		os.Exit(1)
	}

	logLevelZap, err := logLevel.ToZapLevel()
	if err != nil {
		setupLog.Error(err, "unable to map log level to zap level")
		os.Exit(1)
	}

	stackTraceLevel, err := logger.MapLevel(cfg.StackTraceLevel)
	if err != nil {
		setupLog.Error(err, "unable to map supplied stacktrace level to interally used ones")
		os.Exit(1)
	}

	stackTraceLevelZap, err := stackTraceLevel.ToZapLevel()
	if err != nil {
		setupLog.Error(err, "unable to map stacktrace level to zap level")
		os.Exit(1)
	}

	format, err := logger.MapFormat(cfg.LogFormat)
	if err != nil {
		setupLog.Error(err, "unable to recognise log format")
		os.Exit(1)
	}
	encoder, err := format.ToZapEncoder()
	if err != nil {
		setupLog.Error(err, "unable to create zap encoder")
		os.Exit(1)
	}

logger.New(format, logLevel, )

	ctrl.SetLogger(controller_zap.New(
		controller_zap.Level(&logLevelZap),
		controller_zap.StacktraceLevel(&stackTraceLevelZap),
		controller_zap.Encoder(encoder),
	))
}
