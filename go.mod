module github.com/Orange-OpenSource/casskop

go 1.13

require (
	github.com/allamand/godebug v0.0.0-20190404121221-3ec752cd7166
	github.com/banzaicloud/k8s-objectmatcher v1.3.3
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/go-resty/resty/v2 v2.1.0
	github.com/google/uuid v1.1.1
	github.com/jarcoal/httpmock v1.0.4
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/prometheus/client_golang v1.5.1
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.5.1
	github.com/swarvanusg/go_jolokia v0.0.0-20190213021437-3cd2b3fc4f36
	github.com/thoas/go-funk v0.4.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200316234421-82d701f24f9d
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
