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

	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

func (rcc *ReconcileCassandraCluster) GetPVC(namespace, name string) (*v1.PersistentVolumeClaim, error) {

	o := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return o, rcc.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, o)
}

func (rcc *ReconcileCassandraCluster) ListPVC(namespace string,
	selector map[string]string) (*v1.PersistentVolumeClaimList, error) {

	clientOpt := &client.ListOptions{Namespace: namespace, LabelSelector: labels.SelectorFromSet(selector)}
	opt := []client.ListOption{
		clientOpt,
	}

	o := &v1.PersistentVolumeClaimList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
	}

	return o, rcc.Client.List(context.TODO(), o, opt...)
}

func (rcc *ReconcileCassandraCluster) deletePVC(pvc *v1.PersistentVolumeClaim) error {

	return rcc.Client.Delete(context.TODO(), pvc)

}
