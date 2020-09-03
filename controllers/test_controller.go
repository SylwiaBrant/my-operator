/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/types"

	testcomv1alpha1 "github.com/SylwiaBrant/test-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestReconciler reconciles a Test object
type TestReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=test.com.test.com,resources=tests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test.com.test.com,resources=tests/status,verbs=get;update;patch
//reconciler - compares provided state with actual cluster state and updates the cluster on finding state differences using a Client
func (reconciler *TestReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := reconciler.Log.WithValues("test", request.Name)

	testCR := &testcomv1alpha1.Test{}
	ctx := context.TODO()
	err := reconciler.Get(ctx, request.NamespacedName, testCR)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Custom Resource object not found. Ignoring request.")
			return ctrl.Result{}, nil
		}
		// Reconcile failed due to error - requeue
		reqLogger.Info("An error occured. Retrying...")
		return ctrl.Result{}, err
	}
	reqLogger.Info("Custom Resource created successfully.")

	testConfigMap := &corev1.ConfigMap{}
	err = reconciler.Get(context.TODO(), types.NamespacedName{Name: testCR.Name, Namespace: testCR.Namespace}, testConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			testConfigMap = reconciler.createConfigMap(testCR)
			if err = reconciler.Create(context.TODO(), testConfigMap); err != nil {
				reqLogger.Info("Failed to create new ConfigMap")
				return ctrl.Result{}, err
			}
			reqLogger.Info("Custom Resource object not found. Ignoring request")
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	testConfigMap.Data["message"] = testCR.Spec.Message
	err = reconciler.Update(context.TODO(), testConfigMap)
	reqLogger.Info("ConfigMap created successfully")

	return ctrl.Result{}, nil
}

func (reconciler *TestReconciler) createConfigMap(cr *testcomv1alpha1.Test) *corev1.ConfigMap {

	testConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"message": cr.Spec.Message,
		},
	}
	return testConfigMap
}

// specifies how the controller is built to watch a CR and other resources that are owned
// and managed by that controller
func (reconciler *TestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testcomv1alpha1.Test{}).
		Complete(reconciler)
}
