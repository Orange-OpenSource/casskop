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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultBaseImage              string        = "orangeopensource/cassandra-image"
	defaultVersion                string        = "latest"
	defaultNbMaxConcurrentCleanup               = 2
	defaultMaxPodUnavailable                    = 1
	defaultNumTokens                            = 256
	defaultImagePullPolicy        v1.PullPolicy = v1.PullAlways

	DefaultCassandraDC   string = "dc1"
	DefaultCassandraRack string = "rack1"

	DefaultTerminationGracePeriodSeconds = 1800

	//DefaultDelayWait: wait 20 seconds (2x resyncPeriod) prior to follow status of an operation
	DefaultResyncPeriod = 10
	DefaultDelayWait    = 2 * DefaultResyncPeriod

	//DefaultDelayWaitForDecommission is the time to wait for the decommission to happen on the Pod
	//The operator will start again if it is not the case
	DefaultDelayWaitForDecommission = 120

	//DefaultUserID is the default ID to use in cassandra image (RunAsUser)
	DefaultUserID int64 = 1000
)

const (
	AnnotationLastApplied string = "cassandraclusters.db.orange.com/last-applied-configuration"
	//Phase du Cluster
	ClusterPhaseInitial string = "Initializing"
	ClusterPhaseRunning string = "Running"
	ClusterPhasePending string = "Pending"

	StatusOngoing     string = "Ongoing"    // The Action is Ongoing
	StatusDone        string = "Done"       // The Action id Done
	StatusToDo        string = "ToDo"       // The Action is marked as To-Do
	StatusFinalizing  string = "Finalizing" // The Action is between Ongoing and Done
	StatusContinue    string = "Continue"
	StatusConfiguring string = "Configuring"
	StatusManual      string = "Manual"
	StatusError       string = "Error"

	//Available actions
	ActionUpdateConfigMap   string = "UpdateConfigMap"
	ActionUpdateDockerImage string = "UpdateDockerImage"
	ActionUpdateSeedList    string = "UpdateSeedList"
	ActionRollingRestart    string = "RollingRestart"
	ActionUpdateResources   string = "UpdateResources"
	ActionUpdateStatefulSet string = "UpdateStatefulSet"
	ActionScaleUp           string = "ScaleUp"
	ActionScaleDown         string = "ScaleDown"

	ActionDeleteDC string = "ActionDeleteDC"

	ActionCorrectCRDConfig string = "CorrectCRDConfig" //The Operator has correct a bad CRD configuration

	//List of Pods Operations
	OperationUpgradeSSTables string = "upgradesstables"
	OperationCleanup         string = "cleanup"
	OperationDecommission    string = "decommission"
	OperationRebuild         string = "rebuild"
	OperationRemove          string = "remove"
)

// SetDefaults sets the default values for the cassandra spec and returns true if the spec was changed
func (cc *CassandraCluster) SetDefaults() bool {
	changed := false
	ccs := &cc.Spec
	if ccs.NodesPerRacks == 0 {
		ccs.NodesPerRacks = 1
		changed = true
	}
	if len(ccs.BaseImage) == 0 {
		ccs.BaseImage = defaultBaseImage
		changed = true
	}
	if len(ccs.ImagePullPolicy) == 0 {
		ccs.ImagePullPolicy = defaultImagePullPolicy
		changed = true
	}
	if len(ccs.Version) == 0 {
		ccs.Version = defaultVersion
		changed = true
	}
	if ccs.RunAsUser == nil {
		ccs.RunAsUser = func(i int64) *int64 { return &i }(DefaultUserID)
	}
	if len(cc.Status.Phase) == 0 {
		cc.Status.Phase = ClusterPhaseInitial
		if cc.InitCassandraRackList() < 1 {
			logrus.Errorf("[%s]: We should have at list One Rack, Please correct the Error", cc.Name)
		}
		cc.Status.SeedList = cc.InitSeedList()

		ccs.CheckStatefulSetsAreEqual = true
		ccs.GCStdout = true
		changed = true
	}
	if ccs.MaxPodUnavailable == 0 {
		ccs.MaxPodUnavailable = defaultMaxPodUnavailable
	}

	return changed
}

