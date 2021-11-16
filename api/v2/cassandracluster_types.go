package v2

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
)

const (
	DefaultLivenessInitialDelaySeconds int32 = 120
	DefaultLivenessHealthCheckTimeout  int32 = 20
	DefaultLivenessHealthCheckPeriod   int32 = 10

	DefaultReadinessInitialDelaySeconds int32 = 60
	DefaultReadinessHealthCheckTimeout  int32 = 10
	DefaultReadinessHealthCheckPeriod   int32 = 10

	defaultCassandraImage     = "cassandra:3.11.10"
	defaultBootstrapImage     = "orangeopensource/cassandra-bootstrap:0.1.9"
	defaultConfigBuilderImage = "datastax/cass-config-builder:1.0.4"

	DefaultBackRestImage      = "gcr.io/cassandra-operator/instaclustr-icarus:1.1.0"
	defaultServiceAccountName = "cassandra-cluster-node"
	defaultMaxPodUnavailable  = 1
	defaultImagePullPolicy    = v1.PullAlways

	DefaultCassandraDC   = "dc1"
	DefaultCassandraRack = "rack1"

	DefaultTerminationGracePeriodSeconds = 1800

	DefaultResyncPeriod = 10
	//DefaultDelayWait wait 20 seconds (2x resyncPeriod) prior to follow status of an operation
	DefaultDelayWait = 2 * DefaultResyncPeriod

	//DefaultDelayWaitForDecommission is the time to wait for the decommission to happen on the Pod
	//The operator will start again if it is not the case
	DefaultDelayWaitForDecommission = 120
)

// ClusterStateInfo describe a cluster state
type ClusterStateInfo struct {
	ID   float64
	Name string
}

var (
	//Cluster phases
	ClusterPhaseInitial = ClusterStateInfo{1, "Initializing"}
	ClusterPhaseRunning = ClusterStateInfo{2, "Running"}
	ClusterPhasePending = ClusterStateInfo{3, "Pending"}

	//Available actions
	ActionUpdateConfigMap   = ClusterStateInfo{1, "UpdateConfigMap"}
	ActionUpdateDockerImage = ClusterStateInfo{2, "UpdateDockerImage"}
	ActionUpdateSeedList    = ClusterStateInfo{3, "UpdateSeedList"}
	ActionRollingRestart    = ClusterStateInfo{4, "RollingRestart"}
	ActionUpdateResources   = ClusterStateInfo{5, "UpdateResources"}
	ActionUpdateStatefulSet = ClusterStateInfo{6, "UpdateStatefulSet"}
	ActionScaleUp           = ClusterStateInfo{7, "ScaleUp"}
	ActionScaleDown         = ClusterStateInfo{8, "ScaleDown"}

	ActionDeleteDC   = ClusterStateInfo{9, "ActionDeleteDC"}
	ActionDeleteRack = ClusterStateInfo{10, "ActionDeleteRack"}

	ActionCorrectCRDConfig = ClusterStateInfo{11, "CorrectCRDConfig"} //The Operator has correct a bad CRD configuration

	regexDCRackName = regexp.MustCompile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
)

const (
	AnnotationLastApplied string = "cassandraclusters.db.orange.com/last-applied-configuration"

	StatusOngoing     string = "Ongoing"    // The Action is Ongoing
	StatusDone        string = "Done"       // The Action id Done
	StatusToDo        string = "ToDo"       // The Action is marked as To-Do
	StatusFinalizing  string = "Finalizing" // The Action is between Ongoing and Done
	StatusContinue    string = "Continue"
	StatusConfiguring string = "Configuring"
	StatusManual      string = "Manual"
	StatusError       string = "Error"

	//List of Pods Operations
	OperationUpgradeSSTables string = "upgradesstables"
	OperationCleanup         string = "cleanup"
	OperationDecommission    string = "decommission"
	OperationRebuild         string = "rebuild"
	OperationRemove          string = "remove"

	BreakResyncLoop    = true
	ContinueResyncLoop = false
)

