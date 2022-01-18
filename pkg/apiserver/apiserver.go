/*
Copyright 2019 The KubeSphere Authors.

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

package apiserver

import (
	"bytes"
	"context"
	clusterv1alpha1 "devops.kubesphere.io/plugin/pkg/api/cluster/v1alpha1"
	tenantv1alpha1 "devops.kubesphere.io/plugin/pkg/api/tenant/v1alpha1"
	"devops.kubesphere.io/plugin/pkg/apiserver/authentication/authenticators/jwttoken"
	"devops.kubesphere.io/plugin/pkg/apiserver/authentication/request/anonymous"
	"devops.kubesphere.io/plugin/pkg/apiserver/authorization/rbac"
	"devops.kubesphere.io/plugin/pkg/apiserver/filters"
	"devops.kubesphere.io/plugin/pkg/apiserver/request"
	iamapi "devops.kubesphere.io/plugin/pkg/kapis/iam/v1alpha2"
	resourcesv1alpha2 "devops.kubesphere.io/plugin/pkg/kapis/resources/v1alpha2"
	resourcev1alpha3 "devops.kubesphere.io/plugin/pkg/kapis/resources/v1alpha3"
	tenantv1alpha2 "devops.kubesphere.io/plugin/pkg/kapis/tenant/v1alpha2"
	"devops.kubesphere.io/plugin/pkg/models/auth"
	"devops.kubesphere.io/plugin/pkg/models/iam/am"
	"devops.kubesphere.io/plugin/pkg/models/iam/group"
	"devops.kubesphere.io/plugin/pkg/models/iam/im"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/loginrecord"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/user"
	"fmt"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	unionauth "k8s.io/apiserver/pkg/authentication/request/union"
	"net/http"
	rt "runtime"
	"time"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/klog"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"

	"devops.kubesphere.io/plugin/pkg/client/cache"
	"devops.kubesphere.io/plugin/pkg/client/devops"
	"devops.kubesphere.io/plugin/pkg/client/k8s"
	apiserverconfig "devops.kubesphere.io/plugin/pkg/config"
	"devops.kubesphere.io/plugin/pkg/informers"
	utilnet "devops.kubesphere.io/plugin/pkg/utils/net"
)

const (
	// ApiRootPath defines the root path of all KubeSphere apis.
	ApiRootPath = "/kapis"

	// MimeMergePatchJson is the mime header used in merge request
	MimeMergePatchJson = "application/merge-patch+json"

	//
	MimeJsonPatchJson = "application/json-patch+json"
)

type APIServer struct {
	// number of kubesphere apiserver
	ServerCount int

	Server *http.Server

	Config *apiserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	// kubeClient is a collection of all kubernetes(include CRDs) objects clientset
	KubernetesClient k8s.Client

	// informerFactory is a collection of all kubernetes(include CRDs) objects informers,
	// mainly for fast query
	InformerFactory informers.InformerFactory

	// cache is used for short lived objects, like session
	CacheClient cache.Interface

	DevopsClient devops.Interface

	// controller-runtime cache
	RuntimeCache runtimecache.Cache
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	s.installKubeSphereAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container

	s.buildHandlerChain(stopCh)

	return nil
}

// Install all kubesphere api groups
// Installation happens before all informers start to cache objects, so
//   any attempt to list objects using listers will get empty results.
func (s *APIServer) installKubeSphereAPIs() {
	imOperator := im.NewOperator(s.KubernetesClient.KubeSphere(),
		user.New(s.InformerFactory.KubeSphereSharedInformerFactory(),
			s.InformerFactory.KubernetesSharedInformerFactory()),
		loginrecord.New(s.InformerFactory.KubeSphereSharedInformerFactory()),
		s.Config.AuthenticationOptions)
	amOperator := am.NewOperator(s.KubernetesClient.KubeSphere(),
		s.KubernetesClient.Kubernetes(),
		s.InformerFactory)
	rbacAuthorizer := rbac.NewRBACAuthorizer(amOperator)

	urlruntime.Must(tenantv1alpha2.AddToContainer(s.container, s.InformerFactory,
		s.KubernetesClient.Kubernetes(),
		s.KubernetesClient.KubeSphere(), rbacAuthorizer, s.RuntimeCache))

	urlruntime.Must(resourcev1alpha3.AddToContainer(s.container, s.InformerFactory, s.RuntimeCache))
	urlruntime.Must(resourcesv1alpha2.AddToContainer(s.container, s.KubernetesClient.Kubernetes(), s.InformerFactory,
		s.KubernetesClient.Master()))

	urlruntime.Must(iamapi.AddToContainer(s.container, imOperator, amOperator,
		group.New(s.InformerFactory, s.KubernetesClient.KubeSphere(), s.KubernetesClient.Kubernetes()),
		rbacAuthorizer))
}

func (s *APIServer) Run(stopCh <-chan struct{}) (err error) {

	err = s.waitForResourceSync(stopCh)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stopCh
		_ = s.Server.Shutdown(ctx)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.NewString("api", "kapi"),
		GlobalResources: []schema.GroupResource{
			tenantv1alpha1.Resource(tenantv1alpha1.ResourcePluralWorkspace),
			tenantv1alpha2.Resource(tenantv1alpha1.ResourcePluralWorkspace),
			tenantv1alpha2.Resource(clusterv1alpha1.ResourcesPluralCluster),
		},
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, s.KubernetesClient.Config(), &errorResponder{})

	authenticators := make([]authenticator.Request, 0)
	authenticators = append(authenticators, anonymous.NewAuthenticator())

	switch s.Config.AuthMode {
	case apiserverconfig.AuthModeToken:
		authenticators = append(authenticators,
			//bearertoken.New(devopsbearertoken.New()),
			bearertoken.New(jwttoken.NewTokenAuthenticator(auth.NewTokenOperator(s.CacheClient,
				s.Config.AuthenticationOptions),
				s.InformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().Users().Lister())),
		)
	default:
		// TODO error handle
	}

	handler = filters.WithAuthentication(handler, unionauth.New(authenticators...))
	handler = filters.WithRequestInfo(handler, requestInfoResolver)

	s.Server.Handler = handler
}

func (s *APIServer) waitForResourceSync(stopCh <-chan struct{}) error {
	klog.V(0).Info("Start cache objects")

	discoveryClient := s.KubernetesClient.Kubernetes().Discovery()
	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	// resources we have to create informer first
	k8sGVRs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "", Version: "v1", Resource: "resourcequotas"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "serviceaccounts"},

		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "controllerrevisions"},
		{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
		{Group: "autoscaling", Version: "v2beta2", Resource: "horizontalpodautoscalers"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	}

	for _, gvr := range k8sGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := s.InformerFactory.KubernetesSharedInformerFactory().ForResource(gvr)
			if err != nil {
				klog.Errorf("cannot create informer for %s", gvr)
				return err
			}
		}
	}

	s.InformerFactory.KubernetesSharedInformerFactory().Start(stopCh)
	s.InformerFactory.KubernetesSharedInformerFactory().WaitForCacheSync(stopCh)

	ksInformerFactory := s.InformerFactory.KubeSphereSharedInformerFactory()

	ksGVRs := []schema.GroupVersionResource{}

	devopsGVRs := []schema.GroupVersionResource{
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibinaries"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuildertemplates"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2iruns"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuilders"},
		{Group: "devops.kubesphere.io", Version: "v1alpha3", Resource: "devopsprojects"},
		{Group: "devops.kubesphere.io", Version: "v1alpha3", Resource: "pipelines"},
	}

	// skip caching devops resources if devops not enabled
	if s.DevopsClient != nil {
		ksGVRs = append(ksGVRs, devopsGVRs...)
	}

	for _, gvr := range ksGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = ksInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}

	ksInformerFactory.Start(stopCh)
	ksInformerFactory.WaitForCacheSync(stopCh)

	apiextensionsInformerFactory := s.InformerFactory.ApiExtensionSharedInformerFactory()
	apiextensionsGVRs := []schema.GroupVersionResource{
		{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
	}

	for _, gvr := range apiextensionsGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err = apiextensionsInformerFactory.ForResource(gvr)
			if err != nil {
				return err
			}
		}
	}
	apiextensionsInformerFactory.Start(stopCh)
	apiextensionsInformerFactory.WaitForCacheSync(stopCh)

	// controller runtime cache for resources
	go s.RuntimeCache.Start(stopCh)
	s.RuntimeCache.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")

	return nil

}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		utilnet.GetRequestIP(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}