func (cc *CassandraCluster) ComputeLastAppliedConfiguration() ([]byte, error) {
	lastcc := cc.DeepCopy()
	//remove unnecessary fields
	lastcc.Annotations = nil
	lastcc.ResourceVersion = ""
	lastcc.Status = CassandraClusterStatus{}

	lastApplied, err := json.Marshal(lastcc)
	if err != nil {
		logrus.Errorf("[%s]: Cannot create last-applied-configuration = %v", cc.Name, err)
	}
	return lastApplied, err
}

//GetDCSize Return the Numbers of declared DC
func (cc *CassandraCluster) GetDCSize() int {
	return len(cc.Spec.Topology.DC)
}

func (cc *CassandraCluster) GetDCRackSize() int {
	var nb int = 0
	dcsize := cc.GetDCSize()
	for dc := 0; dc < dcsize; dc++ {
		nb += cc.GetRackSize(dc)
	}
	return nb
}

func (cc *CassandraCluster) GetStatusDCRackSize() int {
	return len(cc.Status.CassandraRackStatus)
}

//GetDCName return the name of the DC a indice dc
//or defaultName
func (cc *CassandraCluster) GetDCName(dc int) string {
	if dc >= cc.GetDCSize() {
		return DefaultCassandraDC
	}
	return cc.Spec.Topology.DC[dc].Name
}

func (cc *CassandraCluster) getDCNodesPerRacksFromIndex(dc int) int32 {
	if dc >= cc.GetDCSize() {
		return cc.Spec.NodesPerRacks
	}
	storeDC := cc.Spec.Topology.DC[dc]
	if storeDC.NodesPerRacks == nil {
		return cc.Spec.NodesPerRacks
	}
	return *storeDC.NodesPerRacks
}

func (cc *CassandraCluster) getDCNumTokensPerRacksFromIndex(dc int) int32 {
	if dc >= cc.GetDCSize() {
		return defaultNumTokens
	}
	storeDC := cc.Spec.Topology.DC[dc]
	if storeDC.NumTokens == nil {
		return defaultNumTokens
	}
	return *storeDC.NumTokens
}

//GetRAckSize return the numbers of the Rack in the DCat indice dc
func (cc *CassandraCluster) GetRackSize(dc int) int {
	if dc >= cc.GetDCSize() {
		return 0
	}
	return len(cc.Spec.Topology.DC[dc].Rack)
}

//GetRackName return the Name of the rack for DC at indice dc and Rack at indice rack
func (cc *CassandraCluster) GetRackName(dc int, rack int) string {
	if dc >= cc.GetDCSize() {
		return DefaultCassandraRack
	}
	if rack >= cc.GetRackSize(dc) {
		return DefaultCassandraRack
	}
	return cc.Spec.Topology.DC[dc].Rack[rack].Name
}

// GetDCRackName compute dcName + RackName to be used in statefulsets, services..
// it return empty if the name don't match with kubernetes domain name validation regexp
func (cc *CassandraCluster) GetDCRackName(dcName string, rackName string) string {
	var dcRackName string
	dcRackName = dcName + "-" + rackName
	var regex_name = regexp.MustCompile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
	if !regex_name.MatchString(dcRackName) {
		logrus.Errorf("%s don't match valide name service: a DNS-1035 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character", dcRackName)
		return ""
	}
	return dcRackName
}

//GetDCFromDCRackName send dc name from dcRackName (dc-rack)
func (cc *CassandraCluster) GetDCFromDCRackName(dcRackName string) string {
	dc, _ := cc.GetDCAndRackFromDCRackName(dcRackName)
	return dc
}

//GetDCAndRackFromDCRackName send dc and rack from dcRackName (dc-rack)
func (cc *CassandraCluster) GetDCAndRackFromDCRackName(dcRackName string) (string, string) {
	dc := strings.Split(dcRackName, "-")
	return dc[0], dc[1]
}

