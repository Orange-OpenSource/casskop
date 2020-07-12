// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/Orange-OpenSource/casskop/pkg/apis"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/controller"
	"github.com/Orange-OpenSource/casskop/pkg/controller/cassandracluster"
	"github.com/Orange-OpenSource/casskop/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/sirupsen/logrus"
	"github.com/zput/zxcTool/ztLog/zt_formatter"
	v1 "k8s.io/api/core/v1"
	prometheusMetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)

const (
	logLevelEnvVar     = "LOG_LEVEL"
	resyncPeriodEnvVar = "RESYNC_PERIOD"
)

//to be set by compilator with -ldflags "-X main.compileDate=`date -u +.%Y%m%d.%H%M%S`"
var compileDate string

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("casskop Version: %v", version.Version)
	logrus.Infof("casskop Compilation Date: %s", compileDate)
	logrus.Infof("casskop LogLevel: %v", getLogLevel())
	logrus.Infof("casskop ResyncPeriod: %v", getResyncPeriod())
}

func getLogLevel() logrus.Level {
	logLevel, found := os.LookupEnv(logLevelEnvVar)
	if !found {
		return logrus.InfoLevel
	}
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return logrus.DebugLevel
	case "INFO":
		return logrus.InfoLevel
	case "ERROR":
		return logrus.ErrorLevel
	case "WARN":
		return logrus.WarnLevel
	}
	return logrus.InfoLevel
}

func getResyncPeriod() int {
	var resyncPeriod int
	var err error
	resync, found := os.LookupEnv(resyncPeriodEnvVar)
	if !found {
		resyncPeriod = api.DefaultResyncPeriod
	} else {
		resyncPeriod, err = strconv.Atoi(resync)
		if err != nil {
			resyncPeriod = api.DefaultResyncPeriod
		}
	}
	return resyncPeriod
}

func main() {
	logLevel := getLogLevel()
	logrus.SetLevel(logLevel)
	if logLevel == logrus.DebugLevel {
		ztFormatter := &zt_formatter.ZtFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		}
		logrus.SetReportCaller(true)
		logrus.SetFormatter(ztFormatter)
	}
	if logType, _ := os.LookupEnv("LOG_TYPE"); logType == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "casskop-lock")
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// NewFileReady returns a Ready that uses the presence of a file on disk to
	// communicate whether the operator is ready. The operator's Pod definition
	// "stat /tmp/operator-sdk-ready".
	logrus.Info("Writing ready file.")
	r := ready.NewFileReady()
	err = r.Set()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	defer r.Unset()

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	logrus.Info("Registering Components.")

	prometheusMetrics.Registry.MustRegister(cassandracluster.ClusterActionMetric, cassandracluster.ClusterPhaseMetric)

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Create Service object to expose the metrics port.
	servicePorts := []v1.ServicePort{
		{Name: metrics.OperatorPortName, Port: metricsPort,
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
	}

	if _, err := metrics.CreateMetricsService(ctx, cfg, servicePorts); err != nil {
		logrus.Info(err.Error())
	}

	logrus.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logrus.Error(err, "manager exited non-zero")
		os.Exit(1)
	}
}
