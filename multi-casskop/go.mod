module github.com/Orange-OpenSource/casskop/multi-casskop

go 1.17

require (
	admiralty.io/multicluster-controller v0.6.0
	admiralty.io/multicluster-service-account v0.6.0
	github.com/Orange-OpenSource/casskop v0.4.1
	github.com/imdario/mergo v0.3.12
	github.com/jessevdk/go-flags v1.5.0
	github.com/kylelemons/godebug v1.1.0
	github.com/sirupsen/logrus v1.8.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/sample-controller v0.19.13
	sigs.k8s.io/controller-runtime v0.10.0
)

require (
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.15.0 // indirect
)

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/antihax/optional v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/instaclustr/instaclustr-icarus-go-client v0.0.0-20210427160512-5264f1cbba08 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.22.1 // indirect
	k8s.io/apiextensions-apiserver v0.22.1 // indirect
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/utils v0.0.0-20210707171843-4b05e18ac7d9 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/Orange-OpenSource/casskop => ../

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v1.13.0
	k8s.io/api => k8s.io/api v0.22.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.0
	k8s.io/apiserver => k8s.io/apiserver v0.22.0
	k8s.io/client-go => k8s.io/client-go v0.22.0 // Required by prometheus-operator
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
