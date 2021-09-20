package common

func MockRestoreResponse(
	noDeleteDownloads bool,
	concurrentConnections int32,
	state,
	snapshotTag,
	operationId,
	k8sSecretName,
	storageLocation,
	restorationPhase,
	schemaVersion string) map[string]interface{} {

	return map[string]interface{}{
		"type":                    "restore",
		"id":                      operationId,
		"state":                   state,
		"progress":                0.1,
		"creationTime":            "2020-06-10T04:53:05.976Z",
		"startTime":               "2020-06-10T05:53:05.976Z",
		"completionTime":          "2020-06-10T06:53:05.976Z",
		"storageLocation":         storageLocation,
		"concurrentConnections":   concurrentConnections,
		"cassandraDirectory":      "/var/lib/cassandra",
		"snapshotTag":             snapshotTag,
		"entities":                "",
		"restorationStrategyType": "HARDLINKS",
		"restorationPhase":        restorationPhase,
		"noDeleteDownloads":       noDeleteDownloads,
		"schemaVersion":           schemaVersion,
		"k8sNamespace":            "default",
		"k8sSecretName":           k8sSecretName,
		"globalRequest":           true,
	}
}