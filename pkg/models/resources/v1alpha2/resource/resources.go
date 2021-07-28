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

package resource

import (
	"errors"

	"k8s.io/klog"

	"devops.kubesphere.io/plugin/pkg/informers"
	"devops.kubesphere.io/plugin/pkg/models"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2/namespace"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2/s2buildertemplate"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2/s2ibuilder"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha2/s2irun"
	"devops.kubesphere.io/plugin/pkg/server/params"
	"devops.kubesphere.io/plugin/pkg/utils/sliceutil"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type ResourceGetter struct {
	resourcesGetters map[string]v1alpha2.Interface
}

func (r ResourceGetter) Add(resource string, getter v1alpha2.Interface) {
	if r.resourcesGetters == nil {
		r.resourcesGetters = make(map[string]v1alpha2.Interface)
	}
	r.resourcesGetters[resource] = getter
}

func NewResourceGetter(factory informers.InformerFactory) *ResourceGetter {
	resourceGetters := make(map[string]v1alpha2.Interface)

	resourceGetters[v1alpha2.S2iBuilders] = s2ibuilder.NewS2iBuilderSearcher(factory.KubeSphereSharedInformerFactory())
	resourceGetters[v1alpha2.S2iRuns] = s2irun.NewS2iRunSearcher(factory.KubeSphereSharedInformerFactory())
	resourceGetters[v1alpha2.S2iBuilderTemplates] = s2buildertemplate.NewS2iBuidlerTemplateSearcher(factory.KubeSphereSharedInformerFactory())

	resourceGetters[v1alpha2.Namespaces] = namespace.NewNamespaceSearcher(factory.KubernetesSharedInformerFactory())

	return &ResourceGetter{resourcesGetters: resourceGetters}
}

var (
	clusterResources = []string{v1alpha2.Nodes, v1alpha2.Workspaces, v1alpha2.Namespaces, v1alpha2.ClusterRoles, v1alpha2.StorageClasses, v1alpha2.S2iBuilderTemplates}
)

func (r *ResourceGetter) GetResource(namespace, resource, name string) (interface{}, error) {
	// none namespace resource
	if namespace != "" && sliceutil.HasString(clusterResources, resource) {
		return nil, ErrResourceNotSupported
	}
	if searcher, ok := r.resourcesGetters[resource]; ok {
		resource, err := searcher.Get(namespace, name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return resource, nil
	}
	return nil, ErrResourceNotSupported
}

func (r *ResourceGetter) ListResources(namespace, resource string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)
	var err error
	var result []interface{}

	// none namespace resource
	if namespace != "" && sliceutil.HasString(clusterResources, resource) {
		return nil, ErrResourceNotSupported
	}

	if searcher, ok := r.resourcesGetters[resource]; ok {
		result, err = searcher.Search(namespace, conditions, orderBy, reverse)
	} else {
		return nil, ErrResourceNotSupported
	}

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if limit == -1 || limit+offset > len(result) {
		limit = len(result) - offset
	}

	items = result[offset : offset+limit]

	return &models.PageableResponse{TotalCount: len(result), Items: items}, nil
}
