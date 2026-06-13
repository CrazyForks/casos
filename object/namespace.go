package object

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func GetNamespaces(cfg *rest.Config) ([]corev1.Namespace, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetNamespace(cfg *rest.Config, name string) (*corev1.Namespace, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
}

func AddNamespace(cfg *rest.Config, ns *corev1.Namespace) (*corev1.Namespace, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
}

func UpdateNamespace(cfg *rest.Config, ns *corev1.Namespace) (*corev1.Namespace, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Namespaces().Update(context.Background(), ns, metav1.UpdateOptions{})
}

func DeleteNamespace(cfg *rest.Config, name string) error {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	return client.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
}

// ForceDeleteNamespace clears finalizers on a stuck Terminating namespace so
// the API server can complete the deletion.
func ForceDeleteNamespace(cfg *rest.Config, name string) error {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	// First ensure the delete is initiated (no-op if already deleting).
	_ = client.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})

	// Patch spec.finalizers to [] so the namespace controller unblocks.
	patch, _ := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{"finalizers": []string{}},
	})
	_, err = client.CoreV1().Namespaces().Finalize(context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       corev1.NamespaceSpec{Finalizers: []corev1.FinalizerName{}},
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		// Fall back to a merge-patch on the metadata finalizers.
		metaPatch, _ := json.Marshal(map[string]interface{}{
			"metadata": map[string]interface{}{"finalizers": []string{}},
		})
		_, err = client.CoreV1().Namespaces().Patch(
			context.Background(), name,
			types.MergePatchType, metaPatch,
			metav1.PatchOptions{},
		)
		_ = patch
	}
	return err
}
