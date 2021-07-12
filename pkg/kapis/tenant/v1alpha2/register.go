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

package v1alpha2

import (
	"net/http"

	resourcev1alpha3 "devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/resource"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"devops.kubesphere.io/plugin/pkg/api"
	"devops.kubesphere.io/plugin/pkg/apiserver/authorization/authorizer"
	"devops.kubesphere.io/plugin/pkg/apiserver/runtime"
	kubesphere "devops.kubesphere.io/plugin/pkg/client/clientset/versioned"
	"devops.kubesphere.io/plugin/pkg/constants"
	"devops.kubesphere.io/plugin/pkg/informers"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, k8sclient kubernetes.Interface,
	ksclient kubesphere.Interface, authorizer authorizer.Authorizer,
	cache cache.Cache) error {

	ws := runtime.NewWebService(GroupVersion)
	handler := newTenantHandler(factory, k8sclient, ksclient, authorizer, resourcev1alpha3.NewResourceGetter(factory, cache))

	ws.Route(ws.GET("/workspaces/{workspace}/devops").
		To(handler.ListDevOpsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the devops projects of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/devops").
		To(handler.ListDevOpsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspacemember username")).
		Doc("List the devops projects of specified workspace for the workspace member").
		Reads(corev1.Namespace{}).
		Returns(http.StatusOK, api.StatusOK, corev1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectTag}))

	c.Add(ws)
	return nil
}
