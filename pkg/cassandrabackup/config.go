package cassandrabackup

import (
	"crypto/tls"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultCassandraSidecarPort  = 4567
	DefaultCassandraBackupSecure = false
)

type Config struct {
	UseSSL    bool
	TLSConfig *tls.Config
	Port      int32
	Host      string
	Timeout   time.Duration
}

func ClusterConfig(client controllerclient.Client, cluster *api.CassandraCluster, pod *corev1.Pod) *Config {
	conf        := &Config{}
	conf.UseSSL = DefaultCassandraBackupSecure
	conf.Port   = DefaultCassandraSidecarPort
	conf.Host   = k8s.PodHostname(*pod)

	if conf.Timeout == 0 {
		conf.Timeout = 1 * time.Minute
	}

	return conf
}
