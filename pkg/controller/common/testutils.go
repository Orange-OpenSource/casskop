package common

import (
	"fmt"
	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	v12 "k8s.io/api/apps/v1"
	"os"
	"path/filepath"
	"testing"
)

func HelperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func HelperGetStatefulset(t *testing.T, dcRackName string) *v12.StatefulSet {
	var sts v12.StatefulSet
	name := fmt.Sprintf("cassandracluster-2DC-%s-sts.yaml", dcRackName)
	yaml.Unmarshal(HelperLoadBytes(t, name), &sts)
	return &sts
}

func HelperInitCassandraBackup(cassandraBackupYaml string) v1alpha1.CassandraBackup {
	var cassandraBackup v1alpha1.CassandraBackup
	if err := yaml.Unmarshal([]byte(cassandraBackupYaml), &cassandraBackup); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	return cassandraBackup
}

func AssertEvent(t *testing.T, event chan string, message string) {
	assert := assert.New(t)
	eventMessage := <-event
	assert.Contains(eventMessage, message)
}

