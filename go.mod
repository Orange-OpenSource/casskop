module github.com/Orange-OpenSource/casskop

go 1.13

require (
	emperror.dev/errors v0.7.0
	github.com/allamand/godebug v0.0.0-20190404121221-3ec752cd7166
	github.com/antihax/optional v1.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.0.1
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/erdrix/cassandrasidecar-go-client v0.0.0-20200611115046-09f54690a507
	// v0.0.0-20200610140227-df43b1138272 > works
	// github.com/erdrix/cassandrasidecar-go-client v0.0.0-20200610142749-d41f80ea3264 # Object
	// github.com/erdrix/cassandrasidecar-go-client v0.0.0-20200610133341-a1a47e86a0c1 # interface
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/go-resty/resty/v2 v2.1.0
	github.com/google/uuid v1.1.1
	github.com/jarcoal/httpmock v1.0.4
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/operator-framework/operator-sdk v0.15.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.2.1
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/swarvanusg/go_jolokia v0.0.0-20190213021437-3cd2b3fc4f36
	github.com/thoas/go-funk v0.4.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200120151820-655fe14d7479 // indirect
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	sigs.k8s.io/controller-runtime v0.4.0
//	sigs.k8s.io/controller-runtime v0.1.12
//	sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde // indirect
)

// Pinned to kubernetes-1.16.2
replace (
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
