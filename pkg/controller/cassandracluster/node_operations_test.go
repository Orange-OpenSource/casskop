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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	v1 "k8s.io/api/core/v1"
)

var allKeyspaces = []string{"system", "system_auth", "system_schema", "demo1", "demo2"}

const (
	host                   = "cassandra-0.cassandra.cassie1"
	port                   = 8778
	KeyspacesJolokiaQueryP = `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
					       "attribute": "Keyspaces",
					       "type": "read"},
				   "value": [%v],
                                   "timestamp": 1528850319,
                                   "status": 200}`
)

/* The wrapper structure for an EXEC json request */
type execRequestData struct {
	Type      string        `json:"type"`
	Mbean     string        `json:"mbean"`
	Attribute string        `json:"attribute"`
	Arguments []interface{} `json:"arguments"`
}

func keyspaceListString() string {
	return fmt.Sprintf(KeyspacesJolokiaQueryP,
		`"`+strings.Join(allKeyspaces, `","`)+`"`)
}
func TestJolokiaURL(t *testing.T) {
	jolokiaURL := JolokiaURL(host, port)
	if jolokiaURL != fmt.Sprintf("http://%s:%d/jolokia/", host, port) {
		t.Errorf("Malformed jolokia_url")
	}
}

func TestNodeCleanupKeyspace(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		httpmock.NewStringResponder(200,
			`{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
                          "arguments": ["demo", []],
                          "type": "exec",
                          "operation": "forceKeyspaceCleanup(java.lang.String,[Ljava.lang.String;)"},
              "value": 0,
              "timestamp": 1528848808,
	      "status": 200}`))
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	err := jolokiaClient.NodeCleanupKeyspaces([]string{"demo"})
	if err != nil {
		t.Errorf("NodeCleanupKeyspace failed with : %s", err)
	}
}
func TestNodeCleanup(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	keyspacescleaned := []string{}

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			var execrequestdata execRequestData
			if err := json.NewDecoder(req.Body).Decode(&execrequestdata); err != nil {
				t.Error("Can't decode request received")
			}
			if execrequestdata.Attribute == "Keyspaces" {
				return httpmock.NewStringResponse(200, keyspaceListString()), nil
			}
			keyspace, ok := execrequestdata.Arguments[0].(string)

			if !ok {
				t.Error("Keyspace can't be nil")
			}

			keyspacescleaned = append(keyspacescleaned, keyspace)

			response := `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
						  "arguments": ["%s", []],
						  "type": "EXEC",
						  "operation":"forceKeyspaceCleanup(java.lang.String,[Ljava.lang.String;)"},
				      "value": 0,
				      "timestamp": 1528850319,
				      "status": 200}`
			return httpmock.NewStringResponse(200, fmt.Sprintf(response, keyspace)), nil
		},
	)
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	err := jolokiaClient.NodeCleanup()
	if err != nil {
		t.Errorf("NodeCleanupKeyspace failed with : %s", err)
	}

	if !reflect.DeepEqual(keyspacescleaned, []string{"system_auth", "demo1", "demo2"}) {
		t.Errorf("Keyspaces cleaned are incorrect %v", keyspacescleaned)
	}
}

func TestNodeUpgradeSSTables(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	keyspacesUpgraded := []string{}

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			var execrequestdata execRequestData
			if err := json.NewDecoder(req.Body).Decode(&execrequestdata); err != nil {
				t.Error("Can't decode request received")
			}
			if execrequestdata.Attribute == "Keyspaces" {
				return httpmock.NewStringResponse(200, keyspaceListString()), nil
			}
			keyspace, ok := execrequestdata.Arguments[0].(string)

			if !ok {
				t.Error("Keyspace can't be nil")
			}

			keyspacesUpgraded = append(keyspacesUpgraded, keyspace)

			response := `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
						  "arguments": ["%s", true, 0, []],
						  "type": "EXEC",
						  "operation":"upgradeSSTables(java.lang.String,boolean,int,[Ljava.lang.String;)"},
				      "value": 0,
				      "timestamp": 1528850319,
				      "status": 200}`
			return httpmock.NewStringResponse(200, fmt.Sprintf(response, keyspace)), nil
		},
	)
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	err := jolokiaClient.NodeUpgradeSSTables(0)
	if err != nil {
		t.Errorf("NodeUpgradeSSTables failed with : %s", err)
	}

	if !reflect.DeepEqual(keyspacesUpgraded, allKeyspaces) {
		t.Errorf("Keyspaces cleaned are incorrect %v", keyspacesUpgraded)
	}
}

