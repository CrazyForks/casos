package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

type HelmCompatibilityErrorCode string

const (
	HelmCompatibilityInvalidManifest   HelmCompatibilityErrorCode = "INVALID_MANIFEST"
	HelmCompatibilityResourceNotServed HelmCompatibilityErrorCode = "RESOURCE_NOT_SERVED"
	HelmCompatibilityResourceNotFound  HelmCompatibilityErrorCode = "RESOURCE_NOT_FOUND"
	HelmCompatibilityForbiddenByRBAC   HelmCompatibilityErrorCode = "FORBIDDEN_BY_RBAC"
	HelmCompatibilityConflict          HelmCompatibilityErrorCode = "CONFLICT"
	HelmCompatibilityDiscoveryFailed   HelmCompatibilityErrorCode = "DISCOVERY_FAILED"
	HelmCompatibilityAPIUnavailable    HelmCompatibilityErrorCode = "API_UNAVAILABLE"
	HelmCompatibilityDryRunFailed      HelmCompatibilityErrorCode = "HELM_DRY_RUN_FAILED"
)

type HelmCompatibilityErrorInfo struct {
	Code HelmCompatibilityErrorCode `json:"code"`
	GVK  string                     `json:"gvk,omitempty"`
}

type HelmCompatibilityError struct {
	Code HelmCompatibilityErrorCode
	GVK  schema.GroupVersionKind
	Err  error
}

func (e *HelmCompatibilityError) Error() string {
	if e.GVK.Empty() {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("%s for %s: %v", e.Code, e.GVK.String(), e.Err)
}

func (e *HelmCompatibilityError) Unwrap() error {
	return e.Err
}

func HelmCompatibilityErrorInfoFrom(err error) (HelmCompatibilityErrorInfo, bool) {
	var compatibilityErr *HelmCompatibilityError
	if !errors.As(err, &compatibilityErr) {
		return HelmCompatibilityErrorInfo{}, false
	}
	info := HelmCompatibilityErrorInfo{Code: compatibilityErr.Code}
	if !compatibilityErr.GVK.Empty() {
		info.GVK = compatibilityErr.GVK.String()
	}
	return info, true
}

// kubernetesResourceAdapter validates rendered Helm resources against the
// target cluster. Helm remains responsible for release dry-runs and writes.
type kubernetesResourceAdapter struct {
	mapper  meta.RESTMapper
	refresh func()
	getCRD  func(context.Context, string) (*unstructured.Unstructured, error)
}

func newKubernetesResourceAdapter(getter action.RESTClientGetter) (*kubernetesResourceAdapter, error) {
	if getter == nil {
		return nil, &HelmCompatibilityError{
			Code: HelmCompatibilityAPIUnavailable,
			Err:  fmt.Errorf("Helm REST client getter is missing"),
		}
	}
	discoveryClient, err := getter.ToDiscoveryClient()
	if err != nil {
		return nil, classifyKubernetesResourceError(schema.GroupVersionKind{}, err)
	}
	restConfig, err := getter.ToRESTConfig()
	if err != nil {
		return nil, classifyKubernetesResourceError(schema.GroupVersionKind{}, err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, classifyKubernetesResourceError(schema.GroupVersionKind{}, err)
	}
	crds := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	})
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	return &kubernetesResourceAdapter{
		mapper:  mapper,
		refresh: mapper.Reset,
		getCRD: func(ctx context.Context, name string) (*unstructured.Unstructured, error) {
			return crds.Get(ctx, name, metav1.GetOptions{})
		},
	}, nil
}

func (a *kubernetesResourceAdapter) validateManifest(ctx context.Context, manifest string, chartCRDs map[schema.GroupVersionKind]crdDefinition) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if a == nil || a.mapper == nil {
		return &HelmCompatibilityError{
			Code: HelmCompatibilityAPIUnavailable,
			Err:  fmt.Errorf("Kubernetes REST mapper is missing"),
		}
	}

	decoder := utilyaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)
	for {
		if err := ctx.Err(); err != nil {
			return classifyKubernetesResourceError(schema.GroupVersionKind{}, err)
		}
		var document map[string]interface{}
		if err := decoder.Decode(&document); err != nil {
			if err == io.EOF {
				break
			}
			return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("decode rendered Helm manifest: %w", err)}
		}
		if len(document) == 0 {
			continue
		}
		if err := a.validateDocument(ctx, document, chartCRDs); err != nil {
			return err
		}
	}
	return nil
}

