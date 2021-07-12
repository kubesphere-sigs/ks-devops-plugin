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

package tenant

import (
	"devops.kubesphere.io/plugin/pkg/models/iam/am"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"

	"devops.kubesphere.io/plugin/pkg/api"
	"devops.kubesphere.io/plugin/pkg/apiserver/authorization/authorizer"
	"devops.kubesphere.io/plugin/pkg/apiserver/query"
	kubesphere "devops.kubesphere.io/plugin/pkg/client/clientset/versioned"
	"devops.kubesphere.io/plugin/pkg/informers"
	resourcesv1alpha3 "devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/resource"
	resourcev1alpha3 "devops.kubesphere.io/plugin/pkg/models/resources/v1alpha3/resource"
)

const orphanFinalizer = "orphan.finalizers.kubesphere.io"

type Interface interface {
	ListDevOpsProjects(user user.Info, workspace string, query *query.Query) (*api.ListResult, error)
}

type tenantOperator struct {
	am             am.AccessManagementInterface
	authorizer     authorizer.Authorizer
	k8sclient      kubernetes.Interface
	ksclient       kubesphere.Interface
	resourceGetter *resourcesv1alpha3.ResourceGetter
}

func New(informers informers.InformerFactory, k8sclient kubernetes.Interface,
	ksclient kubesphere.Interface, authorizer authorizer.Authorizer,
	resourceGetter *resourcev1alpha3.ResourceGetter) Interface {
	return &tenantOperator{
		authorizer:     authorizer,
		resourceGetter: resourcesv1alpha3.NewResourceGetter(informers, nil),
		k8sclient:      k8sclient,
		ksclient:       ksclient,
	}
}

func contains(objects []runtime.Object, object runtime.Object) bool {
	for _, item := range objects {
		if item == object {
			return true
		}
	}
	return false
}
