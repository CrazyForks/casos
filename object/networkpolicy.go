package object

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func GetNetworkPolicies(cfg *rest.Config, namespace string) ([]networkingv1.NetworkPolicy, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}
	list, err := client.NetworkingV1().NetworkPolicies(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetNetworkPolicy(cfg *rest.Config, namespace, name string) (*networkingv1.NetworkPolicy, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.NetworkingV1().NetworkPolicies(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func AddNetworkPolicy(cfg *rest.Config, np *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.NetworkingV1().NetworkPolicies(np.Namespace).Create(context.Background(), np, metav1.CreateOptions{})
}

func UpdateNetworkPolicy(cfg *rest.Config, np *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.NetworkingV1().NetworkPolicies(np.Namespace).Update(context.Background(), np, metav1.UpdateOptions{})
}

func DeleteNetworkPolicy(cfg *rest.Config, namespace, name string) error {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	return client.NetworkingV1().NetworkPolicies(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
