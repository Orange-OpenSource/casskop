package sidecarclient

import (
	"crypto/tls"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


const (
	DefaultCassandraSidecarPort = 4567
	DefaultCassandraSidecarSecure = false
)

type CassandraSidecarConfig struct {
	UseSSL    bool
	TLSConfig *tls.Config
	Port int32
	Pods []corev1.Pod
}

func ClusterConfig(client client.Client, cluster *api.CassandraCluster, podList *corev1.PodList) (*CassandraSidecarConfig, error) {
	conf := &CassandraSidecarConfig{}
	conf.UseSSL = DefaultCassandraSidecarSecure
	conf.Port = DefaultCassandraSidecarPort
	conf.Pods = podList.Items
}
