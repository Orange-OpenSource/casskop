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

	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//GetPodDisruptionBudget return the PodDisruptionBudget name from the cluster in the namespace
func (rcc *ReconcileCassandraCluster) GetPodDisruptionBudget(namespace,
	name string) (*policyv1beta1.PodDisruptionBudget, error) {

	pdb := &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return pdb, rcc.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, pdb)
}

//CreatePodDisruptionBudget create a new PodDisruptionBudget pdb
func (rcc *ReconcileCassandraCluster) CreatePodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) error {
	err := rcc.Client.Create(context.TODO(), pdb)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("PodDisruptionBudget already exists: %cc", err)
		}
		return fmt.Errorf("failed to create cassandra PodDisruptionBudget: %cc", err)
	}
	rcc.cc.Labels["PDB-UID"] = string(pdb.UID)
	needUpdate = true
	return nil
}

//DeletePodDisruptionBudget delete a new PodDisruptionBudget pdb
func (rcc *ReconcileCassandraCluster) DeletePodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) error {
	var err error

	if uidStr, ok := rcc.cc.Labels["PDB-UID"]; ok && uidStr != "" {
		uid := types.UID(uidStr)
		err = rcc.Client.Delete(context.TODO(), pdb, &client.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
	} else {
		err = rcc.Client.Delete(context.TODO(), pdb)
	}

	if err != nil {
		return fmt.Errorf("failed to delete cassandra PodDisruptionBudget: %cc", err)
	}
	rcc.cc.Labels["PDB-UID"] = ""
	needUpdate = true
	return nil
}

//UpdatePodDisruptionBudget updates an existing PodDisruptionBudget pdb
func (rcc *ReconcileCassandraCluster) UpdatePodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) error {
	err := rcc.Client.Update(context.TODO(), pdb)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("PodDisruptionBudget already exists: %cc", err)
		}
		return fmt.Errorf("failed to update cassandra PodDisruptionBudget: %cc", err)
	}
	return nil
}

//CreateOrUpdatePodDisruptionBudget Create PodDisruptionBudget if not existing, or update it if existing.
func (rcc *ReconcileCassandraCluster) CreateOrUpdatePodDisruptionBudget(pdb *policyv1beta1.PodDisruptionBudget) error {
	var err error
	rcc.storedPdb, err = rcc.GetPodDisruptionBudget(pdb.Namespace, pdb.Name)
	if err != nil {
		// If no resource we need to create.
		if apierrors.IsNotFound(err) {
			return rcc.CreatePodDisruptionBudget(pdb)
		}
		return err
	}

	if *rcc.storedPdb.Spec.MaxUnavailable != *pdb.Spec.MaxUnavailable {
		if err := rcc.DeletePodDisruptionBudget(pdb); err != nil {
			return err
		}
		return rcc.CreatePodDisruptionBudget(pdb)
	}
	//We can't Update PorDisruptionBudget
	return nil
	/*
		// Already exists, need to Update.
		pdb.ResourceVersion = rcc.storedPdb.ResourceVersion

		return rcc.UpdatePodDisruptionBudget(pdb)
	*/
}
