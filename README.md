[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/kubesphere-sigs/ks-devops-plugin)
[![codecov](https://codecov.io/gh/kubesphere-sigs/ks-devops-plugin/branch/master/graph/badge.svg?token=XS8g2CjdNL)](https://codecov.io/gh/kubesphere-sigs/ks-devops-plugin)

## Get started

1. Install this plugin via `make install-chart`
1. Add the following configuration into ConfigMap `kubesphere-system/kubesphere-config`,
    and restart `ks-apiserver`

```
data:
  kubesphere.yaml: |
    devops:
      devopsPluginServiceAddress: ks-devops-plugin.kubesphere-devops-system:9090
```
