module github.com/Orange-OpenSource/casskop

go 1.13

require github.com/Azure/go-autorest v12.2.0+incompatible // indirect

require (
	admiralty.io/multicluster-controller v0.2.0 // indirect
	admiralty.io/multicluster-service-account v0.6.0 // indirect
	bitbucket.org/ww/goautoneg v0.0.0-20120707110453-75cd24fc2f2c // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.4.11 // indirect
	github.com/Masterminds/sprig v0.0.0-20190301161902-9f8fceff796f // indirect
	github.com/allamand/godebug v0.0.0-20190404121221-3ec752cd7166
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/banzaicloud/k8s-objectmatcher v1.0.1
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-openapi/spec v0.19.4
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/jarcoal/httpmock v1.0.4
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/maxbrunsfeld/counterfeiter v0.0.0-20181017030959-1aadac120687 // indirect
	github.com/operator-framework/operator-sdk v0.15.0
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stevvooe/resumable v0.0.0-20180830230917-22b14a53ba50 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/swarvanusg/go_jolokia v0.0.0-20190213021437-3cd2b3fc4f36
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	github.com/thoas/go-funk v0.4.0
	go.etcd.io/bbolt v1.3.3 // indirect
	golang.org/x/build v0.0.0-20190314133821-5284462c4bec // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/square/go-jose.v2 v2.3.0 // indirect
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	k8s.io/sample-controller v0.0.0-20191121021213-d454fe81777c // indirect
	sigs.k8s.io/controller-runtime v0.4.0
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
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
