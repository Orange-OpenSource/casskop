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

package models

import (
	"admiralty.io/multicluster-controller/pkg/cluster"
	"fmt"
	apicmc "github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/apis"
	apicc "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cluster defines k8s cluster information we need.
type Cluster struct {
	Name string
	Cluster *cluster.Cluster
}

// Clusters defines each kubernetes cluster we want to connect on it.
type Clusters struct {
	Local *Cluster
	Remotes []*Cluster
}

// Apply to a clusters struct, it will proceed to client set up associated to each
// k8s cluster.
func (clusters *Clusters) SetUpClients() (*Clients, error){
	var clients *Clients

	// Init local
	logrus.Infof("Create Client for local Cluster %s", clusters.Local.Name)
	client, err := clusters.Local.SetUpClient()
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	local := &Client{clusters.Local.Name, client}

	// Init remotes
	var remotes []*Client
	for i, cluster := range clusters.Remotes {
		logrus.Infof("Create Client %d for remote Cluster %s", i+1, cluster.Name)
		client, err := cluster.SetUpClient()
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
		remotes = append(remotes, &Client{cluster.Name, client})
	}

	// Set up clients
	clients = &Clients{Local: local, Remotes: remotes}

	return clients, nil
}

// Apply to a cluster, it will proceed to the client setup :
//   - Get delegating client
//   - Add to scheme : CassandraCluster & CassandraMultiCluster
func (cluster *Cluster) SetUpClient() (*client.DelegatingClient, error) {
	client, err := cluster.Cluster.GetDelegatingClient()
	if err != nil {
		return nil, fmt.Errorf("getting delegating client for Cluster %s Cluster: %v", cluster.Name,
			err)
	}

	logrus.Infof("Add CRDs to Cluster %s Scheme", cluster.Name)
	if err := apicc.AddToScheme(cluster.Cluster.GetScheme()); err != nil {
		return nil, fmt.Errorf("adding APIs CassandraCluster to Cluster %s Cluster's scheme: %v", cluster.Name, err)
	}
	if err := apicmc.AddToScheme(cluster.Cluster.GetScheme()); err != nil {
		return nil, fmt.Errorf("adding APIs MultiCasskop to Cluster %s Cluster's scheme: %v", cluster.Name,
			err)
	}
	return client, nil
}
