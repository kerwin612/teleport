/*
Copyright 2020 Gravitational, Inc.

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

package proxy

import (
	"path"
	"strings"

	"github.com/gravitational/teleport/api/types"
	apievents "github.com/gravitational/teleport/api/types/events"
)

type apiResource struct {
	apiGroup        string
	apiGroupVersion string
	namespace       string
	resourceKind    string
	resourceName    string
	skipEvent       bool
}

// parseResourcePath does best-effort parsing of a Kubernetes API request path.
// All fields of the returned apiResource may be empty.
func parseResourcePath(p string) apiResource {
	// Kubernetes API reference: https://kubernetes.io/docs/reference/kubernetes-api/
	// Let's try to parse this. Here be dragons!
	//
	// URLs have a prefix that defines an "API group":
	// - /api/v1/ - the special "core" API group (e.g. pods, secrets, etc. belong here)
	// - /apis/{group}/{version} - the other properly named groups (e.g. apps/v1 or rbac.authorization.k8s.io/v1beta1)
	//
	// After the prefix, we have the resource info:
	// - /namespaces/{namespace}/{resource kind}/{resource name} for namespaced resources
	//   - turns out, namespace is optional when you query across all
	//     namespaces (e.g. /api/v1/pods to get pods in all namespaces)
	// - /{resource kind}/{resource name} for cluster-scoped resources (e.g. namespaces or nodes)
	//
	// If {resource name} is missing, the request refers to all resources of
	// that kind (e.g. list all pods).
	//
	// There can be more items after {resource name} (a "subresource"), like
	// pods/foo/exec, but the depth is arbitrary (e.g.
	// /api/v1/namespaces/{namespace}/pods/{name}/proxy/{path})
	//
	// And the cherry on top - watch endpoints, e.g.
	// /api/v1/watch/namespaces/{namespace}/pods/{name}
	// for live updates on resources (specific resources or all of one kind)
	var r apiResource

	// Clean up the path and make it absolute.
	p = path.Clean(p)
	if !path.IsAbs(p) {
		p = "/" + p
	}

	parts := strings.Split(p, "/")
	switch {
	// Core API group has a "special" URL prefix /api/v1/.
	case len(parts) >= 3 && parts[1] == "api" && parts[2] == "v1":
		r.apiGroup = "core"
		r.apiGroupVersion = parts[2]
		parts = parts[3:]
	// Other API groups have URL prefix /apis/{group}/{version}.
	case len(parts) >= 4 && parts[1] == "apis":
		r.apiGroup, r.apiGroupVersion = parts[2], parts[3]
		parts = parts[4:]
	case len(parts) >= 2 && (parts[1] == "api" || parts[1] == "apis"):
		// /api or /apis.
		// This is part of API discovery. Don't emit to audit log to reduce
		// noise.
		r.skipEvent = true
		return r
	default:
		// Doesn't look like a k8s API path, return empty result.
		return r
	}

	// Watch API endpoints have an extra /watch/ prefix. For now, silently
	// strip it from our result.
	if len(parts) > 0 && parts[0] == "watch" {
		parts = parts[1:]
	}

	switch len(parts) {
	case 0:
		// e.g. /apis/apps/v1
		// This is part of API discovery. Don't emit to audit log to reduce
		// noise.
		r.skipEvent = true
		return r
	case 1:
		// e.g. /api/v1/pods - list pods in all namespaces
		r.resourceKind = parts[0]
	case 2:
		// e.g. /api/v1/clusterroles/{name} - read a cluster-level resource
		r.resourceKind = parts[0]
		r.resourceName = parts[1]
	case 3:
		if parts[0] == "namespaces" {
			// e.g. /api/v1/namespaces/{namespace}/pods - list pods in a
			// specific namespace
			r.namespace = parts[1]
			r.resourceKind = parts[2]
		} else {
			// e.g. /apis/apiregistration.k8s.io/v1/apiservices/{name}/status
			kind := append([]string{parts[0]}, parts[2:]...)
			r.resourceKind = strings.Join(kind, "/")
			r.resourceName = parts[1]
		}
	default:
		// e.g. /api/v1/namespaces/{namespace}/pods/{name} - get a specific pod
		// or /api/v1/namespaces/{namespace}/pods/{name}/exec - exec command in a pod
		if parts[0] == "namespaces" {
			r.namespace = parts[1]
			kind := append([]string{parts[2]}, parts[4:]...)
			r.resourceKind = strings.Join(kind, "/")
			r.resourceName = parts[3]
		} else {
			// e.g. /api/v1/nodes/{name}/proxy/{path}
			kind := append([]string{parts[0]}, parts[2:]...)
			r.resourceKind = strings.Join(kind, "/")
			r.resourceName = parts[1]
		}
	}
	return r
}

func (r apiResource) populateEvent(e *apievents.KubeRequest) {
	e.ResourceAPIGroup = path.Join(r.apiGroup, r.apiGroupVersion)
	e.ResourceNamespace = r.namespace
	e.ResourceKind = r.resourceKind
	e.ResourceName = r.resourceName
}

// allowedResourcesKey is a key used to identify a resource in the allowedResources map.
type allowedResourcesKey struct {
	apiGroup     string
	resourceKind string
}

// allowedResources is a map of supported resources and their corresponding
// teleport resource kind for the purpose of resource rbac.
var allowedResources = map[allowedResourcesKey]string{
	{apiGroup: "core", resourceKind: "pods"}:                                      types.KindKubePod,
	{apiGroup: "core", resourceKind: "secrets"}:                                   types.KindKubeSecret,
	{apiGroup: "core", resourceKind: "configmaps"}:                                types.KindKubeConfigmap,
	{apiGroup: "core", resourceKind: "namespaces"}:                                types.KindKubeNamespace,
	{apiGroup: "core", resourceKind: "services"}:                                  types.KindKubeService,
	{apiGroup: "core", resourceKind: "serviceaccounts"}:                           types.KindKubeServiceAccount,
	{apiGroup: "core", resourceKind: "nodes"}:                                     types.KindKubeNode,
	{apiGroup: "core", resourceKind: "persistentvolumes"}:                         types.KindKubePersistentVolume,
	{apiGroup: "core", resourceKind: "persistentvolumeclaims"}:                    types.KindKubePersistentVolumeClaim,
	{apiGroup: "apps", resourceKind: "deployments"}:                               types.KindKubeDeployment,
	{apiGroup: "apps", resourceKind: "replicasets"}:                               types.KindKubeReplicaSet,
	{apiGroup: "apps", resourceKind: "statefulsets"}:                              types.KindKubeStatefulset,
	{apiGroup: "apps", resourceKind: "daemonsets"}:                                types.KindKubeDaemonSet,
	{apiGroup: "rbac.authorization.k8s.io", resourceKind: "clusterroles"}:         types.KindKubeClusterRole,
	{apiGroup: "rbac.authorization.k8s.io", resourceKind: "roles"}:                types.KindKubeRole,
	{apiGroup: "rbac.authorization.k8s.io", resourceKind: "clusterrolebindings"}:  types.KindKubeClusterRoleBinding,
	{apiGroup: "rbac.authorization.k8s.io", resourceKind: "rolebindings"}:         types.KindKubeRoleBinding,
	{apiGroup: "batch", resourceKind: "cronjobs"}:                                 types.KindKubeCronjob,
	{apiGroup: "batch", resourceKind: "jobs"}:                                     types.KindKubeJob,
	{apiGroup: "certificates.k8s.io", resourceKind: "certificatesigningrequests"}: types.KindKubeCertificateSigningRequest,
	{apiGroup: "networking.k8s.io", resourceKind: "ingresses"}:                    types.KindKubeIngress,
}

// getKubeResourceAndAPIGroupFromType returns the Kubernetes resource kind and
// API group for a given Teleport resource kind. If the Teleport resource kind
// is not supported, it returns the Teleport resource kind as the Kubernetes
// resource kind and an empty string as the API group.
func getKubeResourceAndAPIGroupFromType(s string) (kind string, apiGroup string) {
	for k, v := range allowedResources {
		if v == s {
			apiGroup := ""
			if k.apiGroup != "core" {
				apiGroup = k.apiGroup
			}
			return k.resourceKind, apiGroup
		}
	}
	return s + "s", ""
}

// getResourceWithKey returns the teleport resource kind for a given resource key if
// it exists, otherwise returns an empty string.
func getResourceWithKey(k allowedResourcesKey) string {
	if k.apiGroup == "" {
		k.apiGroup = "core"
	}
	return allowedResources[k]
}

// getResourceFromRequest returns a KubernetesResource if the user tried to access
// a specific endpoint that Teleport support resource filtering. Otherwise, returns nil.
func getResourceFromRequest(requestURI string) (*types.KubernetesResource, apiResource) {
	apiResource := parseResourcePath(requestURI)
	resourceType, ok := getTeleportResourceKindFromAPIResource(apiResource)
	// if the resource is not supported, return nil.
	// if the resource is supported but the resource name is not present, return nil because it's a list request.
	if !ok || apiResource.resourceName == "" {
		return nil, apiResource
	}
	return &types.KubernetesResource{
		Kind:      resourceType,
		Namespace: apiResource.namespace,
		Name:      apiResource.resourceName,
	}, apiResource
}

func getTeleportResourceKindFromAPIResource(r apiResource) (string, bool) {
	resource := getResourceFromAPIResource(r.resourceKind)
	resourceType, ok := allowedResources[allowedResourcesKey{apiGroup: r.apiGroup, resourceKind: resource}]
	return resourceType, ok
}

// getResourceFromAPIResource returns the resource kind from the api resource.
// If the resource kind contains sub resources (e.g. pods/exec), it returns the
// resource kind without the subresource.
func getResourceFromAPIResource(resourceKind string) string {
	if idx := strings.Index(resourceKind, "/"); idx != -1 {
		return resourceKind[:idx]
	}
	return resourceKind
}