// initTopology Initialisation of topology section in CRD
func (cc *CassandraCluster) initTopology(dcName string, rackName string) {
	cc.Spec.Topology = Topology{
		DC: []DC{
			DC{
				Name: dcName,
				Rack: []Rack{
					Rack{
						Name: rackName,
					},
				},
			},
		},
	}
}

// InitCassandraRack Initialisation of a CassandraRack Structure which is appended to the CRD status
func (cc *CassandraCluster) initCassandraRack(dcName string, rackName string) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	var rackStatus = CassandraRackStatus{
		Phase: ClusterPhaseInitial,
		CassandraLastAction: CassandraLastAction{
			Name:   ClusterPhaseInitial,
			Status: StatusOngoing,
		},
	}

	//The key of each CassandraRackStatus is the name of "<dcName>-<rackName>"
	cc.Status.CassandraRackStatus[dcRackName] = &rackStatus
}

// InitCassandraRack Initialisation of a CassandraRack Structure which is appended to the CRD status
// In this method we create it in status var instead of directly in cc object
// This is because except for init the cc, ca always work with a separate status which updates the cc
// in a defer statement in Reconcile method
func (cc *CassandraCluster) InitCassandraRackinStatus(status *CassandraClusterStatus, dcName string, rackName string) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	var rackStatus CassandraRackStatus = CassandraRackStatus{
		Phase: ClusterPhaseInitial,
		CassandraLastAction: CassandraLastAction{
			Name:   ClusterPhaseInitial,
			Status: StatusOngoing,
		},
	}

	//The key of each CassandraRackStatus is the name of "<dcName>-<rackName>"
	status.CassandraRackStatus[dcRackName] = &rackStatus
}

// Initialisation of the Cassandra SeedList
// We want 3 sides nodes for each DC
func (cc *CassandraCluster) InitSeedList() []string {

	var dcName, rackName string
	var nbRack int = 0
	var indice int32
	var seedList []string

	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		dcName = DefaultCassandraDC
		rackName = DefaultCassandraRack
		nbRack++
		for indice = 0; indice < cc.Spec.NodesPerRacks && indice < 3; indice++ {
			cc.addNewSeed(&seedList, dcName, rackName, indice)
		}
	} else {
		for dc := 0; dc < dcsize; dc++ {
			dcName = cc.GetDCName(dc)
			var nbSeedInDC int = 0

			racksize := cc.GetRackSize(dc)
			if racksize < 1 {
				rackName = DefaultCassandraRack
				nbRack++
				for indice = 0; indice < cc.Spec.NodesPerRacks && indice < 3; indice++ {
					cc.addNewSeed(&seedList, dcName, rackName, indice)
				}
			} else {

				for rack := 0; rack < racksize; rack++ {
					rackName = cc.GetRackName(dc, rack)
					dcRackName := cc.GetDCRackName(dcName, rackName)
					nbRack++
					nodesPerRacks := cc.GetNodesPerRacks(dcRackName)

					switch racksize {
					case 1:
						for indice = 0; indice < nodesPerRacks && indice < 3 && nbSeedInDC < 3; indice++ {
							cc.addNewSeed(&seedList, dcName, rackName, indice)
							nbSeedInDC++
						}
					case 2:
						for indice = 0; indice < nodesPerRacks && indice < 2 && nbSeedInDC < 3; indice++ {
							cc.addNewSeed(&seedList, dcName, rackName, indice)
							nbSeedInDC++
						}
					default:
						if nbSeedInDC < 3 {
							cc.addNewSeed(&seedList, dcName, rackName, 0)
							nbSeedInDC++
						}
					}

				}
			}
		}
	}
	return seedList
}

func (cc *CassandraCluster) GetSeedList(seedListTab *[]string) string {
	seedList := strings.Join(*seedListTab, ",")
	return seedList
}

func (cc *CassandraCluster) addNewSeed(seedList *[]string, dcName string, rackName string, indice int32) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	seed := fmt.Sprintf("%s-%s-%d.%s-%s.%s", cc.Name, dcRackName, indice, cc.Name, dcRackName, cc.Namespace)
	*seedList = append(*seedList, seed)
}