// CheckDefaults checks that required fields haven't good values
func (cc *CassandraCluster) CheckDefaults() {
	ccs := &cc.Spec

	if len(ccs.CassandraImage) == 0 {
		ccs.CassandraImage = defaultCassandraImage
	}

	if len(ccs.ImagePullPolicy) == 0 {
		ccs.ImagePullPolicy = defaultImagePullPolicy
	}

	if len(ccs.BootstrapImage) == 0 {
		ccs.BootstrapImage = defaultBootstrapImage
	}

	if len(ccs.ConfigBuilderImage) == 0 {
		ccs.ConfigBuilderImage = defaultConfigBuilderImage
	}

	if len(ccs.ServiceAccountName) == 0 {
		ccs.ServiceAccountName = defaultServiceAccountName
	}

	if ccs.ReadOnlyRootFilesystem == nil {
		ccs.ReadOnlyRootFilesystem = func(b bool) *bool { return &b }(true)
	}

	// LivenessProbe dynamic config
	if ccs.LivenessInitialDelaySeconds == nil {
		ccs.LivenessInitialDelaySeconds = func(i int32) *int32 { return &i }(DefaultLivenessInitialDelaySeconds)
	}
	if ccs.LivenessHealthCheckTimeout == nil {
		ccs.LivenessHealthCheckTimeout = func(i int32) *int32 { return &i }(DefaultLivenessHealthCheckTimeout)
	}
	if ccs.LivenessHealthCheckPeriod == nil {
		ccs.LivenessHealthCheckPeriod = func(i int32) *int32 { return &i }(DefaultLivenessHealthCheckPeriod)
	}

	// ReadinessProbe dynamic config
	if ccs.ReadinessInitialDelaySeconds == nil {
		ccs.ReadinessInitialDelaySeconds = func(i int32) *int32 { return &i }(DefaultReadinessInitialDelaySeconds)
	}
	if ccs.ReadinessHealthCheckTimeout == nil {
		ccs.ReadinessHealthCheckTimeout = func(i int32) *int32 { return &i }(DefaultReadinessHealthCheckTimeout)
	}
	if ccs.ReadinessHealthCheckPeriod == nil {
		ccs.ReadinessHealthCheckPeriod = func(i int32) *int32 { return &i }(DefaultReadinessHealthCheckPeriod)
	}

	// BackupRestore default config
	if ccs.BackRestSidecar == nil {
		ccs.BackRestSidecar = &BackRestSidecar{Image: DefaultBackRestImage}
	} else if ccs.BackRestSidecar.Image == "" {
		ccs.BackRestSidecar.Image = DefaultBackRestImage
	}
}

