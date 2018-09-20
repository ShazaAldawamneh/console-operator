package stub

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/console-operator/pkg/apis/console/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/console-operator/pkg/operator"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
)

func NewHandler() sdk.Handler {
	// andy bootstrapping? if so, we can do it here.
	return &handler{}
}

type handler struct {
}

func (h *handler) Handle(_ context.Context, _ sdk.Event) error {

	// checking a few things to test if deleted now, see isDeleted
	// if event.Deleted {
	//	// Garbage collector cleans up most via ownersRefs
	//	operator.DeleteOauthClient()
	//	return nil
	// }

	cr, err := getCR()
	if err != nil {
		return err
	}
	if isDeleted(cr, err) {
		logrus.Info("console has been deleted.")
		// Garbage collector cleans up most via ownersRefs
		operator.DeleteOauthClient()
		return nil
	}

	if isPaused(cr) {
		logrus.Info("console has been paused.")
		return nil
	}

	logrus.Info("Time to do real things now!  Nothing is deleted, nothing is paused....")
	return nil
}

func getCR() (*v1alpha1.Console, error) {
	namespace, _ := k8sutil.GetWatchNamespace()
	cr := &v1alpha1.Console{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "Console",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.OpenShiftConsoleName,
			Namespace: namespace,
		},
	}
	err := sdk.Get(cr)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get console custom resource: %s", err)
		}
		return nil, nil
	}
	return cr, nil
}

func isDeleted(object runtime.Object, err error) bool {
	// 404 deleted
	if errors.IsNotFound(err) {
		return true
	}
	// this is an error on the get request
	if err != nil {
		return false
	}
	// in process of being deleted
	obj, err := meta.Accessor(object)
	if obj.GetDeletionTimestamp() != nil {
		return false
	}
	return false
}

func isPaused(console *v1alpha1.Console) bool {
	if console.Spec.Paused {
		return true
	}
	return false
}