func (cc *CassandraCluster) IsPodInSeedList(podName string) bool {
	for i := range cc.Status.SeedList {
		if cc.Status.SeedList[i] == podName {
			return true
		}
	}
	return false
}

//FixCassandraRackList will remove additional rack-list that don't exists anymore in Topology
//we recalculate new dcrackStatus from actual topology and we apply diff to original
func (cc *CassandraCluster) FixCassandraRackList(status *CassandraClusterStatus) []string {
	newcc := cc.DeepCopy()
	newcc.InitCassandraRackList()

	rackList := []string{}
	for dcRackName := range cc.Status.CassandraRackStatus {
		if _, ok := newcc.Status.CassandraRackStatus[dcRackName]; !ok {
			//The item does not exists anymore
			//we need to remove it
			delete(status.CassandraRackStatus, dcRackName)
			rackList = append(rackList, dcRackName)
		}
	}
	return rackList
}

func (cc *CassandraCluster) GetRemovedDCName(oldCRD *CassandraCluster) string {
	//dcsize := cc.GetDCSize()
	olddcsize := oldCRD.GetDCSize()

	for dc := 0; dc < olddcsize; dc++ {
		olddcName := oldCRD.GetDCName(dc)
		dcName := cc.GetDCName(dc)
		if olddcName != dcName {
			return olddcName
		}
	}
	return ""
}

//InitCassandraRackList initiate the Status structure for CassandraRack
func (cc *CassandraCluster) InitCassandraRackList() int {
	var dcName, rackName string
	var nbRack int = 0

	cc.Status.CassandraRackStatus = make(map[string]*CassandraRackStatus)
	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		dcName = DefaultCassandraDC
		rackName = DefaultCassandraRack
		nbRack++
		cc.initCassandraRack(dcName, rackName)
		cc.initTopology(dcName, rackName)
	} else {
		for dc := 0; dc < dcsize; dc++ {
			dcName = cc.GetDCName(dc)
			racksize := cc.GetRackSize(dc)
			if racksize < 1 {
				rackName = DefaultCassandraRack
				nbRack++
				cc.initCassandraRack(dcName, rackName)
				cc.initTopology(dcName, rackName)
			} else {

				for rack := 0; rack < racksize; rack++ {
					rackName = cc.GetRackName(dc, rack)
					nbRack++
					cc.initCassandraRack(dcName, rackName)
				}
			}
		}

	}
	return nbRack
}

// GetNodesPerRacks sends back the number of cassandra nodes to uses for this dc-rack
func (cc *CassandraCluster) GetNodesPerRacks(dcRackName string) int32 {
	nodesPerRacks := cc.GetDCNodesPerRacksFromDCRackName(dcRackName)
	return nodesPerRacks
}

//GetDCNodesPerRacksFromDCRackName send NodesPerRack used for the given dcRackName
func (cc *CassandraCluster) GetDCNodesPerRacksFromDCRackName(dcRackName string) int32 {
	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		return cc.Spec.NodesPerRacks
	}
	for dc := 0; dc < dcsize; dc++ {
		dcName := cc.GetDCName(dc)
		racksize := cc.GetRackSize(dc)
		if racksize < 1 {
			return cc.Spec.NodesPerRacks
		}
		for rack := 0; rack < racksize; rack++ {
			rackName := cc.GetRackName(dc, rack)
			if dcRackName == cc.GetDCRackName(dcName, rackName) {
				return cc.getDCNodesPerRacksFromIndex(dc)
			}
		}
	}
	return cc.Spec.NodesPerRacks
}

