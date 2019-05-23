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

package cassandracluster

import (
	"context"
	"fmt"

	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"

	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	cassandraPort     = 9042
	cassandraPortName = "cassandra-port"

	cassandraThrift     = 9160
	cassandraThriftName = "cassandra-thrift"
)

const (
	exporterPort                 = 9121
	exporterPortName             = "http-metrics"
	exporterCassandraJmxPort     = 1234
	exporterCassandraJmxPortName = "http-promcassjmx"
)

func (rcc *ReconcileCassandraCluster) ensureCassandraService(cc *api.CassandraCluster, dcName, rackName string) error {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	selector := k8s.LabelsForCassandraDCRack(cc, dcName, rackName)
	svc := generateCassandraService(cc, dcRackName, selector, nil)

	k8s.AddOwnerRefToObject(svc, k8s.AsOwner(cc))
	err := rcc.client.Create(context.TODO(), svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cassandra service (%v)", err)
	}
	return nil
}

func (rcc *ReconcileCassandraCluster) ensureCassandraDCService(cc *api.CassandraCluster, dcName string) error {
	selector := k8s.LabelsForCassandraDC(cc, dcName)
	svc := generateCassandraService(cc, dcName, selector, nil)

	k8s.AddOwnerRefToObject(svc, k8s.AsOwner(cc))
	err := rcc.client.Create(context.TODO(), svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cassandra DC service (%v)", err)
	}
	return nil
}

func (rcc *ReconcileCassandraCluster) ensureCassandraServiceMonitoring(cc *api.CassandraCluster,
	dcName string) error {
	selector := k8s.LabelsForCassandraDC(cc, dcName)
	svc := generateCassandraExporterService(cc, dcName, selector, nil)

	k8s.AddOwnerRefToObject(svc, k8s.AsOwner(cc))
	err := rcc.client.Create(context.TODO(), svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cassandra service Monitoring: %v", err)
	}
	return nil
}

// ensureCassandraPodDisruptionBudget generate and apply the PodDisruptionBudget
// take dcName to accordingly named the pdb, and target the pods
func (rcc *ReconcileCassandraCluster) ensureCassandraPodDisruptionBudget(cc *api.CassandraCluster) error {
	labels := k8s.LabelsForCassandra(cc)

	pdb := generatePodDisruptionBudget(cc.Name, cc.Namespace, labels, k8s.AsOwner(cc),
		intstr.FromInt(int(cc.Spec.MaxPodUnavailable)))
	err := rcc.CreateOrUpdatePodDisruptionBudget(pdb)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Errorf("CreateOrUpdatePodDisruptionBudget Error: %v", err)
	}
	return err
}

// ensureCassandraStatefulSet generate and apply the statefulset
// take dcRackName to accordingly named the statefulset
// take dc and rack index of dc and rack in conf to retrieve according  nodeselectors labels
func (rcc *ReconcileCassandraCluster) ensureCassandraStatefulSet(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus, dcRackName string, dc int, rack int) error {

	labels, nodeSelector := k8s.GetDCRackLabelsAndNodeSelectorForStatefulSet(cc, dc, rack)

	ss := generateCassandraStatefulSet(cc, status, dcRackName, labels, nodeSelector, nil)
	k8s.AddOwnerRefToObject(ss, k8s.AsOwner(cc))

	err := rcc.CreateOrUpdateStatefulSet(ss, status, dcRackName)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cassandra statefulset: %v", err)
	}

	return nil
}
