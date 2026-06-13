package object

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func GetServiceAccounts(cfg *rest.Config, namespace string) ([]corev1.ServiceAccount, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	list, err := client.CoreV1().ServiceAccounts(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetServiceAccount(cfg *rest.Config, namespace, name string) (*corev1.ServiceAccount, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ServiceAccounts(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func AddServiceAccount(cfg *rest.Config, sa *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ServiceAccounts(sa.Namespace).Create(context.Background(), sa, metav1.CreateOptions{})
}

func UpdateServiceAccount(cfg *rest.Config, sa *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().ServiceAccounts(sa.Namespace).Update(context.Background(), sa, metav1.UpdateOptions{})
}

func DeleteServiceAccount(cfg *rest.Config, namespace, name string) error {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	return client.CoreV1().ServiceAccounts(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