// GetNodesPerRacks sends back the number of cassandra nodes to uses for this dc-rack
func (cc *CassandraCluster) GetNumTokensPerRacks(dcRackName string) int32 {
	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		return defaultNumTokens
	}
	for dc := 0; dc < dcsize; dc++ {
		dcName := cc.GetDCName(dc)
		racksize := cc.GetRackSize(dc)
		if racksize < 1 {
			return defaultNumTokens
		}
		for rack := 0; rack < racksize; rack++ {
			rackName := cc.GetRackName(dc, rack)
			if dcRackName == cc.GetDCRackName(dcName, rackName) {
				return cc.getDCNumTokensPerRacksFromIndex(dc)
			}
		}
	}
	return defaultNumTokens
}

// GetRollingPartitionPerRacks return rollingPartition defined in spec.topology.dc[].rack[].rollingPartition
func (cc *CassandraCluster) GetRollingPartitionPerRacks(dcRackName string) int32 {
	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		return 0
	}
	for dc := 0; dc < dcsize; dc++ {
		dcName := cc.GetDCName(dc)
		racksize := cc.GetRackSize(dc)
		if racksize < 1 {
			return 0
		}
		for rack := 0; rack < racksize; rack++ {
			rackName := cc.GetRackName(dc, rack)
			if dcRackName == cc.GetDCRackName(dcName, rackName) {
				return cc.Spec.Topology.DC[dc].Rack[rack].RollingPartition
			}
		}
	}
	return 0
}

//GetDCNodesPerRacksFromName send NodesPerRack which is applied for the specified dc name
//return true if we found, and false if not
func (cc *CassandraCluster) GetDCNodesPerRacksFromName(dctarget string) (bool, int32) {
	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		return false, cc.Spec.NodesPerRacks
	}
	for dc := 0; dc < dcsize; dc++ {
		dcName := cc.GetDCName(dc)
		if dctarget == dcName {
			return true, cc.getDCNodesPerRacksFromIndex(dc)
		}
	}
	return false, cc.Spec.NodesPerRacks
}

//FindDCWithNodesTo0
func (cc *CassandraCluster) FindDCWithNodesTo0() (bool, string, int) {
	for dc := 0; dc < cc.GetDCSize(); dc++ {
		if cc.getDCNodesPerRacksFromIndex(dc) == int32(0) {
			dcName := cc.GetDCName(dc)
			return true, dcName, dc
		}
	}
	return false, "", 0
}

//knownDCs returns list of datacenters
func (cc *CassandraCluster) knownDCs(dcName string) []string {
	var dcList []string
	for _, dc := range cc.Spec.Topology.DC {
		dcList = append(dcList, dc.Name)
	}
	return dcList
}

//IsValidDC returns true if dcName is known
func (cc *CassandraCluster) IsValidDC(dcName string) bool {
	for _, dc := range cc.Spec.Topology.DC {
		if dc.Name == dcName {
			return true
		}
	}
	return false
}

//Remove elements from DC slice
func (dc *DCSlice) Remove(i int) {
	*dc = append((*dc)[:i], (*dc)[i+1:]...)
}

//Remove elements from Rack slice
func (rack *RackSlice) Remove(i int) {
	*rack = append((*rack)[:i], (*rack)[i+1:]...)
}

