module github.com/Orange-OpenSource/casskop/multi-casskop

go 1.13

require (
	admiralty.io/multicluster-controller v0.2.0
	admiralty.io/multicluster-service-account v0.6.0
	github.com/Orange-OpenSource/casskop v0.4.1
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/go-bindata/go-bindata v3.1.2+incompatible // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/helm/helm-2to3 v0.5.1 // indirect
	github.com/imdario/mergo v0.3.8
	github.com/jessevdk/go-flags v1.4.0
	github.com/kylelemons/godebug v1.1.0
	github.com/martinlindhe/base36 v1.0.0 // indirect
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200321030439-57b580e57e88 // indirect
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/sirupsen/logrus v1.5.0
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/sample-controller v0.0.0-20181221200518-c3a5aa93b2bf
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/Orange-OpenSource/casskop => ../

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
