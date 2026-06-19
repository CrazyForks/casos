package controllers

import (
	"encoding/json"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/casosorg/casos/object"
)

type networkPolicySummary struct {
	Namespace       string   `json:"namespace"`
	Name            string   `json:"name"`
	PodSelector     string   `json:"podSelector"`
	PolicyTypes     []string `json:"policyTypes"`
	IngressRules    int      `json:"ingressRules"`
	EgressRules     int      `json:"egressRules"`
	CreatedAt       string   `json:"createdAt"`
	ResourceVersion string   `json:"resourceVersion"`
}

func toNetworkPolicySummary(np networkingv1.NetworkPolicy) networkPolicySummary {
	types := make([]string, 0, len(np.Spec.PolicyTypes))
	for _, t := range np.Spec.PolicyTypes {
		types = append(types, string(t))
	}
	podSel := ""
	if b, err := json.Marshal(np.Spec.PodSelector.MatchLabels); err == nil {
		podSel = string(b)
	}
	return networkPolicySummary{
		Namespace:       np.Namespace,
		Name:            np.Name,
		PodSelector:     podSel,
		PolicyTypes:     types,
		IngressRules:    len(np.Spec.Ingress),
		EgressRules:     len(np.Spec.Egress),
		CreatedAt:       np.CreationTimestamp.UTC().Format("2006-01-02 15:04:05"),
		ResourceVersion: np.ResourceVersion,
	}
}

// GetNetworkPolicies
// @router /api/get-networkpolicies [get]
func (c *ApiController) GetNetworkPolicies() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	namespace := c.GetString("namespace")
	nps, err := object.GetNetworkPolicies(cfg, namespace)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	result := make([]networkPolicySummary, 0, len(nps))
	for _, np := range nps {
		result = append(result, toNetworkPolicySummary(np))
	}
	c.ResponseOk(result)
}

// GetNetworkPolicy
// @router /api/get-networkpolicy [get]
func (c *ApiController) GetNetworkPolicy() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	namespace := c.GetString("namespace")
	name := c.GetString("name")
	np, err := object.GetNetworkPolicy(cfg, namespace, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toNetworkPolicySummary(*np))
}

type networkPolicyRequest struct {
	Namespace         string                                  `json:"namespace"`
	Name              string                                  `json:"name"`
	PodSelectorLabels map[string]string                       `json:"podSelectorLabels"`
	PolicyTypes       []string                                `json:"policyTypes"`
	Ingress           []networkingv1.NetworkPolicyIngressRule `json:"ingress"`
	Egress            []networkingv1.NetworkPolicyEgressRule  `json:"egress"`
	ResourceVersion   string                                  `json:"resourceVersion"`
}

func buildNetworkPolicy(req networkPolicyRequest) *networkingv1.NetworkPolicy {
	types := make([]networkingv1.PolicyType, 0, len(req.PolicyTypes))
	for _, t := range req.PolicyTypes {
		types = append(types, networkingv1.PolicyType(t))
	}
	ingress := req.Ingress
	if ingress == nil {
		ingress = []networkingv1.NetworkPolicyIngressRule{}
	}
	egress := req.Egress
	if egress == nil {
		egress = []networkingv1.NetworkPolicyEgressRule{}
	}
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:            req.Name,
			Namespace:       req.Namespace,
			ResourceVersion: req.ResourceVersion,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: req.PodSelectorLabels,
			},
			PolicyTypes: types,
			Ingress:     ingress,
			Egress:      egress,
		},
	}
}

// AddNetworkPolicy
// @router /api/add-networkpolicy [post]
func (c *ApiController) AddNetworkPolicy() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req networkPolicyRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	np := buildNetworkPolicy(req)
	created, err := object.AddNetworkPolicy(cfg, np)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toNetworkPolicySummary(*created))
}

// UpdateNetworkPolicy
// @router /api/update-networkpolicy [post]
func (c *ApiController) UpdateNetworkPolicy() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req networkPolicyRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	np := buildNetworkPolicy(req)
	updated, err := object.UpdateNetworkPolicy(cfg, np)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toNetworkPolicySummary(*updated))
}

// DeleteNetworkPolicy
// @router /api/delete-networkpolicy [post]
func (c *ApiController) DeleteNetworkPolicy() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req networkPolicyRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if err := object.DeleteNetworkPolicy(cfg, req.Namespace, req.Name); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}