// CassandraClusterSpec defines the configuration of CassandraCluster
type CassandraClusterSpec struct {
	// Number of nodes to deploy for a Cassandra deployment in each Racks.
	// Default: 1.
	// If NodesPerRacks = 2 and there is 3 racks, the cluster will have 6 Cassandra Nodes
	NodesPerRacks int32 `json:"nodesPerRacks,omitempty"`

	// Base image to use for a Cassandra deployment.
	BaseImage string `json:"baseImage"`

	// Version of Cassandra to be deployed.
	Version string `json:"version"`

	//ImagePullPolicy define the pull poicy for C* docker image
	ImagePullPolicy v1.PullPolicy `json:"imagepullpolicy"`

	//RunAsUser define the id of the user to run in the Cassandra image
	RunAsUser *int64 `json:"runAsUser"`

	// Pod defines the policy for pods owned by cassandra operator.
	// This field cannot be updated once the CR is created.
	//Pod       *PodPolicy         `json:"pod,omitempty"`
	Resources CassandraResources `json:"resources,omitempty"`

	// HardAntiAffinity defines if the PodAntiAffinity of the
	// statefulsets has to be hard (it's soft by default)
	HardAntiAffinity bool `json:"hardAntiAffinity,omitempty"`

	//DeletePVC defines if the PVC must be deleted when the cluster is deleted
	//it is false by default
	DeletePVC bool `json:"deletePVC,omitempty"`

	//AutoPilot defines if the Operator can fly alone or if we need human action to trigger
	//Actions on specific Cassandra nodes
	//If autoPilot=true, the operator will set labels pod-operation-status=To-Do on Pods which allows him to
	// automatically triggers Action
	//If autoPilot=false, the operator will set labels pod-operation-status=Manual on Pods which won't automatically triggers Action
	AutoPilot                 bool `json:"autoPilot,omitempty"`
	CheckStatefulSetsAreEqual bool `json:"checkStatefulsetsAreEqual,omitempty"`

	//GCStdout set the parameter CASSANDRA_GC_STDOUT which configure the JVM -Xloggc: true by default
	GCStdout bool `json:"gcStdout,omitempty"`

	//AutoUpdateSeedList defines if the Operator automatically update the SeedList according to new cluster CRD topology
	//by default a boolean is false
	AutoUpdateSeedList bool `json:"autoUpdateSeedList,omitempty"`

	MaxPodUnavailable int32 `json:"maxPodUnavailable"` //Number of MasPodUnavailable used in the PDB

	//NbMaxConcurrentCleanup int32 `json:"nbMaxConcurrentCleanup,omitempty"`

	//Define the Capacity for Persistent Volume Claims in the local storage
	DataCapacity string `json:"dataCapacity,omitempty"`

	//Define StorageClass for Persistent Volume Claims in the local storage.
	DataStorageClass string `json:"dataStorageClass,omitempty"`

	// Deploy or Not Service that provide access to monitoring metrics
	//Exporter bool `json:"exporter,omitempty"`

	// Name of the ConfigMap for Cassandra configuration (cassandra.yaml)
	// If this is empty, operator will uses default cassandra.yaml from the baseImage
	// If this is not empty, operator will uses the cassandra.yaml from the Configmap instead
	ConfigMapName string `json:"configMapName,omitempty"`

	// Name of the secret to uses to authenticate on Docker registries
	// If this is empty, operator do nothing
	// If this is not empty, propagate the imagePullSecrets to the statefulsets
	ImagePullSecret v1.LocalObjectReference `json:"imagePullSecret,omitempty"`

	// JMX Secret if Set is used to set JMX_USER and JMX_PASSWORD
	ImageJolokiaSecret v1.LocalObjectReference `json:"imageJolokiaSecret,omitempty"`

	//Topology to create Cassandra DC and Racks and to target appropriate Kubernetes Nodes
	Topology Topology `json:"topology,omitempty"`
}

// Topology allow to configure the Cassandra Topology according to kubernetes Nodes labels
type Topology struct {
	//Liste of DC defined in the CassandraCluster
	DC DCSlice `json:"dc,omitempty"`
}

type DCSlice []DC
type RackSlice []Rack

// DC allow to configure Cassandra RC according to kubernetes nodeselector labels
type DC struct {
	//Name of the CassandraDC
	Name string `json:"name,omitempty"`
	//Labels used to target Kubernetes nodes
	Labels map[string]string `json:"labels,omitempty"`
	//List of Racks defined in the Cassandra DC
	Rack RackSlice `json:"rack,omitempty"`

	// Number of nodes to deploy for a Cassandra deployment in each Racks.
	// Default: 1.
	// Optional, if not filled, used value define in CassandraClusterSpec
	NodesPerRacks *int32 `json:"nodesPerRacks,omitempty"`

	//NumTokens : configure the CASSANDRA_NUM_TOKENS parameter which can be different for each DD
	NumTokens *int32 `json:"numTokens,omitempty"`
}

