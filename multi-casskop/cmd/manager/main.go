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
	"log"
	"os"

	"admiralty.io/multicluster-controller/pkg/cluster"
	"admiralty.io/multicluster-controller/pkg/manager"
	"admiralty.io/multicluster-service-account/pkg/config"
	mc "github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/controller/multi-casskop"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/sample-controller/pkg/signals"
)

const (
	LogLevelEnvVar = "LOG_LEVEL"
)

func getLogLevel() logrus.Level {
	logLevel, found := os.LookupEnv(LogLevelEnvVar)
	if !found {
		return logrus.InfoLevel
	}
	switch logLevel {
	case "Debug":
		return logrus.DebugLevel
	case "Info":
		return logrus.InfoLevel
	case "Error":
		return logrus.ErrorLevel
	case "Warn":
		return logrus.WarnLevel
	}
	return logrus.InfoLevel
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalf("Usage: MultiCasskop cluster-1 cluster-2 .. cluster-n")
	}

	logType, found := os.LookupEnv("LOG_TYPE")
	if found && logType == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetLevel(getLogLevel())

	var clusters []mc.Clusters

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("failed to get watch namespace: %v", err)
	}

	for i := 0; i < flag.NArg(); i++ {
		clusterName := flag.Arg(i)
		logrus.Infof("Configuring Client %d for cluster %s.", i+1, clusterName)
		cfg, _, err := config.NamedConfigAndNamespace(clusterName)
		if err != nil {
			log.Fatal(err)
		}
		clusters = append(clusters,
			mc.Clusters{Name: clusterName, Cluster: cluster.New(clusterName, cfg,
				cluster.Options{CacheOptions: cluster.CacheOptions{Namespace: namespace}})})
	}

	logrus.Info("Creating Controller")
	co, err := mc.NewController(clusters, namespace)
	if err != nil {
		log.Fatalf("creating Cassandra Multi Cluster controller: %v", err)
	}

	m := manager.New()
	m.AddController(co)

	logrus.Info("Starting Manager.")
	if err := m.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatalf("while or after starting manager: %v", err)
	}
}
