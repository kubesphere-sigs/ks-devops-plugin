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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	scheme "devops.kubesphere.io/plugin/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "kubesphere.io/api/tenant/v1alpha1"
)

// WorkspacesGetter has a method to return a WorkspaceInterface.
// A group's client should implement this interface.
type WorkspacesGetter interface {
	Workspaces() WorkspaceInterface
}

// WorkspaceInterface has methods to work with Workspace resources.
type WorkspaceInterface interface {
	Create(ctx context.Context, workspace *v1alpha1.Workspace, opts v1.CreateOptions) (*v1alpha1.Workspace, error)
	Update(ctx context.Context, workspace *v1alpha1.Workspace, opts v1.UpdateOptions) (*v1alpha1.Workspace, error)
	UpdateStatus(ctx context.Context, workspace *v1alpha1.Workspace, opts v1.UpdateOptions) (*v1alpha1.Workspace, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Workspace, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.WorkspaceList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Workspace, err error)
	WorkspaceExpansion
}

// workspaces implements WorkspaceInterface
type workspaces struct {
	client rest.Interface
}

// newWorkspaces returns a Workspaces
func newWorkspaces(c *TenantV1alpha1Client) *workspaces {
	return &workspaces{
		client: c.RESTClient(),
	}
}

// Get takes name of the workspace, and returns the corresponding workspace object, and an error if there is any.
func (c *workspaces) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Workspace, err error) {
	result = &v1alpha1.Workspace{}
	err = c.client.Get().
		Resource("workspaces").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Workspaces that match those selectors.
func (c *workspaces) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.WorkspaceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.WorkspaceList{}
	err = c.client.Get().
		Resource("workspaces").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested workspaces.
func (c *workspaces) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("workspaces").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a workspace and creates it.  Returns the server's representation of the workspace, and an error, if there is any.
func (c *workspaces) Create(ctx context.Context, workspace *v1alpha1.Workspace, opts v1.CreateOptions) (result *v1alpha1.Workspace, err error) {
	result = &v1alpha1.Workspace{}
	err = c.client.Post().
		Resource("workspaces").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(workspace).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a workspace and updates it. Returns the server's representation of the workspace, and an error, if there is any.
func (c *workspaces) Update(ctx context.Context, workspace *v1alpha1.Workspace, opts v1.UpdateOptions) (result *v1alpha1.Workspace, err error) {
	result = &v1alpha1.Workspace{}
	err = c.client.Put().
		Resource("workspaces").
		Name(workspace.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(workspace).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *workspaces) UpdateStatus(ctx context.Context, workspace *v1alpha1.Workspace, opts v1.UpdateOptions) (result *v1alpha1.Workspace, err error) {
	result = &v1alpha1.Workspace{}
	err = c.client.Put().
		Resource("workspaces").
		Name(workspace.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(workspace).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the workspace and deletes it. Returns an error if one occurs.
func (c *workspaces) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("workspaces").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *workspaces) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("workspaces").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched workspace.
func (c *workspaces) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Workspace, err error) {
	result = &v1alpha1.Workspace{}
	err = c.client.Patch(pt).
		Resource("workspaces").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