// Rack allow to configure Cassandra Rack according to kubernetes nodeselector labels
type Rack struct {
	//Name of the Rack
	Name string `json:"name,omitempty"`
	// Flag to tell the operator to trigger a rolling restart of the Rack
	RollingRestart bool `json:"rollingRestart,omitempty"`

	//The Partition to control the Statefulset Upgrade
	RollingPartition int32 `json:"rollingPartition,omitempty"`

	//Labels used to target Kubernetes nodes
	Labels map[string]string `json:"labels,omitempty"`
}

// PodPolicy defines the policy for pods owned by vault operator.
type PodPolicy struct {
	// Resources is the resource requirements for the containers.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// CassandraClusterResources sets the limits and requests for a container
type CassandraResources struct {
	Requests CPUAndMem `json:"requests,omitempty"`
	Limits   CPUAndMem `json:"limits,omitempty"`
}

// CPUAndMem defines how many cpu and ram the container will request/limit
type CPUAndMem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

//CassandraRackStatus defines states of Cassandra for 1 rack (1 statefulset)
type CassandraRackStatus struct {
	// Phase indicates the state this Cassandra cluster jumps in.
	// Phase goes as one way as below:
	//   Initial -> Running <-> updating
	Phase string `json:"phase,omitempty"`

	// CassandraLastAction is the set of Cassandra State & Actions: Active, Standby..
	CassandraLastAction CassandraLastAction `json:"cassandraLastAction,omitempty"`

	// PodLastOperation manage status for Pod Operation (nodetool cleanup, upgradesstables..)
	PodLastOperation PodLastOperation `json:"podLastOperation,omitempty"`
}

//CassandraClusterStatus defines Global state of CassandraCluster
type CassandraClusterStatus struct {
	// Phase indicates the state this Cassandra cluster jumps in.
	// Phase goes as one way as below:
	//   Initial -> Running <-> updating
	Phase string `json:"phase,omitempty"`

	// Store last action at cluster level
	LastClusterAction       string `json:"lastClusterAction,omitempty"`
	LastClusterActionStatus string `json:"lastClusterActionStatus,omitempty"`

	// Indicates if we need to paused specific actions
	//ActionPaused bool `json:"actionPaused,omitempty"`

	//seeList to be used in Cassandra's Pods (computed by the Operator)
	SeedList []string `json:"seedlist,omitempty"`

	//CassandraRackStatusList list les Status pour chaque Racks
	CassandraRackStatus map[string]*CassandraRackStatus `json:"cassandraRackStatus,omitempty"`
}

// CassandraLastAction defines status of the CassandraStatefulset
type CassandraLastAction struct {
	// Action is the specific actions that can be done on a Cassandra Cluster
	// such as cleanup, upgradesstables..
	Status string `json:"status,omitempty"`

	// Type d'action a effectuer : UpdateVersion, UpdateBaseImage, UpdateConfigMap..
	Name string `json:"Name,omitempty"`

	StartTime *metav1.Time `json:"startTime,omitempty"`
	EndTime   *metav1.Time `json:"endTime,omitempty"`

	// PodNames of updated Cassandra nodes. Updated means the Cassandra container image version
	// matches the spec's version.
	UpdatedNodes []string `json:"updatedNodes,omitempty"`
}

// PodLastOperation is managed via labels on Pods set by an administrator
type PodLastOperation struct {
	Name string `json:"Name,omitempty"`

	Status string `json:"status,omitempty"`

	StartTime *metav1.Time `json:"startTime,omitempty"`
	EndTime   *metav1.Time `json:"endTime,omitempty"`

	//List of pods running an operation
	Pods []string `json:"pods,omitempty"`
	//List of pods that run an operation successfully
	PodsOK []string `json:"podsOK,omitempty"`
	//List of pods that fail to run an operation
	PodsKO []string `json:"podsKO,omitempty"`

	// Name of operator
	OperatorName string `json:"operatorName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraCluster is the Schema for the cassandraclusters API
// +k8s:openapi-gen=true
type CassandraCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraClusterSpec   `json:"spec,omitempty"`
	Status CassandraClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraClusterList contains a list of CassandraCluster
type CassandraClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraCluster{}, &CassandraClusterList{})
}
