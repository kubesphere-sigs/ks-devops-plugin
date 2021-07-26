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
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"devops.kubesphere.io/plugin/pkg/api"
	"devops.kubesphere.io/plugin/pkg/informers"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2/resource"
	"devops.kubesphere.io/plugin/pkg/server/params"
)

type resourceHandler struct {
	resourcesGetter     *resource.ResourceGetter
}

func newResourceHandler(k8sClient kubernetes.Interface, factory informers.InformerFactory, masterURL string) *resourceHandler {

	return &resourceHandler{
		resourcesGetter:     resource.NewResourceGetter(factory),
	}
}

func (r *resourceHandler) handleGetNamespacedResources(request *restful.Request, response *restful.Response) {
	r.handleListNamespaceResources(request, response)
}

func (r *resourceHandler) handleListNamespaceResources(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resource := request.PathParameter("resources")
	orderBy := params.GetStringValueWithDefault(request, params.OrderByParam, v1alpha2.CreateTime)
	limit, offset := params.ParsePaging(request)
	reverse := params.GetBoolValueWithDefault(request, params.ReverseParam, false)
	conditions, err := params.ParseConditions(request)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	result, err := r.resourcesGetter.ListResources(namespace, resource, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Error(err)
		api.HandleInternalError(response, nil, err)
		return
	}

	response.WriteEntity(result)
}