func (a *kubernetesResourceAdapter) validateDocument(ctx context.Context, document map[string]interface{}, chartCRDs map[schema.GroupVersionKind]crdDefinition) error {
	if err := ctx.Err(); err != nil {
		return classifyKubernetesResourceError(schema.GroupVersionKind{}, err)
	}
	kind, _, err := unstructured.NestedString(document, "kind")
	if err != nil {
		return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("read kind: %w", err)}
	}
	kind = strings.TrimSpace(kind)
	if kind == "List" {
		items, _, err := unstructured.NestedSlice(document, "items")
		if err != nil {
			return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("decode List items: %w", err)}
		}
		for _, item := range items {
			object, ok := item.(map[string]interface{})
			if !ok {
				return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("List item is not a Kubernetes object")}
			}
			if err := a.validateDocument(ctx, object, chartCRDs); err != nil {
				return err
			}
		}
		return nil
	}

	apiVersion, _, err := unstructured.NestedString(document, "apiVersion")
	if err != nil {
		return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("read apiVersion: %w", err)}
	}
	apiVersion = strings.TrimSpace(apiVersion)
	if apiVersion == "" || kind == "" {
		return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("apiVersion and kind are required")}
	}
	groupVersion, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return &HelmCompatibilityError{Code: HelmCompatibilityInvalidManifest, Err: fmt.Errorf("parse apiVersion %q: %w", apiVersion, err)}
	}
	gvk := groupVersion.WithKind(kind)

	_, err = a.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if meta.IsNoMatchError(err) && a.refresh != nil {
		a.refresh()
		if err := ctx.Err(); err != nil {
			return classifyKubernetesResourceError(gvk, err)
		}
		_, err = a.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}
	if err != nil {
		if meta.IsNoMatchError(err) {
			definition, declaredByChart := chartCRDs[gvk]
			if !declaredByChart {
				return &HelmCompatibilityError{Code: HelmCompatibilityResourceNotServed, GVK: gvk, Err: err}
			}
			return a.validateChartCRD(ctx, gvk, definition)
		}
		return classifyKubernetesResourceError(gvk, err)
	}
	return nil
}

func (a *kubernetesResourceAdapter) validateChartCRD(ctx context.Context, gvk schema.GroupVersionKind, definition crdDefinition) error {
	if a.getCRD == nil {
		return &HelmCompatibilityError{
			Code: HelmCompatibilityAPIUnavailable,
			GVK:  gvk,
			Err:  fmt.Errorf("CustomResourceDefinition client is missing"),
		}
	}
	existing, err := a.getCRD(ctx, definition.Name)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return classifyKubernetesResourceError(gvk, err)
	}
	return validateExistingCRD(gvk, existing)
}

func validateExistingCRD(gvk schema.GroupVersionKind, crd *unstructured.Unstructured) error {
	if crd == nil {
		return crdDoesNotServeError(gvk, "existing CustomResourceDefinition is empty")
	}
	definition, err := parseCRDDefinition(crd.Object)
	if err != nil {
		return &HelmCompatibilityError{
			Code: HelmCompatibilityInvalidManifest,
			GVK:  gvk,
			Err:  fmt.Errorf("parse existing CRD %s: %w", crd.GetName(), err),
		}
	}
	if definition.Group != gvk.Group {
		return crdDoesNotServeError(gvk, "existing CustomResourceDefinition has group %q", definition.Group)
	}
	if definition.Kind != gvk.Kind {
		return crdDoesNotServeError(gvk, "existing CustomResourceDefinition has kind %q", definition.Kind)
	}
	if !slices.Contains(definition.ServedVersions, gvk.Version) {
		return crdDoesNotServeError(gvk, "existing CustomResourceDefinition does not serve version %s", gvk.Version)
	}
	return nil
}

func crdDoesNotServeError(gvk schema.GroupVersionKind, format string, args ...interface{}) error {
	return &HelmCompatibilityError{
		Code: HelmCompatibilityResourceNotServed,
		GVK:  gvk,
		Err:  fmt.Errorf(format, args...),
	}
}

func classifyKubernetesResourceError(gvk schema.GroupVersionKind, err error) error {
	return classifyKubernetesErrorWithDefault(gvk, err, HelmCompatibilityAPIUnavailable)
}

func classifyHelmDryRunError(gvk schema.GroupVersionKind, err error) error {
	if err != nil && isUnregisteredKindDryRunError(err) {
		return &HelmCompatibilityError{Code: HelmCompatibilityResourceNotServed, GVK: gvk, Err: err}
	}
	return classifyKubernetesErrorWithDefault(gvk, err, HelmCompatibilityDryRunFailed)
}

func classifyKubernetesErrorWithDefault(gvk schema.GroupVersionKind, err error, defaultCode HelmCompatibilityErrorCode) error {
	if err == nil {
		return nil
	}
	code := defaultCode
	switch {
	case meta.IsNoMatchError(err):
		code = HelmCompatibilityResourceNotServed
	case apierrors.IsForbidden(err), apierrors.IsUnauthorized(err):
		code = HelmCompatibilityForbiddenByRBAC
	case apierrors.IsAlreadyExists(err), apierrors.IsConflict(err):
		code = HelmCompatibilityConflict
	case apierrors.IsNotFound(err):
		code = HelmCompatibilityResourceNotFound
	case apierrors.IsTimeout(err), apierrors.IsServerTimeout(err), apierrors.IsServiceUnavailable(err), apierrors.IsInternalError(err), apierrors.IsTooManyRequests(err), errors.Is(err, io.EOF), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		code = HelmCompatibilityAPIUnavailable
	default:
		var networkError net.Error
		if errors.As(err, &networkError) {
			code = HelmCompatibilityAPIUnavailable
		}
		var discoveryError *discovery.ErrGroupDiscoveryFailed
		if errors.As(err, &discoveryError) {
			code = HelmCompatibilityDiscoveryFailed
		}
	}
	return &HelmCompatibilityError{Code: code, GVK: gvk, Err: err}
}
