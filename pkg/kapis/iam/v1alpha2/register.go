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

	"devops.kubesphere.io/plugin/pkg/apiserver/authorization/authorizer"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"devops.kubesphere.io/plugin/pkg/api"
	"devops.kubesphere.io/plugin/pkg/apiserver/runtime"
	"devops.kubesphere.io/plugin/pkg/constants"
	"devops.kubesphere.io/plugin/pkg/models/iam/am"
	"devops.kubesphere.io/plugin/pkg/models/iam/group"
	"devops.kubesphere.io/plugin/pkg/models/iam/im"
	"devops.kubesphere.io/plugin/pkg/server/errors"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface, am am.AccessManagementInterface, group group.GroupOperator, authorizer authorizer.Authorizer) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newIAMHandler(im, am, group, authorizer)

	ws.Route(ws.GET("/devops/{devops}/members").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.GET("/devops/{devops}/members/{member}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve devops project member details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.POST("/devops/{devops}/members").
		To(handler.CreateNamespaceMembers).
		Doc("Add members to the DevOps project in bulk.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.PUT("/devops/{devops}/members/{member}").
		To(handler.UpdateNamespaceMember).
		Doc("Update the role bind of the member.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.DELETE("/devops/{devops}/members/{member}").
		To(handler.RemoveNamespaceMember).
		Doc("Delete a member from the DevOps project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))

	// roles
	ws.Route(ws.POST("/devops/{devops}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified devops project.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.DELETE("/devops/{devops}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.PUT("/devops/{devops}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update devops project role.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.PATCH("/devops/{devops}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch devops project role.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve devops project role details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))

	ws.Route(ws.GET("/devops/{devops}/members/{member}/roles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))

	container.Add(ws)
	return nil
}
