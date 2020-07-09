module github.com/Orange-OpenSource/casskop/multi-casskop

go 1.13

require (
	admiralty.io/multicluster-controller v0.6.0
	admiralty.io/multicluster-service-account v0.6.0
	github.com/Orange-OpenSource/casskop v0.4.1
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/imdario/mergo v0.3.8
	github.com/jessevdk/go-flags v1.4.0
	github.com/kylelemons/godebug v1.1.0
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/sirupsen/logrus v1.5.0
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/sample-controller v0.0.0-20181221200518-c3a5aa93b2bf
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/Orange-OpenSource/casskop => ../

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.2 // Required by admiralty
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by admiralty
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.0
)
