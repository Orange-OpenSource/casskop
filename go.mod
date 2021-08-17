module github.com/Orange-OpenSource/casskop

go 1.17

require (
	emperror.dev/errors v0.7.0
	github.com/Jeffail/gabs v1.4.0
	github.com/allamand/godebug v0.0.0-20190404121221-3ec752cd7166
	github.com/antihax/optional v1.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.3
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/instaclustr/instaclustr-icarus-go-client v0.0.0-20210427160512-5264f1cbba08
	github.com/jarcoal/httpmock v1.0.4
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/nsf/jsondiff v0.0.0-20200515183724-f29ed568f4ce
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/prometheus/client_golang v1.5.1
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.5.1
	github.com/swarvanusg/go_jolokia v0.0.0-20190213021437-3cd2b3fc4f36
	github.com/thoas/go-funk v0.4.0
	github.com/zput/zxcTool v1.3.6
	golang.org/x/net v0.0.0-20210716203947-853a461950ff  // indirect
	google.golang.org/appengine v1.6.6 // indirect
	k8s.io/api v0.19.13
	k8s.io/apimachinery v0.19.13
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d
	sigs.k8s.io/controller-runtime v0.6.5
	sigs.k8s.io/structured-merge-diff/v2 v2.0.1 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	k8s.io/api => k8s.io/api v0.19.13
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.13
	k8s.io/client-go => k8s.io/client-go v0.19.13 // Required by prometheus-operator
)
