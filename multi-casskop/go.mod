module github.com/Orange-OpenSource/casskop/multi-casskop

go 1.17

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

require (
	cloud.google.com/go v0.51.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/antihax/optional v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/instaclustr/instaclustr-icarus-go-client v0.0.0-20210427160512-5264f1cbba08 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/magefile/mage v1.10.0 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.19.13 // indirect
	k8s.io/klog/v2 v2.2.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d // indirect
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/Orange-OpenSource/casskop => ../

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.3 // Required by admiralty
	k8s.io/client-go => k8s.io/client-go v0.19.3 // Required by admiralty
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