func TestNodeRebuild(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		httpmock.NewStringResponder(200, `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
							       "arguments": ["dc1"],
							       "type": "exec",
							       "operation": "rebuild(java.lang.String)"},
						   "value": null,
					  	   "timestamp": 1528848808,
						   "status": 200}`))
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	err := jolokiaClient.NodeRebuild("dc1")
	if err != nil {
		t.Errorf("NodeRebuild failed with : %s", err)
	}
}

func TestHasStreamingSessions(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{"request":
                       {"mbean": "org.apache.cassandra.net:type=StreamManager",
                        "attribute": "CurrentStreams",
                        "type": "read"},
                    "value": [
                        {
                          "rxPercentage": 0,
                          "currentTxBytes": 0,
                          "sessions": [
                            {
                              "connecting": "10.101.101.101",
                              "receivingFiles": [
                                {
                                  "direction": "IN",
                                  "planId": "47635800-a162-11e8-a49e-f17b1b4ecefc",
                                  "totalBytes": 167812827,
                                  "peer": "10.101.101.101",
                                  "currentBytes": 20270895,
                                  "sessionIndex": 1,
                                  "fileName": "k1/t1"
                                }
                              ],
                              "sessionIndex": 1,
                              "sendingFiles": [],
                              "sendingSummaries": [],
                              "peer": "10.101.101.101",
                              "receivingSummaries": [
                                {
                                  "files": 134,
                                  "totalSize": 18069892972,
                                  "cfId": "5ea36500-a1c3-11e6-a76c-d559db24120b"
                                }
                              ],
                              "planId": "47635800-a162-11e8-a49e-f17b1b4ecefc",
                              "state": "PREPARING"
                            }
                          ],
                          "description": "Bulk Load",
                          "planId": "47635800-a162-11e8-a49e-f17b1b4ecefc",
                          "totalTxBytes": 0,
                          "txPercentage": 100,
                          "totalRxBytes": 50037695339,
                          "currentRxBytes": 313601038
                        }
                      ],
                    "timestamp": 1528850319,
                    "status": 200}`), nil
		},
	)
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	isRebuilding, err := jolokiaClient.hasStreamingSessions()
	if err != nil {
		t.Errorf("hasStreamingSessions failed with : %s", err)
	}
	if isRebuilding != true {
		t.Errorf("hasStreamingSessions returns a bad answer")
	}

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{"request":
                   {"mbean": "org.apache.cassandra.net:type=StreamManager",
                    "attribute": "CurrentStreams",
                    "type": "read"},
                "value": [],
                "timestamp": 1528850319,
                "status": 200}`), nil
		},
	)
	isRebuilding, err = jolokiaClient.hasStreamingSessions()
	if err != nil {
		t.Errorf("hasStreamingSessions failed with : %s", err)
	}
	if isRebuilding == true {
		t.Errorf("hasStreamingSessions returns a bad answer")
	}
}

