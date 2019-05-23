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
	"errors"
	"fmt"
	"regexp"

	"context"

	"github.com/sirupsen/logrus"
	"github.com/swarvanusg/go_jolokia"
	funk "github.com/thoas/go-funk"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var localSystemKeyspaces = []string{"system", "system_schema"}

/*JolokiaURL returns the url used to connect to a Jolokia server based on a host and a port*/
func JolokiaURL(host string, port int) string {
	return fmt.Sprintf("http://%s:%d/jolokia/", host, port)
}

// JolokiaClient is a structure that exposes a host and a jolokia client
type JolokiaClient struct {
	client *go_jolokia.JolokiaClient
	host   string
}

func (jolokiaClient *JolokiaClient) executeReadRequest(jolokiaRequest *go_jolokia.JolokiaRequest) (*go_jolokia.JolokiaReadResponse, error) {
	return (*go_jolokia.JolokiaClient)(jolokiaClient.client).ExecuteReadRequest(jolokiaRequest)
}

func (jolokiaClient *JolokiaClient) executeOperation(mBean, operation string,
	arguments interface{}, pattern string) (*go_jolokia.JolokiaReadResponse, error) {
	return (*go_jolokia.JolokiaClient)(jolokiaClient.client).ExecuteOperation(mBean, operation, arguments, pattern)
}

/*NewJolokiaClient returns a new Joloka client for the host name and port provided*/
func NewJolokiaClient(host string, port int, rcc *ReconcileCassandraCluster,
	secretRef v1.LocalObjectReference, namespace string) (*JolokiaClient, error) {
	jolokiaClient := JolokiaClient{go_jolokia.NewJolokiaClient(JolokiaURL(host, port)), host}
	logrus.WithFields(logrus.Fields{"host": host, "port": port,
		"secretRef": secretRef, "namespace": namespace}).Debug("Creating Jolokia connection")
	if (secretRef != v1.LocalObjectReference{}) {
		logrus.WithFields(logrus.Fields{"host": host, "port": port,
			"secretRef": secretRef, "namespace": namespace}).Debug("Using Secret for Jolokia connection")
		secret := &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretRef.Name,
				Namespace: namespace,
			},
		}
		err := rcc.client.Get(context.TODO(), types.NamespacedName{Name: secretRef.Name, Namespace: namespace}, secret)

		if err != nil {
			logrus.WithFields(logrus.Fields{"host": host, "port": port,
				"secretRef": secretRef, "namespace": namespace}).Error("Can't get Jolokia secret")
			return nil, err
		}
		jolokiaClient.client.SetCredential(string(secret.Data["username"]), string(secret.Data["password"]))
	}
	return (*JolokiaClient)(&jolokiaClient), nil
}

func checkJolokiaErrors(resp *go_jolokia.JolokiaReadResponse, err error) (*go_jolokia.JolokiaReadResponse, error) {
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}

func (jolokiaClient *JolokiaClient) leavingNodes() ([]string, error) {
	request := go_jolokia.NewJolokiaRequest(go_jolokia.READ, "org.apache.cassandra.db:type=StorageService", nil, "LeavingNodes")
	result, err := checkJolokiaErrors(jolokiaClient.client.ExecuteReadRequest(request))
	if err != nil {
		return nil, fmt.Errorf("Cannot get list of leaving nodes: %v", err.Error())
	}
	v, isSlice := result.Value.([]interface{})
	if isSlice {
		leavingNodes := []string{}
		for _, value := range v {
			str, isString := value.(string)
			if isString {
				leavingNodes = append(leavingNodes, str)
			}
		}
		logrus.WithFields(logrus.Fields{"leavingNodes": leavingNodes}).Debug("List of leaving nodes")
		return leavingNodes, nil
	}
	return nil, fmt.Errorf("Value returned by Jolokia is not a slice: %v", result.Value)
}

func (jolokiaClient *JolokiaClient) hostIDMap() (map[string]string, error) {
	request := go_jolokia.NewJolokiaRequest(go_jolokia.READ, "org.apache.cassandra.db:type=StorageService", nil, "HostIdMap")
	result, err := checkJolokiaErrors(jolokiaClient.client.ExecuteReadRequest(request))
	if err != nil {
		return nil, fmt.Errorf("Cannot get host id map: %v", err.Error())
	}
	if m, ok := result.Value.(map[string]interface{}); ok {
		hostIDMap := map[string]string{}
		for k, v := range m {
			str, isString := v.(string)
			if isString {
				hostIDMap[k] = str
			}
		}
		logrus.WithFields(logrus.Fields{"hostIDMap": hostIDMap}).Debug("Map of hosts IPs and IDs")
		return hostIDMap, nil
	}
	return nil, fmt.Errorf("Value returned by Jolokia is not a map: %v", result.Value)
}

