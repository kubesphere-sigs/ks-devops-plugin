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

package v1alpha2

import (
	"devops.kubesphere.io/plugin/pkg/apiserver/authorization/authorizer"
	"devops.kubesphere.io/plugin/pkg/informers"
	"fmt"
	"k8s.io/client-go/kubernetes"

	"github.com/emicklei/go-restful"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"

	"devops.kubesphere.io/plugin/pkg/api"
	"devops.kubesphere.io/plugin/pkg/apiserver/query"
	"devops.kubesphere.io/plugin/pkg/apiserver/request"
	kubesphere "devops.kubesphere.io/plugin/pkg/client/clientset/versioned"
	resourcev1alpha3 "devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/resource"
	"devops.kubesphere.io/plugin/pkg/models/tenant"
)

type tenantHandler struct {
	tenant tenant.Interface
}

func newTenantHandler(factory informers.InformerFactory, k8sclient kubernetes.Interface,
	ksclient kubesphere.Interface, authorizer authorizer.Authorizer, resourceGetter *resourcev1alpha3.ResourceGetter) *tenantHandler {
	return &tenantHandler{
		tenant: tenant.New(factory, k8sclient, ksclient, authorizer, resourceGetter),
	}
}

func (h *tenantHandler) ListDevOpsProjects(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(req)

	var workspaceMember user.Info
	if username := req.PathParameter("workspacemember"); username != "" {
		workspaceMember = &user.DefaultInfo{
			Name: username,
		}
	} else {
		requestUser, ok := request.UserFrom(req.Request.Context())
		if !ok {
			err := fmt.Errorf("cannot obtain user info")
			klog.Errorln(err)
			api.HandleForbidden(resp, nil, err)
			return
		}
		workspaceMember = requestUser
	}

	fmt.Println(workspaceMember, workspace, queryParam)
	result, err := h.tenant.ListDevOpsProjects(workspaceMember, workspace, queryParam)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}
