/*
Copyright 2018 The Multicluster-Controller Authors.

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
	"flag"
	"github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/controller/multi-casskop/models"
	"log"
	"os"
	"strings"

	"k8s.io/client-go/rest"

	"admiralty.io/multicluster-controller/pkg/cluster"
	"admiralty.io/multicluster-controller/pkg/manager"
	"admiralty.io/multicluster-service-account/pkg/config"
	mc "github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/controller/multi-casskop"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/sample-controller/pkg/signals"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	logLevelEnvVar = "LOG_LEVEL"
)

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

func main() {

	logType, found := os.LookupEnv("LOG_TYPE")
	if found && logType == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetLevel(getLogLevel())

	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalf("Usage: MultiCasskop cluster-1 cluster-2 .. cluster-n")
	}

	var clusters models.Clusters

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("failed to get watch namespace: %v", err)
	}

	for i := 0; i < flag.NArg(); i++ {
		clusterName := flag.Arg(i)
		var cfg *rest.Config
		var err error

		if i == 0 {
			logrus.Infof("Configuring Client %d for local cluster %s (first in arg list). using local k8s api access",
				i+1, clusterName)
			cfg, err = kconfig.GetConfig()
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
			// Set up master cluster
			clusters.Master = models.Cluster{Name: clusterName, Cluster: cluster.New(clusterName, cfg,
				cluster.Options{CacheOptions: cluster.CacheOptions{Namespace: namespace}})}
		} else {
			logrus.Infof("Configuring Client %d for distant cluster %s. using imported secret of same name", i+1,
				clusterName)
			cfg, _, err = config.NamedConfigAndNamespace(clusterName)
			if err != nil {
				log.Fatal(err)
			}

			// Set up Remotes clusters
			clusters.Remotes = append(clusters.Remotes,
				models.Cluster{Name: clusterName, Cluster: cluster.New(clusterName, cfg,
					cluster.Options{CacheOptions: cluster.CacheOptions{Namespace: namespace}})})
		}

	}

	logrus.Info("Creating Controller")
	co, err := mc.NewController(clusters, namespace)
	if err != nil {
		log.Fatalf("creating Cassandra Multi Cluster controller: %v", err)
	}

	m := manager.New()
	m.AddController(co)

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

	logrus.Info("Starting Manager.")
	if err := m.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatalf("while or after starting manager: %v", err)
	}
}