// SetDefaults sets the default values for the cassandra spec and returns true if the spec was changed
// SetDefault mus be done only once at startup
func (cc *CassandraCluster) SetDefaults() bool {
	changed := false
	ccs := &cc.Spec
	if ccs.NodesPerRacks == 0 {
		ccs.NodesPerRacks = 1
		changed = true
	}
	if len(cc.Status.Phase) == 0 {
		cc.Status.Phase = ClusterPhaseInitial.Name
		if cc.InitCassandraRackList() < 1 {
			logrus.Errorf("[%s]: We should have at list One Rack, Please correct the Error", cc.Name)
		}
		if cc.Status.SeedList == nil {
			cc.Status.SeedList = cc.InitSeedList()
		}
		changed = true
	}
	if ccs.MaxPodUnavailable == 0 {
		ccs.MaxPodUnavailable = defaultMaxPodUnavailable
		changed = true
	}
	if cc.Spec.Resources.Limits == nil {
		cc.Spec.Resources.Limits = cc.Spec.Resources.Requests
		changed = true
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

//GetRackSize return the numbers of the Rack in the DC at indice dc
func (cc *CassandraCluster) GetRackSize(dc int) int {
	if dc >= cc.GetDCSize() {
		return 0
	}
	return len(cc.Spec.Topology.DC[dc].Rack)
}

//GetRackName return the Name of the rack for DC at index dc and Rack at index rack
func (cc *CassandraCluster) GetRackName(dc int, rack int) string {
	if dc >= cc.GetDCSize() || rack >= cc.GetRackSize(dc) {
		return DefaultCassandraRack
	}
	return cc.Spec.Topology.DC[dc].Rack[rack].Name
}

// GetDCRackName compute dcName + RackName to be used in statefulsets, services..
// it returns empty if the name don't match with kubernetes domain name validation regexp
func (cc *CassandraCluster) GetDCRackName(dcName string, rackName string) string {
	dcRackName := dcName + "-" + rackName
	if regexDCRackName.MatchString(dcRackName) {
		return dcRackName
	}
	logrus.Errorf("%s is not a valid service name: a DNS-1035 label must consist of lower case "+
		"alphanumeric characters or '-', and must start and end with an alphanumeric character", dcRackName)
	return ""
}

//GetDCNameFromDCRackName send dc name from dcRackName (dc-rack)
func (cc *CassandraCluster) GetDCNameFromDCRackName(dcRackName string) string {
	dc, _ := cc.GetDCNameAndRackNameFromDCRackName(dcRackName)
	return dc
}

//GetDCAndRackFromDCRackName send dc and rack from dcRackName (dc-rack)
func (cc *CassandraCluster) GetDCNameAndRackNameFromDCRackName(dcRackName string) (string, string) {
	dc := strings.Split(dcRackName, "-")
	return dc[0], dc[1]
}

// initTopology Initialisation of topology section in CRD
func (cc *CassandraCluster) initTopology(dcName string, rackName string) {
	cc.Spec.Topology = Topology{
		DC: []DC{
			{
				Name: dcName,
				Rack: []Rack{
					{
						Name: rackName,
					},
				},
			},
		},
	}
}

// InitCassandraRackStatus Initializes a CassandraRack Structure
// In this method we create it in status var instead of directly in cc object
// because except for init the cc can always work with a separate status which updates the cc
// in a defer statement in Reconcile method
func (cc *CassandraCluster) InitCassandraRackStatus(status *CassandraClusterStatus, dcName string, rackName string) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	rackStatus := CassandraRackStatus{
		Phase: ClusterPhaseInitial.Name,
		CassandraLastAction: CassandraLastAction{
			Name:   ClusterPhaseInitial.Name,
			Status: StatusOngoing,
		},
	}

	status.CassandraRackStatus[dcRackName] = &rackStatus
}

// Initialisation of the Cassandra SeedList
// We want 3 seed nodes for each DC
func (cc *CassandraCluster) InitSeedList() []string {

	var dcName, rackName string
	var nbRack = 0
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
		return seedList
	}
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
			continue
		}
		for rack := 0; rack < racksize; rack++ {
			rackName = cc.GetRackName(dc, rack)
			dcRackName := cc.GetDCRackName(dcName, rackName)
			nbRack++
			nodesPerRacks := cc.GetNodesPerRacks(dcRackName)

			switch racksize {
			case 1, 2:
				for indice = 0; indice < nodesPerRacks && indice < int32(4-racksize) &&
					nbSeedInDC < 3; indice++ {
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
	return seedList
}

func (cc *CassandraCluster) SeedList(seedListTab *[]string) string {
	seedList := strings.Join(*seedListTab, ",")
	return seedList
}

func (cc *CassandraCluster) addNewSeed(seedList *[]string, dcName string, rackName string, indice int32) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	seed := fmt.Sprintf("%s-%s-%d.%s.%s", cc.Name, dcRackName, indice, cc.Name, cc.Namespace)
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
	var nbRack = 0

	cc.Status.CassandraRackStatus = make(map[string]*CassandraRackStatus)
	dcsize := cc.GetDCSize()

	if dcsize < 1 {
		dcName = DefaultCassandraDC
		rackName = DefaultCassandraRack
		nbRack++
		cc.InitCassandraRackStatus(&cc.Status, dcName, rackName)
		cc.initTopology(dcName, rackName)
	} else {
		for dc := 0; dc < dcsize; dc++ {
			dcName = cc.GetDCName(dc)
			racksize := cc.GetRackSize(dc)
			if racksize < 1 {
				rackName = DefaultCassandraRack
				nbRack++
				cc.InitCassandraRackStatus(&cc.Status, dcName, rackName)
				cc.initTopology(dcName, rackName)
			} else {
				for rack := 0; rack < racksize; rack++ {
					rackName = cc.GetRackName(dc, rack)
					nbRack++
					cc.InitCassandraRackStatus(&cc.Status, dcName, rackName)
				}
			}
		}

	}
	return nbRack
}

// GetDataCapacityForDC sends back the data capacity of cassandra nodes to uses for this dc
func (cc *CassandraCluster) GetDataCapacityForDC(dcName string) string {
	return cc.GetDataCapacityFromDCName(dcName)
}

// GetDataCapacityFromDCName send DataCapacity used for the given dcName
func (cc *CassandraCluster) GetDataCapacityFromDCName(dcName string) string {
	dcIndex := cc.GetDCIndexFromDCName(dcName)
	if dcIndex >= 0 {
		dc := cc.getDCFromIndex(dcIndex)
		if dc != nil && dc.DataCapacity != "" {
			return dc.DataCapacity
		}
		return cc.Spec.DataCapacity
	}
	return cc.Spec.DataCapacity
}

// GetDataCapacityForDC sends back the data storage class of cassandra nodes to uses for this dc
func (cc *CassandraCluster) GetDataStorageClassForDC(dcName string) string {
	return cc.GetDataStorageClassFromDCName(dcName)
}

// GetDataCapacityFromDCName send DataStorageClass used for the given dcName
func (cc *CassandraCluster) GetDataStorageClassFromDCName(dcName string) string {
	dcIndex := cc.GetDCIndexFromDCName(dcName)
	if dcIndex >= 0 {
		dc := cc.getDCFromIndex(dcIndex)
		if dc != nil && dc.DataCapacity != "" {
			return dc.DataStorageClass
		}
		return cc.Spec.DataStorageClass
	}
	return cc.Spec.DataStorageClass
}

func (cc *CassandraCluster) GetDCIndexFromDCName(dcName string) int {
	dcSize := cc.GetDCSize()
	if dcSize < 1 {
		return -1
	}

	for dc := 0; dc < dcSize; dc++ {
		if dcName == cc.GetDCName(dc) {
			return dc
		}
	}
	return -1
}

// getDCFromIndex send DC for the given index
func (cc *CassandraCluster) getDCFromIndex(dc int) *DC {
	if dc >= cc.GetDCSize() {
		return nil
	}
	return &cc.Spec.Topology.DC[dc]
}

// Get DC by one of its rack name
func (cc *CassandraCluster) GetDCFromDCRackName(dcRackName string) *DC {
	index := cc.GetDCIndexFromDCName(cc.GetDCNameFromDCRackName(dcRackName))
	return cc.getDCFromIndex(index)
}

// Get Rack by its rack name
func (cc *CassandraCluster) GetRackFromDCRackName(dcRackName string) *Rack {
	_, rackName := cc.GetDCNameAndRackNameFromDCRackName(dcRackName)
	dc := cc.GetDCFromDCRackName(dcRackName)
	for _, rack := range dc.Rack {
		if rack.Name == rackName {
			return &rack
		}
	}
	return nil
}

// GetNodesPerRacks sends back the number of cassandra nodes to uses for this dc-rack
func (cc *CassandraCluster) GetNodesPerRacks(dcRackName string) int32 {
	nodesPerRacks := cc.GetDCNodesPerRacksFromDCRackName(dcRackName)
	return nodesPerRacks
}

//GetDCNodesPerRacksFromDCRackName send NodesPerRack used for the given dcRackName
func (cc *CassandraCluster) GetDCRackNames() []string {
	dcsize := cc.GetDCSize()

	var dcRackNames = []string{}
	if dcsize < 1 {
		return dcRackNames
	}
	for dc := 0; dc < dcsize; dc++ {
		dcName := cc.GetDCName(dc)
		racksize := cc.GetRackSize(dc)
		if racksize < 1 {
			return dcRackNames
		}
		for rack := 0; rack < racksize; rack++ {
			rackName := cc.GetRackName(dc, rack)
			dcRackNames = append(dcRackNames, cc.GetDCRackName(dcName, rackName))
		}
	}
	return dcRackNames
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
func (dc *DCSlice) Remove(idx int) {
	*dc = append((*dc)[:idx], (*dc)[idx+1:]...)
}

//Remove elements from Rack slice
func (rack *RackSlice) Remove(idx int) {
	*rack = append((*rack)[:idx], (*rack)[idx+1:]...)
}

// CassandraClusterSpec defines the configuration of CassandraCluster

type CassandraClusterSpec struct {
	// Number of nodes to deploy for a Cassandra deployment in each Racks.
	// Default: 1.
	// If NodesPerRacks = 2 and there is 3 racks, the cluster will have 6 Cassandra Nodes
	NodesPerRacks int32 `json:"nodesPerRacks,omitempty"`

	// Image + version to use for Cassandra
	CassandraImage string `json:"cassandraImage,omitempty"`

	//ImagePullPolicy define the pull policy for C* docker image
	ImagePullPolicy v1.PullPolicy `json:"imagepullpolicy,omitempty"`

	// Image used for bootstrapping cluster (use format base:version)
	BootstrapImage string `json:"bootstrapImage,omitempty"`

	// Image used for configBuilder (use format base:version)
	ConfigBuilderImage string `json:"configBuilderImage,omitempty"`

	// RunAsUser define the id of the user to run in the Cassandra image
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default:=999
	RunAsUser int64 `json:"runAsUser,omitempty"`

	// FSGroup defines the GID owning volumes in the Cassandra image
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default:=1
	FSGroup int64 `json:"fsGroup,omitempty"`

	// Make the pod as Readonly
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`

	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// HardAntiAffinity defines if the PodAntiAffinity of the
	// statefulset has to be hard (it's soft by default)
	HardAntiAffinity bool `json:"hardAntiAffinity,omitempty"`

	Pod *PodPolicy `json:"pod,omitempty"`

	Service *ServicePolicy `json:"service,omitempty"`

	//DeletePVC defines if the PVC must be deleted when the cluster is deleted
	//it is false by default
	DeletePVC bool `json:"deletePVC,omitempty"`

	//Debug is used to surcharge Cassandra pod command to not directly start cassandra but
	//starts an infinite wait to allow user to connect a bash into the pod to make some diagnoses.
	Debug bool `json:"debug,omitempty"`

	//AutoPilot defines if the Operator can fly alone or if we need human action to trigger
	//Actions on specific Cassandra nodes
	//If autoPilot=true, the operator will set labels pod-operation-status=To-Do on Pods which allows him to
	// automatically triggers Action
	//If autoPilot=false, the operator will set labels pod-operation-status=Manual on Pods which won't automatically triggers Action
	AutoPilot          bool `json:"autoPilot,omitempty"`
	NoCheckStsAreEqual bool `json:"noCheckStsAreEqual,omitempty"`

	//AutoUpdateSeedList defines if the Operator automatically update the SeedList according to new cluster CRD topology
	//by default a boolean is false
	AutoUpdateSeedList bool `json:"autoUpdateSeedList,omitempty"`

	MaxPodUnavailable int32 `json:"maxPodUnavailable,omitempty"` //Number of MaxPodUnavailable used in the PDB

	// RestartCountBeforePodDeletion defines the number of restart allowed for a cassandra container allowed before
	// deleting the pod  to force its restart from scratch. if set to 0 or omit,
	// no action will be performed based on restart count.
	RestartCountBeforePodDeletion int32 `json:"restartCountBeforePodDeletion,omitempty"`

	// Very special Flag to hack CassKop reconcile loop - use with really good care
	UnlockNextOperation bool `json:"unlockNextOperation,omitempty"`

	// Define the Capacity for Persistent Volume Claims in the local storage
	// +kubebuilder:validation:Pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	DataCapacity string `json:"dataCapacity,omitempty"`

	//Define StorageClass for Persistent Volume Claims in the local storage.
	DataStorageClass string `json:"dataStorageClass,omitempty"`

	// StorageConfig defines additional storage configurations
	StorageConfigs []StorageConfig `json:"storageConfigs,omitempty"`

	// SidecarsConfig defines additional sidecar configurations
	SidecarConfigs []v1.Container `json:"sidecarConfigs,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`

	// Deploy or Not Service that provide access to monitoring metrics
	//Exporter bool `json:"exporter,omitempty"`

	// Name of the ConfigMap for Cassandra configuration (cassandra.yaml)
	// If this is empty, operator will uses default cassandra.yaml from the baseImage
	// If this is not empty, operator will uses the cassandra.yaml from the Configmap instead
	ConfigMapName string `json:"configMapName,omitempty"`

	// Version string for config builder https://github.com/datastax/cass-config-definitions,
	// used to generate Cassandra server configuration
	ServerVersion string `json:"serverVersion,omitempty"`

	// Server type: "cassandra" or "dse" for config builder, default to cassandra
	// +kubebuilder:validation:Enum=cassandra;dse
	// +kubebuilder:default:=cassandra
	ServerType string `json:"serverType,omitempty"`

	// Config for the Cassandra nodes
	// +kubebuilder:pruning:PreserveUnknownFields
	Config json.RawMessage `json:"config,omitempty"`

	// Name of the secret to uses to authenticate on Docker registries
	// If this is empty, operator do nothing
	// If this is not empty, propagate the imagePullSecrets to the statefulsets
	ImagePullSecret v1.LocalObjectReference `json:"imagePullSecret,omitempty"`

	// JMX Secret if Set is used to set JMX_USER and JMX_PASSWORD
	ImageJolokiaSecret v1.LocalObjectReference `json:"imageJolokiaSecret,omitempty"`

	//Topology to create Cassandra DC and Racks and to target appropriate Kubernetes Nodes
	Topology Topology `json:"topology,omitempty"`

	// LivenessInitialDelaySeconds defines initial delay for the liveness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	LivenessInitialDelaySeconds *int32 `json:"livenessInitialDelaySeconds,omitempty"`
	// LivenessHealthCheckTimeout defines health check timeout for the liveness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	LivenessHealthCheckTimeout *int32 `json:"livenessHealthCheckTimeout,omitempty"`
	// LivenessHealthCheckPeriod defines health check period for the liveness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	LivenessHealthCheckPeriod *int32 `json:"livenessHealthCheckPeriod,omitempty"`
	// LivenessFailureThreshold defines failure threshold for the liveness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	LivenessFailureThreshold *int32 `json:"livenessFailureThreshold,omitempty"`
	//LivenessSuccessThreshold defines success threshold for the liveness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	LivenessSuccessThreshold *int32 `json:"livenessSuccessThreshold,omitempty"`

	// ReadinessInitialDelaySeconds defines initial delay for the readiness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	ReadinessInitialDelaySeconds *int32 `json:"readinessInitialDelaySeconds,omitempty"`
	// ReadinessHealthCheckTimeout defines health check timeout for the readiness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	ReadinessHealthCheckTimeout *int32 `json:"readinessHealthCheckTimeout,omitempty"`
	// ReadinessHealthCheckPeriod defines health check period for the readiness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	ReadinessHealthCheckPeriod *int32 `json:"readinessHealthCheckPeriod,omitempty"`
	// ReadinessFailureThreshold defines failure threshold for the readiness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	ReadinessFailureThreshold *int32 `json:"readinessFailureThreshold,omitempty"`
	// ReadinessSuccessThreshold defines success threshold for the readiness probe of the main
	// cassandra container : https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes
	ReadinessSuccessThreshold *int32 `json:"readinessSuccessThreshold,omitempty"`
	// When process namespace sharing is enabled, processes in a container are visible to all other containers in that pod.
	// https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"varint,27,opt,name=shareProcessNamespace"`

	BackRestSidecar    *BackRestSidecar `json:"backRestSidecar,omitempty"`
	ServiceAccountName string           `json:"serviceAccountName,omitempty"`
}

// StorageConfig defines additional storage configurations
type StorageConfig struct {
	// Mount path into cassandra container
	MountPath string `json:"mountPath"`
	// Name of the pvc
	// +kubebuilder:validation:Pattern=[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	Name string `json:"name"`
	// Persistent volume claim spec
	PVCSpec *v1.PersistentVolumeClaimSpec `json:"pvcSpec"`
}

// Topology allow to configure the Cassandra Topology according to kubernetes Nodes labels
type Topology struct {
	//List of DC defined in the CassandraCluster
	DC DCSlice `json:"dc,omitempty"`
}

type DCSlice []DC
type RackSlice []Rack

// DC allow to configure Cassandra RC according to kubernetes nodeselector labels
type DC struct {
	//Name of the DC
	// +kubebuilder:validation:Pattern=^[^-]+$
	Name string `json:"name,omitempty"`
	//Labels used to target Kubernetes nodes
	Labels map[string]string `json:"labels,omitempty"`

	// Config for the Cassandra nodes
	// +kubebuilder:pruning:PreserveUnknownFields
	Config json.RawMessage `json:"config,omitempty"`

	//List of Racks defined in the Cassandra DC
	Rack RackSlice `json:"rack,omitempty"`

	// Number of nodes to deploy for a Cassandra deployment in each Racks.
	// Default: 1.
	// Optional, if not filled, used value define in CassandraClusterSpec
	NodesPerRacks *int32 `json:"nodesPerRacks,omitempty"`

	// Define the Capacity for Persistent Volume Claims in the local storage
	// +kubebuilder:validation:Pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	DataCapacity string `json:"dataCapacity,omitempty"`

	//Define StorageClass for Persistent Volume Claims in the local storage.
	DataStorageClass string `json:"dataStorageClass,omitempty"`

	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// Rack allow to configure Cassandra Rack according to kubernetes nodeselector labels
type Rack struct {
	//Name of the Rack
	// +kubebuilder:validation:Pattern=^[^-]+$
	Name string `json:"name,omitempty"`

	//Labels used to target Kubernetes nodes
	Labels map[string]string `json:"labels,omitempty"`

	// Config for the Cassandra nodes
	// +kubebuilder:pruning:PreserveUnknownFields
	Config json.RawMessage `json:"config,omitempty"`

	// Flag to tell the operator to trigger a rolling restart of the Rack
	RollingRestart bool `json:"rollingRestart,omitempty"`

	//The Partition to control the Statefulset Upgrade
	RollingPartition int32 `json:"rollingPartition,omitempty"`
}

// PodPolicy defines the policy for pods owned by CassKop operator.
type PodPolicy struct {
	// Annotations specifies the annotations to attach to headless service the CassKop operator creates
	Annotations map[string]string `json:"annotations,omitempty"`
	// Tolerations specifies the tolerations to attach to the pods the CassKop operator creates
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

// ServicePolicy defines the policy for headless service owned by CassKop operator.
type ServicePolicy struct {
	// Annotations specifies the annotations to attach to headless service the CassKop operator creates
	Annotations map[string]string `json:"annotations,omitempty"`
}

// BackRestSidecar defines details about cassandra-sidecar to load along with each C* pod
type BackRestSidecar struct {
	// Image of backup/restore sidecar
	Image string `json:"image,omitempty"`
	// ImagePullPolicy define the pull policy for backrest sidecar docker image
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Kubernetes object : https://godoc.org/k8s.io/api/core/v1#ResourceRequirements
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts []v1.VolumeMount `json:"volumeMount,omitempty"`
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

	//seedList to be used in Cassandra's Pods (computed by the Operator)
	SeedList []string `json:"seedlist,omitempty"`

	//
	CassandraNodesStatus map[string]CassandraNodeStatus `json:"cassandraNodeStatus,omitempty"`

	//CassandraRackStatusList list status for each Rack
	CassandraRackStatus map[string]*CassandraRackStatus `json:"cassandraRackStatus,omitempty"`
}

// CassandraLastAction defines status of the CassandraStatefulset
type CassandraLastAction struct {
	// Action is the specific actions that can be done on a Cassandra Cluster
	// such as cleanup, upgradesstables..
	Status string `json:"status,omitempty"`

	// Type of action to perform : UpdateVersion, UpdateBaseImage, UpdateConfigMap..
	Name string `json:"name,omitempty"`

	StartTime *metav1.Time `json:"startTime,omitempty"`
	EndTime   *metav1.Time `json:"endTime,omitempty"`

	// PodNames of updated Cassandra nodes. Updated means the Cassandra container image version
	// matches the spec's version.
	UpdatedNodes []string `json:"updatedNodes,omitempty"`
}

// PodLastOperation is managed via labels on Pods set by an administrator
type PodLastOperation struct {
	Name string `json:"name,omitempty"`

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

type CassandraNodeStatus struct {
	HostId string `json:"hostId,omitempty"`
	NodeIp string `json:"nodeIp,omitempty"`
}

// +kubebuilder:object:root=true

// CassandraCluster is the Schema for the cassandraclusters API
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:resource:path=cassandraclusters,scope=Namespaced,shortName=cassc;casscs
type CassandraCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	Spec   CassandraClusterSpec   `json:"spec,omitempty"`
	Status CassandraClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CassandraClusterList contains a list of CassandraCluster
type CassandraClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraCluster{}, &CassandraClusterList{})
}
