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
	api "github.com/Orange-OpenSource/casskop/api/v2"
	policyv1beta1 "k8s.io/api/policy/v1beta1"

	"github.com/Orange-OpenSource/casskop/pkg/k8s"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	//max 15 char for port names
	cassandraPort                = 9042
	cassandraPortName            = "cql"
	cassandraIntraNodePort       = 7000
	cassandraIntraNodeName       = "intra-node"
	cassandraIntraNodeTLSPort    = 7001
	cassandraIntraNodeTLSName    = "intra-node-tls"
	cassandraJMX                 = 7199 //used for nodetool+istio
	cassandraJMXName             = "jmx-port"
	JolokiaPort                  = 8778
	JolokiaPortName              = "jolokia"
	exporterCassandraJmxPort     = 9500
	exporterCassandraJmxPortName = "promjmx"
)

func (rcc *CassandraClusterReconciler) ensureCassandraService(cc *api.CassandraCluster) error {
	selector := k8s.LabelsForCassandra(cc)
	svc := generateCassandraService(cc, selector, nil)

	k8s.AddOwnerRefToObject(svc, k8s.AsOwner(cc))
	err := rcc.Client.Create(context.TODO(), svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cassandra service (%v)", err)
	}
	return nil
}

func (rcc *CassandraClusterReconciler) ensureCassandraServiceMonitoring(cc *api.CassandraCluster,
	dcName string) error {
	selector := k8s.LabelsForCassandra(cc)
	svc := generateCassandraExporterService(cc, selector, nil)

	k8s.AddOwnerRefToObject(svc, k8s.AsOwner(cc))
	err := rcc.Client.Create(context.TODO(), svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cassandra service Monitoring: %v", err)
	}
	return nil
}

// ensureCassandraPodDisruptionBudget generate and apply the PodDisruptionBudget
// take dcName to accordingly named the pdb, and target the pods
func (rcc *CassandraClusterReconciler) ensureCassandraPodDisruptionBudget(cc *api.CassandraCluster) error {
	pdb := rcc.podDisruptionBudgetEnvelope(cc)
	err := rcc.CreateOrUpdatePodDisruptionBudget(pdb)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Errorf("CreateOrUpdatePodDisruptionBudget Error: %v", err)
	}
	return err
}

func (rcc *CassandraClusterReconciler) podDisruptionBudgetEnvelope(cc *api.CassandraCluster) *policyv1beta1.PodDisruptionBudget {
	labels := k8s.LabelsForCassandra(cc)
	return generatePodDisruptionBudget(cc.Name, cc.Namespace, labels, k8s.AsOwner(cc),
		intstr.FromInt(int(cc.Spec.MaxPodUnavailable)))
}

// ensureCassandraStatefulSet generate and apply the statefulset
// take dcRackName to accordingly named the statefulset
// take dc and rack index of dc and rack in conf to retrieve according  nodeselectors labels
func (rcc *CassandraClusterReconciler) ensureCassandraStatefulSet(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus, dcName string, dcRackName string, dc int, rack int) (bool, error) {

	labels, nodeSelector := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(cc, dc, rack)

	ss, err := generateCassandraStatefulSet(cc, status, dcName, dcRackName, labels, nodeSelector, nil)
	if err != nil {
		return true, err
	}
	k8s.AddOwnerRefToObject(ss, k8s.AsOwner(cc))

	breakResyncloop, err := rcc.CreateOrUpdateStatefulSet(ss, status, dcRackName)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return breakResyncloop, fmt.Errorf("failed to create cassandra statefulset: %v", err)
	}

	return breakResyncloop, nil
}
