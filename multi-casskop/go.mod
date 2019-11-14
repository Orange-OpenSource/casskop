module github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop

go 1.13

require (
	admiralty.io/multicluster-controller v0.2.0
	admiralty.io/multicluster-service-account v0.5.1
	cloud.google.com/go v0.46.3 // indirect
	github.com/Orange-OpenSource/cassandra-k8s-operator v0.4.1
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.8
	github.com/kylelemons/godebug v1.1.0
	github.com/operator-framework/operator-sdk v0.10.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	go.uber.org/multierr v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/net v0.0.0-20191007182048-72f939374954 // indirect
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/apimachinery v0.0.0-20191005115455-e71eb83a557c
	k8s.io/client-go v11.0.1-0.20190516230509-ae8359b20417+incompatible
	k8s.io/sample-controller v0.0.0-20191005120943-ac9726f261cc
	sigs.k8s.io/controller-runtime v0.1.12
)

replace (
	github.com/Orange-OpenSource/cassandra-k8s-operator => ../
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
)
