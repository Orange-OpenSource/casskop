module github.com/Orange-OpenSource/casskop/multi-casskop

go 1.16

require (
	admiralty.io/multicluster-controller v0.6.0
	admiralty.io/multicluster-service-account v0.6.0
	github.com/Orange-OpenSource/casskop v0.4.1
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/imdario/mergo v0.3.9
	github.com/jessevdk/go-flags v1.4.0
	github.com/kylelemons/godebug v1.1.0
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/sirupsen/logrus v1.7.1
	k8s.io/apimachinery v0.19.13
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/sample-controller v0.19.13
	sigs.k8s.io/controller-runtime v0.6.5
)

replace github.com/Orange-OpenSource/casskop => ../

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.3 // Required by admiralty
	k8s.io/client-go => k8s.io/client-go v0.19.3 // Required by admiralty
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
