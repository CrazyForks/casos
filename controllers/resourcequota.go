package controllers

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/casosorg/casos/object"
)

type resourceQuotaSummary struct {
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	Hard            map[string]string `json:"hard"`
	Used            map[string]string `json:"used"`
	CreatedAt       string            `json:"createdAt"`
	ResourceVersion string            `json:"resourceVersion"`
}

func toResourceQuotaSummary(rq corev1.ResourceQuota) resourceQuotaSummary {
	hard := make(map[string]string, len(rq.Spec.Hard))
	for k, v := range rq.Spec.Hard {
		hard[string(k)] = v.String()
	}
	used := make(map[string]string, len(rq.Status.Used))
	for k, v := range rq.Status.Used {
		used[string(k)] = v.String()
	}
	return resourceQuotaSummary{
		Namespace:       rq.Namespace,
		Name:            rq.Name,
		Hard:            hard,
		Used:            used,
		CreatedAt:       rq.CreationTimestamp.UTC().Format("2006-01-02 15:04:05"),
		ResourceVersion: rq.ResourceVersion,
	}
}

type resourceQuotaRequest struct {
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	Hard            map[string]string `json:"hard"`
	ResourceVersion string            `json:"resourceVersion"`
}

func buildResourceList(hard map[string]string) (corev1.ResourceList, error) {
	rl := make(corev1.ResourceList, len(hard))
	for k, v := range hard {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return nil, err
		}
		rl[corev1.ResourceName(k)] = q
	}
	return rl, nil
}

// GetResourceQuotas
// @router /api/get-resourcequotas [get]
func (c *ApiController) GetResourceQuotas() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	namespace := c.GetString("namespace")
	rqs, err := object.GetResourceQuotas(cfg, namespace)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	result := make([]resourceQuotaSummary, 0, len(rqs))
	for _, rq := range rqs {
		result = append(result, toResourceQuotaSummary(rq))
	}
	c.ResponseOk(result)
}

// GetResourceQuota
// @router /api/get-resourcequota [get]
func (c *ApiController) GetResourceQuota() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	namespace := c.GetString("namespace")
	name := c.GetString("name")
	rq, err := object.GetResourceQuota(cfg, namespace, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toResourceQuotaSummary(*rq))
}

// AddResourceQuota
// @router /api/add-resourcequota [post]
func (c *ApiController) AddResourceQuota() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req resourceQuotaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	rl, err := buildResourceList(req.Hard)
	if err != nil {
		c.ResponseError("invalid resource quantity: " + err.Error())
		return
	}
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: rl,
		},
	}
	created, err := object.AddResourceQuota(cfg, rq)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toResourceQuotaSummary(*created))
}

// UpdateResourceQuota
// @router /api/update-resourcequota [post]
func (c *ApiController) UpdateResourceQuota() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req resourceQuotaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	rl, err := buildResourceList(req.Hard)
	if err != nil {
		c.ResponseError("invalid resource quantity: " + err.Error())
		return
	}
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:            req.Name,
			Namespace:       req.Namespace,
			ResourceVersion: req.ResourceVersion,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: rl,
		},
	}
	updated, err := object.UpdateResourceQuota(cfg, rq)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toResourceQuotaSummary(*updated))
}

// DeleteResourceQuota
// @router /api/delete-resourcequota [post]
func (c *ApiController) DeleteResourceQuota() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req resourceQuotaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if err := object.DeleteResourceQuota(cfg, req.Namespace, req.Name); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}