func (jolokiaClient *JolokiaClient) keyspaces() ([]string, error) {
	request := go_jolokia.NewJolokiaRequest(go_jolokia.READ, "org.apache.cassandra.db:type=StorageService", nil, "Keyspaces")
	result, err := checkJolokiaErrors(jolokiaClient.client.ExecuteReadRequest(request))
	if err != nil {
		return nil, fmt.Errorf("Cannot get list of keyspaces: %v", err.Error())
	}
	v, isSlice := result.Value.([]interface{})
	if isSlice {
		keyspaces := []string{}
		for _, value := range v {
			str, isString := value.(string)
			if isString {
				keyspaces = append(keyspaces, str)
			}
		}
		return keyspaces, nil
	}
	return nil, fmt.Errorf("Value returned by Jolokia is not a slice: %v", result.Value)
}

func (jolokiaClient *JolokiaClient) nonLocalKeyspaces() ([]string, error) {
	keyspaces, err := jolokiaClient.keyspaces()
	if err != nil {
		return nil, err
	}
	nonLocalKeyspaces := []string{}
	for _, keyspace := range keyspaces {
		if !funk.Contains(localSystemKeyspaces, keyspace) {
			nonLocalKeyspaces = append(nonLocalKeyspaces, keyspace)
		}
	}
	return nonLocalKeyspaces, nil
}

/*NodeCleanup triggers a cleanup of all keyspaces on the pod using a jolokia client and return the index of the last keyspace accessed and any error*/
func (jolokiaClient *JolokiaClient) NodeCleanup() error {
	keyspaces, err := jolokiaClient.nonLocalKeyspaces()
	if err != nil {
		return err
	}
	return jolokiaClient.NodeCleanupKeyspaces(keyspaces)
}

/*NodeCleanupKeyspaces triggers a cleanup of each keyspaces on the pod using a jolokia client and returns the index of the last keyspace accessed and any error*/
func (jolokiaClient *JolokiaClient) NodeCleanupKeyspaces(keyspaces []string) error {
	for _, keyspace := range keyspaces {
		logrus.Infof("[%s]: Cleanup of keyspace %s", jolokiaClient.host, keyspace)
		_, err := checkJolokiaErrors(jolokiaClient.executeOperation("org.apache.cassandra.db:type=StorageService",
			"forceKeyspaceCleanup(java.lang.String,[Ljava.lang.String;)",
			[]interface{}{keyspace, []string{}}, ""))
		if err != nil {
			logrus.Errorf("Cleanup of keyspace %s failed: %v", keyspace, err.Error())
			return err
		}
	}
	return nil
}

/*NodeUpgradeSSTables triggers an upgradeSSTables of each keyspaces through a jolokia client and returns any error*/
func (jolokiaClient *JolokiaClient) NodeUpgradeSSTables(threads int) error {
	keyspaces, err := jolokiaClient.keyspaces()
	if err != nil {
		return err
	}
	return jolokiaClient.NodeUpgradeSSTablesKeyspaces(keyspaces, threads)
}

/*NodeUpgradeSSTablesKeyspaces triggers an upgradeSSTables for a list of keyspaces through a jolokia connection and returns any error*/
func (jolokiaClient *JolokiaClient) NodeUpgradeSSTablesKeyspaces(keyspaces []string,
	threads int) error {
	for _, keyspace := range keyspaces {
		logrus.Infof("[%s]: Upgrade SSTables of keyspace %s", jolokiaClient.host, keyspace)
		_, err := checkJolokiaErrors(jolokiaClient.executeOperation("org.apache.cassandra.db:type=StorageService",
			"upgradeSSTables(java.lang.String,boolean,int,[Ljava.lang.String;)",
			[]interface{}{keyspace, true, threads, []string{}}, ""))
		if err != nil {
			logrus.Errorf("Upgrade SSTables of keyspace %s failed: %v", keyspace, err.Error())
			return err
		}
	}
	return nil
}

/*NodeRebuild triggers a rebuild of all keyspaces on the pod using a jolokia client and returns any error*/
func (jolokiaClient *JolokiaClient) NodeRebuild(dc string) error {
	_, err := checkJolokiaErrors(jolokiaClient.executeOperation("org.apache.cassandra.db:type=StorageService",
		"rebuild(java.lang.String)",
		[]interface{}{dc}, ""))
	if err != nil {
		return fmt.Errorf("Cannot rebuild from %s: %v", dc, err.Error())
	}
	return nil
}

/*NodeDecommision decommissions a node using a jolokia client and returns any error*/
func (jolokiaClient *JolokiaClient) NodeDecommision() error {
	_, err := checkJolokiaErrors(jolokiaClient.executeOperation("org.apache.cassandra.db:type=StorageService",
		"decommission", []interface{}{}, ""))
	if err != nil {
		return fmt.Errorf("Cannot decommission: %v", err.Error())
	}
	return nil
}

