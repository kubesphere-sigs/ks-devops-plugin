/*
Copyright 2020 The KubeSphere Authors.

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
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/globalrole"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/globalrolebinding"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/user"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/workspacerole"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/workspacerolebinding"
	"errors"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	devopsv1alpha3 "devops.kubesphere.io/plugin/pkg/api/devops/v1alpha3"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/workspace"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/workspacetemplate"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"devops.kubesphere.io/plugin/pkg/api"
	"devops.kubesphere.io/plugin/pkg/apiserver/query"
	"devops.kubesphere.io/plugin/pkg/informers"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3"
	"devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/devops"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type ResourceGetter struct {
	clusterResourceGetters    map[schema.GroupVersionResource]v1alpha3.Interface
	namespacedResourceGetters map[schema.GroupVersionResource]v1alpha3.Interface
}

func NewResourceGetter(factory informers.InformerFactory, cache cache.Cache) *ResourceGetter {
	namespacedResourceGetters := make(map[schema.GroupVersionResource]v1alpha3.Interface)
	clusterResourceGetters := make(map[schema.GroupVersionResource]v1alpha3.Interface)

	// kubesphere resources
	clusterResourceGetters[devopsv1alpha3.GroupVersion.WithResource(devopsv1alpha3.ResourcePluralDevOpsProject)] = devops.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha1.ResourcePluralWorkspace)] = workspace.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[tenantv1alpha1.SchemeGroupVersion.WithResource(tenantv1alpha2.ResourcePluralWorkspaceTemplate)] = workspacetemplate.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralGlobalRole)] = globalrole.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralWorkspaceRole)] = workspacerole.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralUser)] = user.New(factory.KubeSphereSharedInformerFactory(), factory.KubernetesSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralGlobalRoleBinding)] = globalrolebinding.New(factory.KubeSphereSharedInformerFactory())
	clusterResourceGetters[iamv1alpha2.SchemeGroupVersion.WithResource(iamv1alpha2.ResourcesPluralWorkspaceRoleBinding)] = workspacerolebinding.New(factory.KubeSphereSharedInformerFactory())
	return &ResourceGetter{
		namespacedResourceGetters: namespacedResourceGetters,
		clusterResourceGetters:    clusterResourceGetters,
	}
}

// TryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *ResourceGetter) TryResource(clusterScope bool, resource string) v1alpha3.Interface {
	if clusterScope {
		for k, v := range r.clusterResourceGetters {
			if k.Resource == resource {
				return v
			}
		}
	}
	for k, v := range r.namespacedResourceGetters {
		if k.Resource == resource {
			return v
		}
	}
	return nil
}

func (r *ResourceGetter) Get(resource, namespace, name string) (runtime.Object, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(namespace, name)
}

func (r *ResourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.List(namespace, query)
}
