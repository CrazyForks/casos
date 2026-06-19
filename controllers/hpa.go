package controllers

import (
	"encoding/json"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/casosorg/casos/object"
)

type hpaSummary struct {
	Namespace            string `json:"namespace"`
	Name                 string `json:"name"`
	ScaleTargetKind      string `json:"scaleTargetKind"`
	ScaleTargetName      string `json:"scaleTargetName"`
	ScaleTargetRef       string `json:"scaleTargetRef"`
	MinReplicas          int32  `json:"minReplicas"`
	MaxReplicas          int32  `json:"maxReplicas"`
	CurrentReplicas      int32  `json:"currentReplicas"`
	DesiredReplicas      int32  `json:"desiredReplicas"`
	CPUTargetUtilization *int32 `json:"cpuTargetUtilization"`
	CreatedAt            string `json:"createdAt"`
	ResourceVersion      string `json:"resourceVersion"`
}

func toHpaSummary(hpa autoscalingv2.HorizontalPodAutoscaler) hpaSummary {
	minReplicas := int32(1)
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	var cpuTarget *int32
	for _, m := range hpa.Spec.Metrics {
		if m.Type == autoscalingv2.ResourceMetricSourceType && m.Resource != nil &&
			m.Resource.Name == corev1.ResourceCPU &&
			m.Resource.Target.Type == autoscalingv2.UtilizationMetricType {
			cpuTarget = m.Resource.Target.AverageUtilization
			break
		}
	}

	return hpaSummary{
		Namespace:            hpa.Namespace,
		Name:                 hpa.Name,
		ScaleTargetKind:      hpa.Spec.ScaleTargetRef.Kind,
		ScaleTargetName:      hpa.Spec.ScaleTargetRef.Name,
		ScaleTargetRef:       fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name),
		MinReplicas:          minReplicas,
		MaxReplicas:          hpa.Spec.MaxReplicas,
		CurrentReplicas:      hpa.Status.CurrentReplicas,
		DesiredReplicas:      hpa.Status.DesiredReplicas,
		CPUTargetUtilization: cpuTarget,
		CreatedAt:            hpa.CreationTimestamp.UTC().Format("2006-01-02 15:04:05"),
		ResourceVersion:      hpa.ResourceVersion,
	}
}

// GetHPAs
// @router /api/get-hpas [get]
func (c *ApiController) GetHPAs() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	namespace := c.GetString("namespace")
	hpas, err := object.GetHPAs(cfg, namespace)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	result := make([]hpaSummary, 0, len(hpas))
	for _, hpa := range hpas {
		result = append(result, toHpaSummary(hpa))
	}
	c.ResponseOk(result)
}

// GetHPA
// @router /api/get-hpa [get]
func (c *ApiController) GetHPA() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	namespace := c.GetString("namespace")
	name := c.GetString("name")
	hpa, err := object.GetHPA(cfg, namespace, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toHpaSummary(*hpa))
}

type hpaRequest struct {
	Namespace            string `json:"namespace"`
	Name                 string `json:"name"`
	ScaleTargetKind      string `json:"scaleTargetKind"`
	ScaleTargetName      string `json:"scaleTargetName"`
	MinReplicas          int32  `json:"minReplicas"`
	MaxReplicas          int32  `json:"maxReplicas"`
	CPUTargetUtilization *int32 `json:"cpuTargetUtilization"`
	ResourceVersion      string `json:"resourceVersion"`
}

func buildHpaObject(req hpaRequest) *autoscalingv2.HorizontalPodAutoscaler {
	minReplicas := req.MinReplicas
	if minReplicas < 1 {
		minReplicas = 1
	}

	metrics := []autoscalingv2.MetricSpec{}
	if req.CPUTargetUtilization != nil {
		util := *req.CPUTargetUtilization
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: &util,
				},
			},
		})
	}

	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            req.Name,
			Namespace:       req.Namespace,
			ResourceVersion: req.ResourceVersion,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       req.ScaleTargetKind,
				Name:       req.ScaleTargetName,
			},
			MinReplicas: &minReplicas,
			MaxReplicas: req.MaxReplicas,
			Metrics:     metrics,
		},
	}
}

// AddHPA
// @router /api/add-hpa [post]
func (c *ApiController) AddHPA() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req hpaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	hpa := buildHpaObject(req)
	created, err := object.AddHPA(cfg, hpa)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toHpaSummary(*created))
}

// UpdateHPA
// @router /api/update-hpa [post]
func (c *ApiController) UpdateHPA() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req hpaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	hpa := buildHpaObject(req)
	updated, err := object.UpdateHPA(cfg, hpa)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(toHpaSummary(*updated))
}

// DeleteHPA
// @router /api/delete-hpa [post]
func (c *ApiController) DeleteHPA() {
	cfg := getAdminRestConfig()
	if cfg == nil {
		c.ResponseError("apiserver not ready")
		return
	}
	var req hpaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError("invalid request body: " + err.Error())
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if err := object.DeleteHPA(cfg, req.Namespace, req.Name); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}
