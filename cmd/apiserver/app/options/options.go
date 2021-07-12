/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"crypto/tls"
	"devops.kubesphere.io/plugin/pkg/client/cache"
	"devops.kubesphere.io/plugin/pkg/client/devops/jenkins"
	"flag"
	"fmt"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"

	"devops.kubesphere.io/plugin/pkg/apis"
	"devops.kubesphere.io/plugin/pkg/apiserver"
	"devops.kubesphere.io/plugin/pkg/client/clientset/versioned/scheme"
	apiserverconfig "devops.kubesphere.io/plugin/pkg/config"
	"devops.kubesphere.io/plugin/pkg/informers"
	genericoptions "devops.kubesphere.io/plugin/pkg/server/options"

	"net/http"
	"strings"

	"devops.kubesphere.io/plugin/pkg/client/k8s"
)

const fakeInterface string = "FAKE"

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Config:                  apiserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config: s.Config,
	}

	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		return nil, err
	}
	apiServer.KubernetesClient = kubernetesClient

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes(), kubernetesClient.KubeSphere(),
		kubernetesClient.ApiExtensions())
	apiServer.InformerFactory = informerFactory

	var cacheClient cache.Interface
	if s.RedisOptions != nil && len(s.RedisOptions.Host) != 0 {
		if s.RedisOptions.Host == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.RedisOptions, stopCh)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to redis service, please check redis status, error: %v", err)
			}
			apiServer.CacheClient = cacheClient
		}
	} else {
		klog.Warning("ks-apiserver starts without redis provided, it will use in memory cache. " +
			"This may cause inconsistencies when running ks-apiserver with multiple replicas.")
		apiServer.CacheClient = cache.NewSimpleCache()
	}

	if s.JenkinsOptions.Host != "" {
		devopsClient, err := jenkins.NewDevopsClient(s.JenkinsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jenkins, please check jenkins status, error: %v", err)
		}
		apiServer.DevopsClient = devopsClient
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	if s.GenericServerRunOptions.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(s.GenericServerRunOptions.TlsCertFile, s.GenericServerRunOptions.TlsPrivateKey)
		if err != nil {
			return nil, err
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		server.Addr = fmt.Sprintf(":%d", s.GenericServerRunOptions.SecurePort)
	}

	sch := scheme.Scheme
	if err := apis.AddToScheme(sch); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	apiServer.RuntimeCache, err = runtimecache.New(apiServer.KubernetesClient.Config(), runtimecache.Options{Scheme: sch})
	if err != nil {
		klog.Fatalf("unable to create runtime cache: %v", err)
	}

	apiServer.Server = server

	return apiServer, nil
}
