module github.com/Orange-OpenSource/cassandra-k8s-operator

go 1.12

require (
	admiralty.io/multicluster-controller v0.2.0 // indirect
	admiralty.io/multicluster-service-account v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/allamand/godebug v0.0.0-20190404121221-3ec752cd7166
	github.com/banzaicloud/k8s-objectmatcher v1.0.1
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/envy v1.7.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/jarcoal/httpmock v1.0.4
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/operator-framework/operator-sdk v0.9.0
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/swarvanusg/go_jolokia v0.0.0-20190213021437-3cd2b3fc4f36
	github.com/thoas/go-funk v0.4.0
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/genproto v0.0.0-20190701230453-710ae3a149df // indirect
	google.golang.org/grpc v1.22.0 // indirect
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apimachinery v0.0.0-20191121015412-41065c7a8c2a
	k8s.io/client-go v11.0.1-0.20190516230509-ae8359b20417+incompatible
	k8s.io/sample-controller v0.0.0-20191121021213-d454fe81777c // indirect
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
)
