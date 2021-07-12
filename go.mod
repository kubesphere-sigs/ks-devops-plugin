module devops.kubesphere.io/plugin

go 1.13

require (
	github.com/PuerkitoBio/goquery v1.7.1
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/beevik/etree v1.1.0
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-openapi/spec v0.19.3
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/google/go-cmp v0.5.4
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.3
	github.com/open-policy-agent/opa v0.30.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/apiserver v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/component-base v0.18.6
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	kubesphere.io/api v0.0.0-20210511124541-08f2d682bd07
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/fatih/structs => github.com/fatih/structs v1.1.0
	github.com/go-redis/redis => github.com/go-redis/redis v6.15.2+incompatible
	github.com/golang/example => github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	github.com/kubesphere/sonargo => github.com/kubesphere/sonargo v0.0.2
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apiserver => k8s.io/apiserver v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/component-base => k8s.io/component-base v0.18.6
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	kubesphere.io/devops => github.com/linuxsuren/ks v0.0.37
	sigs.k8s.io/application => sigs.k8s.io/application v0.8.4-0.20201016185654-c8e2959e57a0
	sigs.k8s.io/kubefed => sigs.k8s.io/kubefed v0.6.1
)
