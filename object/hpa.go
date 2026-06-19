package object

import (
	"context"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func GetHPAs(cfg *rest.Config, namespace string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	list, err := client.AutoscalingV2().HorizontalPodAutoscalers(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetHPA(cfg *rest.Config, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func AddHPA(cfg *rest.Config, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Create(context.Background(), hpa, metav1.CreateOptions{})
}

func UpdateHPA(cfg *rest.Config, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.AutoscalingV2().HorizontalPodAutoscalers(hpa.Namespace).Update(context.Background(), hpa, metav1.UpdateOptions{})
}

func DeleteHPA(cfg *rest.Config, namespace, name string) error {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
