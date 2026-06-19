package object

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func GetResourceQuotas(cfg *rest.Config, namespace string) ([]corev1.ResourceQuota, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	list, err := client.CoreV1().ResourceQuotas(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetResourceQuota(cfg *rest.Config, namespace, name string) (*corev1.ResourceQuota, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ResourceQuotas(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func AddResourceQuota(cfg *rest.Config, rq *corev1.ResourceQuota) (*corev1.ResourceQuota, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ResourceQuotas(rq.Namespace).Create(context.Background(), rq, metav1.CreateOptions{})
}

func UpdateResourceQuota(cfg *rest.Config, rq *corev1.ResourceQuota) (*corev1.ResourceQuota, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ResourceQuotas(rq.Namespace).Update(context.Background(), rq, metav1.UpdateOptions{})
}

func DeleteResourceQuota(cfg *rest.Config, namespace, name string) error {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	return client.CoreV1().ResourceQuotas(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