func TestHasCleanupCompactions(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{
				"request":
                       {"mbean": "org.apache.cassandra.db:type=CompactionManager",
                        "attribute": "Compactions",
                        "type": "read"},
                    "value": [
                        {
                            "columnfamily": "c1",
                            "compactionId": "3ecfc0b0-b766-11e8-b858-83f032e630aa",
                            "completed": "1479012280",
                            "id": "07d434c2-482f-3363-8be0-8254efc37c3c",
                            "keyspace": "k1",
                            "taskType": "Cleanup",
                            "total": "18988685023",
                            "unit": "bytes"
                          },
                          {
                            "columnfamily": "c2",
                            "compactionId": "6c60a3a0-b766-11e8-b858-83f032e630aa",
                            "completed": "885288195",
                            "id": "5ea36500-a1c3-11e6-a76c-d559db24120b",
                            "keyspace": "k2",
                            "taskType": "Validation",
                            "total": "269754147",
                            "unit": "bytes"
                          }
                      ],
                    "timestamp": 1528850319,
                    "status": 200}`), nil
		},
	)
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	isCleaningUp, err := jolokiaClient.hasCleanupCompactions()
	if err != nil {
		t.Errorf("hasCleanupCompactions failed with : %s", err)
	}
	if isCleaningUp != true {
		t.Errorf("hasCleanupCompactions returns a bad answer")
	}

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, `{"request":
                    {"mbean": "org.apache.cassandra.db:type=CompactionManager",
                     "attribute": "Compactions",
                     "type": "read"},
                  "value": [
                        {
                            "columnfamily": "c3",
                            "compactionId": "6c60a3a0-c344-11e8-b858-83f032e630aa",
                            "completed": "885288195",
                            "id": "5ea36500-a1c3-22fa-a76c-d559db24120b",
                            "keyspace": "k1",
                            "taskType": "Validation",
                            "total": "269754147",
                            "unit": "bytes"
                        }
                    ],
                  "timestamp": 1528850319,
                  "status": 200}`), nil
		},
	)
	isCleaningUp, err = jolokiaClient.hasCleanupCompactions()
	if err != nil {
		t.Errorf("hasCleanupCompactions failed with : %s", err)
	}
	if isCleaningUp == true {
		t.Errorf("hasCleanupCompactions returns a bad answer")
	}
}

func TestReplicateData(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	keyspacesDescribed := []string{}

	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		func(req *http.Request) (*http.Response, error) {
			var execrequestdata execRequestData
			if err := json.NewDecoder(req.Body).Decode(&execrequestdata); err != nil {
				t.Error("Can't decode request received")
			}
			if execrequestdata.Attribute == "Keyspaces" {
				return httpmock.NewStringResponse(200, keyspaceListString()), nil
			}
			keyspace, ok := execrequestdata.Arguments[0].(string)

			if !ok {
				t.Error("Keyspace can't be nil")
			}

			keyspacesDescribed = append(keyspacesDescribed, keyspace)

			response := `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
						  "arguments": ["%s"],
						  "type": "exec",
						  "operation": "describeRingJMX"},
				      "timestamp": 1541908753,
				      "status": 200,`

			//  For keyspace demo1 and demo2 we return token ranges with some of them assigned to nodes on dc2
			if keyspace[:4] == "demo" {
				response += `"value":
					       ["TokenRange(start_token:4572538884437204647, end_token:4764428918503636065, endpoints:[10.244.3.8], rpc_endpoints:[10.244.3.8], endpoint_details:[EndpointDetails(host:10.244.3.8, datacenter:dc1, rack:rack1)])",
						"TokenRange(start_token:-8547872176065322335, end_token:-8182289314504856691, endpoints:[10.244.2.5], rpc_endpoints:[10.244.2.5], endpoint_details:[EndpointDetails(host:10.244.2.5, datacenter:dc1, rack:rack1)])",
						"TokenRange(start_token:-2246208089217404881, end_token:-2021878843377619999, endpoints:[10.244.2.5], rpc_endpoints:[10.244.2.5], endpoint_details:[EndpointDetails(host:10.244.2.5, datacenter:dc2, rack:rack1)])",
						"TokenRange(start_token:-1308323778199165410, end_token:-1269907200339273513, endpoints:[10.244.2.6], rpc_endpoints:[10.244.2.6], endpoint_details:[EndpointDetails(host:10.244.2.6, datacenter:dc1, rack:rack1)])",
						"TokenRange(start_token:8544184416734424972, end_token:8568577617447026631, endpoints:[10.244.2.6], rpc_endpoints:[10.244.2.6], endpoint_details:[EndpointDetails(host:10.244.2.6, datacenter:dc2, rack:rack1)])",
						"TokenRange(start_token:2799723085723957315, end_token:3289697029162626204, endpoints:[10.244.3.7], rpc_endpoints:[10.244.3.7], endpoint_details:[EndpointDetails(host:10.244.3.7, datacenter:dc1, rack:rack1)])"]}`
				return httpmock.NewStringResponse(200, fmt.Sprintf(response, keyspace)), nil
			}
			return httpmock.NewStringResponse(200, fmt.Sprintf(response+`"value": []}`, keyspace)), nil
		},
	)

	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	keyspacesWithData, err := jolokiaClient.HasDataInDC("dc2")

	if err != nil {
		t.Errorf("hasData failed with : %s", err)
	}

	// Only demo1 and demo2 have token ranges showing data in dc2
	if !reflect.DeepEqual(keyspacesWithData, []string{"demo1", "demo2"}) {
		t.Errorf("Keyspaces having data are incorrect %v", keyspacesWithData)
	}

	// We confirm that all non local keyspaces have been scanned
	if !reflect.DeepEqual(keyspacesDescribed, []string{"system_auth", "demo1", "demo2"}) {
		t.Errorf("Keyspaces described are incorrect %v", keyspacesDescribed)
	}

}

func TestNodeDecommission(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		httpmock.NewStringResponder(200, `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
							       "arguments": [],
							       "type": "exec",
							       "operation": "decommission"},
						   "value": null,
					  	   "timestamp": 1528848808,
						   "status": 200}`))
	jolokiaClient, _ := NewJolokiaClient(host, port, nil, v1.LocalObjectReference{}, "ns")
	err := jolokiaClient.NodeDecommision()
	if err != nil {
		t.Errorf("NodeDecommision failed with : %v", err)
	}
}

func TestNodeOperationMode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		httpmock.NewStringResponder(200, `{"request":
				{"mbean": "org.apache.cassandra.db:type=StorageService",
				 "attribute": "OperationMode",
				 "type": "read"},
			"value": "NORMAL",
			"timestamp": 1528850319,
			"status": 200}`))
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	operationMode, err := jolokiaClient.NodeOperationMode()
	if err != nil {
		t.Errorf("NodeOperationMode failed with : %v", err)
	}
	if operationMode != "NORMAL" {
		t.Errorf("NodeOperationMode returned a bad answer: %s", operationMode)
	}
}

func TestLeavingNodes(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		httpmock.NewStringResponder(200, `{"request":
				{"mbean": "org.apache.cassandra.db:type=StorageService",
				 "attribute": "LeavingNodes",
				 "type": "read"},
			"value": ["127.0.0.1"],
			"timestamp": 1528850319,
			"status": 200}`))
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	leavingNodes, err := jolokiaClient.leavingNodes()
	if err != nil {
		t.Errorf("leavingNodes failed with : %v", err)
	}
	if !reflect.DeepEqual(leavingNodes, []string{"127.0.0.1"}) {
		t.Errorf("leavingNodes returned a bad answer: %s", leavingNodes)
	}
}

func TestHostIDMap(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", JolokiaURL(host, port),
		httpmock.NewStringResponder(200, `{"request":
				{"mbean": "org.apache.cassandra.db:type=StorageService",
				 "attribute": "LeavingNodes",
				 "type": "read"},
			"value": {"10.244.3.20": "ac0b9f2b-1eb4-40ca-bc6e-68b37575f019"},
			"timestamp": 1528850319,
			"status": 200}`))
	jolokiaClient, _ := NewJolokiaClient(host, JolokiaPort, nil,
		v1.LocalObjectReference{}, "ns")
	hostIDMap, err := jolokiaClient.hostIDMap()
	if err != nil {
		t.Errorf("hostIDMap failed with : %v", err)
	}
	if !reflect.DeepEqual(hostIDMap,
		map[string]string{"10.244.3.20": "ac0b9f2b-1eb4-40ca-bc6e-68b37575f019"}) {
		t.Errorf("hostIDMap returned a bad answer: %s", hostIDMap)
	}
}
