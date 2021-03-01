package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-logr/logr"
	pkgErr "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const kymaSystem = "kyma-system"

func main() {
	log := ctrlzap.New()
	if err := mainErr(log); err != nil {
		log.Error(err, "while running test")
		os.Exit(1)
	}
}

func mainErr(log logr.Logger) error {
	clientset, err := getClientSet(nil)
	if err != nil {
		return err
	}
	impersonatedClientSet, err := getClientSet(&rest.ImpersonationConfig{
		UserName: "system:serviceaccount:default:test-deny",
	})
	if err != nil {
		return err
	}

	// create ns in which we forbid action via webhook
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: kymaSystem}},
		metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	if err := test(impersonatedClientSet); err != nil {
		return err
	}

	log.Info("success")
	return nil
}

func test(clientset *kubernetes.Clientset) error {
	// test creation of namespaced resource
	_, err := clientset.CoreV1().ConfigMaps(kymaSystem).Create(
		context.Background(),
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "denied-cm"},
			Data:       map[string]string{"doesnt": "matter"}},
		metav1.CreateOptions{},
	)
	if testErr := handleWebhookDenial(err); testErr != nil {
		return pkgErr.Wrap(testErr, "while creating configmap 'denied-cm' in namespace 'kyma-system'")
	}

	// test creation of clusterwide resource
	_, err = clientset.RbacV1().ClusterRoles().Create(
		context.Background(),
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: "denied-clusterrole"},
			Rules: []rbacv1.PolicyRule{{
				Verbs:     []string{rbacv1.VerbAll},
				APIGroups: []string{rbacv1.GroupName},
				Resources: []string{rbacv1.ResourceAll},
			}},
		},
		metav1.CreateOptions{},
	)
	if testErr := handleWebhookDenial(err); testErr != nil {
		return pkgErr.Wrap(testErr, "while creating clusterrole 'denied-clusterrole' in clusterwide scope")
	}
	return nil
}

func handleWebhookDenial(err error) error {
	if err == nil {
		return errors.New("webhook should have denied this request")
	}
	status, interalErr := getMetaV1Status(err)
	if interalErr != nil {
		err = pkgErr.Wrap(interalErr, err.Error())
		return err
	}
	if status.Code != http.StatusForbidden {
		return fmt.Errorf("expected to get %d, got %d, error obj: %s", http.StatusForbidden, status.Code, status.String())
	}

	admissionWebhookStr := "admission webhook"
	if !strings.Contains(status.Message, admissionWebhookStr) {
		return fmt.Errorf("error message is expected to contain %q string; error: %s", admissionWebhookStr, status.String())
	}

	return nil
}

func getMetaV1Status(err error) (metav1.Status, error) {
	if status := apierrors.APIStatus(nil); errors.As(err, &status) {
		return status.Status(), nil
	}
	return metav1.Status{}, errors.New(fmt.Sprintf("failed to convert err to metav1.Status; original err: %s", err))
}

func getClientSet(impersonateCfg *rest.ImpersonationConfig) (*kubernetes.Clientset, error) {
	config := ctrl.GetConfigOrDie()
	if impersonateCfg != nil {
		config.Impersonate = *impersonateCfg
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
