module github.com/Orange-OpenSource/casskop

go 1.17

require (
	emperror.dev/errors v0.7.0
	github.com/Jeffail/gabs v1.4.0
	github.com/allamand/godebug v0.0.0-20190404121221-3ec752cd7166
	github.com/antihax/optional v1.0.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.3
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.4.0
	github.com/instaclustr/instaclustr-icarus-go-client v0.0.0-20210427160512-5264f1cbba08
	github.com/jarcoal/httpmock v1.0.4
	github.com/mitchellh/mapstructure v1.1.2
	github.com/nsf/jsondiff v0.0.0-20200515183724-f29ed568f4ce
	github.com/prometheus/client_golang v1.11.0
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/swarvanusg/go_jolokia v0.0.0-20190213021437-3cd2b3fc4f36
	github.com/thoas/go-funk v0.4.0
	github.com/zput/zxcTool v1.3.6
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20211110012726-3cc51fd1e909 // indirect
	sigs.k8s.io/controller-runtime v0.9.0
)

require github.com/operator-framework/operator-lib v0.9.0

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/antonfisher/nested-logrus-formatter v1.0.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apiextensions-apiserver v0.21.1 // indirect
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible // Required by OLM
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v1.13.0
	k8s.io/api => k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.3
	k8s.io/apiserver => k8s.io/apiserver v0.19.3
	k8s.io/client-go => k8s.io/client-go v0.19.3 // Required by prometheus-operator
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