/*NodeRemove remove node hostid on the pod using a jolokia client and returns any error*/
func (jolokiaClient *JolokiaClient) NodeRemove(hostid string) error {
	_, err := checkJolokiaErrors(jolokiaClient.executeOperation("org.apache.cassandra.db:type=StorageService",
		"removeNode",
		[]interface{}{hostid}, ""))

	if err != nil {
		return fmt.Errorf("Cannot remove node %s: %v", hostid, err.Error())
	}

	return nil
}

/*NodeOperationMode returns OperationMode of a node using a jolokia client and returns any error*/
func (jolokiaClient *JolokiaClient) NodeOperationMode() (string, error) {
	request := go_jolokia.NewJolokiaRequest(go_jolokia.READ, "org.apache.cassandra.db:type=StorageService", nil, "OperationMode")
	result, err := checkJolokiaErrors(jolokiaClient.executeReadRequest(request))
	if err != nil {
		return "", fmt.Errorf("Cannot get OperationMode: %v", err.Error())
	}
	v, _ := result.Value.(string)
	return v, nil
}

func (jolokiaClient *JolokiaClient) hasStreamingSessions() (bool, error) {
	request := go_jolokia.NewJolokiaRequest(go_jolokia.READ, "org.apache.cassandra.net:type=StreamManager", nil, "CurrentStreams")
	result, err := checkJolokiaErrors(jolokiaClient.executeReadRequest(request))
	if err != nil {
		return true, fmt.Errorf("Cannot get list of current streams: %v", err.Error())
	}
	val, _ := result.Value.([]interface{})
	return len(val) > 0, nil
}

func (jolokiaClient *JolokiaClient) hasCompactions(name string) (bool, error) {
	request := go_jolokia.NewJolokiaRequest(go_jolokia.READ, "org.apache.cassandra.db:type=CompactionManager", nil, "Compactions")
	result, err := checkJolokiaErrors(jolokiaClient.executeReadRequest(request))
	if err != nil {
		logrus.Error(err.Error())
		return true, fmt.Errorf("Cannot get list of current compactions: %v", err.Error())
	}
	compactions, _ := result.Value.([]interface{})
	for _, compaction := range compactions {
		c := compaction.(map[string]interface{})
		if c["taskType"] == name {
			return true, nil
		}
	}
	return false, nil
}

func (jolokiaClient *JolokiaClient) hasCleanupCompactions() (bool, error) {
	return jolokiaClient.hasCompactions("Cleanup")
}

func (jolokiaClient *JolokiaClient) hasUpgradeSSTablesCompactions() (bool, error) {
	return jolokiaClient.hasCompactions("Upgrade sstables")
}

func (jolokiaClient *JolokiaClient) hasLeavingNodes() (bool, error) {
	leavingNodes, err := jolokiaClient.leavingNodes()
	if err != nil {
		return false, err
	}
	return len(leavingNodes) > 0, nil
}

/*HasDataInDC checks partition ranges of all non local keyspaces and ensure no data is replicated to the chosen datacenter*/
func (jolokiaClient *JolokiaClient) HasDataInDC(dc string) ([]string, error) {
	keyspaces, err := jolokiaClient.nonLocalKeyspaces()
	keyspacesWithDataInDC := []string{}
	if err != nil {
		return nil, err
	}
	for _, keyspace := range keyspaces {
		dataFound, err := jolokiaClient.hasKeyspaceDataInDC(keyspace, dc)
		// Returns if there is an error
		if err != nil {
			return nil, err
		}
		if dataFound {
			keyspacesWithDataInDC = append(keyspacesWithDataInDC, keyspace)
		}
	}
	return keyspacesWithDataInDC, nil
}

func (jolokiaClient *JolokiaClient) hasKeyspaceDataInDC(keyspace, dc string) (bool, error) {
	result, err := checkJolokiaErrors(jolokiaClient.executeOperation("org.apache.cassandra.db:type=StorageService",
		"describeRingJMX", []interface{}{keyspace}, ""))
	if err != nil {
		return false, fmt.Errorf("Cannot describe ring using keyspace %s: %v", keyspace, err.Error())
	}
	regexDc := regexp.MustCompile(fmt.Sprintf("datacenter:%s", dc))
	tokenRanges, _ := result.Value.([]interface{})
	for _, tokenRange := range tokenRanges {
		// Returns true as soon as we find one token range that is replicated to the chosen datacenter
		if regexDc.MatchString(tokenRange.(string)) {
			return true, nil
		}
	}
	return false, nil
}
