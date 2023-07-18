/*
Copyright 2023.

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
	"bytes"
	"context"
	"errors"
	"log"
	"strings"

	infrablockspacenetv1alpha1 "github.com/InfraBlockchain/infrablockspace-operator/api/v1alpha1"
	"github.com/InfraBlockchain/infrablockspace-operator/pkg/status"
	"github.com/tae2089/bob-logging/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InfraBlockSpaceReconciler reconciles a InfraBlockSpace object
type InfraBlockSpaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrablockspace.net,resources=infrablockspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrablockspace.net,resources=infrablockspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrablockspace.net,resources=infrablockspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/logs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InfraBlockSpace object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *InfraBlockSpaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	reqInfraBlockspace := &infrablockspacenetv1alpha1.InfraBlockSpace{}
	err := r.Get(ctx, req.NamespacedName, reqInfraBlockspace)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if reqInfraBlockspace.Status.SecretStatus != status.Success {
		err = r.checkSecretExists(ctx, reqInfraBlockspace)
		if err != nil {
			return r.updateStatusAndReturnError(ctx, reqInfraBlockspace, status.Failed, err)
		}
		return r.updateStatusAndReturnResult(ctx, reqInfraBlockspace, status.Success)
	}
	return ctrl.Result{}, nil
}

func (r *InfraBlockSpaceReconciler) updateStatusAndReturnError(ctx context.Context, reqInfraBlockspace *infrablockspacenetv1alpha1.InfraBlockSpace, status string, err error) (ctrl.Result, error) {
	reqInfraBlockspace.Status.SecretStatus = status
	if updateErr := r.Status().Update(ctx, reqInfraBlockspace); updateErr != nil {
		return ctrl.Result{}, errors.Join(err, updateErr)
	}
	return ctrl.Result{}, err
}
func (r *InfraBlockSpaceReconciler) updateStatusAndReturnResult(ctx context.Context, reqInfraBlockspace *infrablockspacenetv1alpha1.InfraBlockSpace, status string) (ctrl.Result, error) {
	reqInfraBlockspace.Status.SecretStatus = status
	if err := r.Status().Update(ctx, reqInfraBlockspace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}
func (r *InfraBlockSpaceReconciler) createSecret(ctx context.Context, namespace, name string, key infrablockspacenetv1alpha1.Key) error {
	if key.KeyType == "" || key.Scheme == "" || key.Seed == "" {
		return errors.New("key type, scheme and seed are required")
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			"type":   key.KeyType,
			"seed":   key.Seed,
			"scheme": key.Scheme,
		},
	}
	if err := r.Create(ctx, secret); err != nil {
		log.Println(err, "Faild to create Secret")
		return err
	}
	return nil
}

func (r *InfraBlockSpaceReconciler) updateSecret(ctx context.Context, namespace, name string, key infrablockspacenetv1alpha1.Key) error {
	if key.KeyType == "" || key.Scheme == "" || key.Seed == "" {
		return errors.New("key type, scheme and seed are required")
	}
	foundSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, foundSecret); err != nil {
		return err
	}
	if !(bytes.Equal(foundSecret.Data["type"], []byte(key.KeyType)) &&
		bytes.Equal(foundSecret.Data["seed"], []byte(key.Seed)) &&
		bytes.Equal(foundSecret.Data["scheme"], []byte(key.Scheme))) {
		foundSecret.Data["type"] = []byte(key.KeyType)
		foundSecret.Data["seed"] = []byte(key.Seed)
		foundSecret.Data["scheme"] = []byte(key.Scheme)
		if err := r.Update(ctx, foundSecret); err != nil {
			return err
		}
	}

	return nil
}

// func (r *InfraBlockSpaceReconciler) createStatefulset(ctx context.Context, reqInfraBlockspace *infrablockspacenetv1alpha1.InfraBlockSpace) error {

// 	secret := &corev1.Secret{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      reqInfraBlockspace.Name,
// 			Namespace: reqInfraBlockspace.Namespace,
// 		},
// 		Data: map[string][]byte{},
// 	}
// 	if err := r.Create(ctx, secret); err != nil {
// 		log.Println(err, "Faild to create Secret")
// 		return err
// 	}
// 	return nil
// }

// func (r *InfraBlockSpaceReconciler) createService(ctx context.Context, reqInfraBlockspace *infrablockspacenetv1alpha1.InfraBlockSpace) error {

// 	service := &corev1.Service{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      generateResoucreName(reqInfraBlockspace.Spec.Region, reqInfraBlockspace.Name),
// 			Namespace: reqInfraBlockspace.Namespace,
// 		},
// 		Spec: corev1.ServiceSpec{
// 			Type:     corev1.ServiceTypeClusterIP,
// 			Selector: map[string]string{},
// 			Ports: []corev1.ServicePort{
// 				{
// 					Protocol: corev1.ProtocolTCP,
// 					Port:     reqInfraBlockspace.Spec.Port.WSPort,
// 				},
// 				{
// 					Protocol: corev1.ProtocolTCP,
// 					Port:     reqInfraBlockspace.Spec.Port.RPCPort,
// 				},
// 				{
// 					Protocol: corev1.ProtocolTCP,
// 					Port:     reqInfraBlockspace.Spec.Port.P2PPort,
// 				},
// 			},
// 		},
// 	}

// 	if err := r.Create(ctx, service); err != nil {
// 		log.Println(err, "Faild to create Secret")
// 		return err
// 	}

// 	headless := &corev1.Service{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      reqInfraBlockspace.Name,
// 			Namespace: reqInfraBlockspace.Namespace,
// 		},
// 		Spec: corev1.ServiceSpec{
// 			Type:      corev1.ServiceTypeClusterIP,
// 			ClusterIP: corev1.ClusterIPNone,
// 		},
// 	}

// 	if err := r.Create(ctx, headless); err != nil {
// 		log.Println(err, "Faild to create Secret")
// 		return err
// 	}
// 	return nil
// }

func generateResoucreName(names ...string) string {
	newNameList := make([]string, 0)
	for _, name := range names {
		if name != "" {
			newNameList = append(newNameList, name)
		}
	}
	return strings.Join(newNameList, "-")
}

func (r *InfraBlockSpaceReconciler) checkSecretExists(ctx context.Context, reqInfraBlockspace *infrablockspacenetv1alpha1.InfraBlockSpace) error {

	for _, key := range *reqInfraBlockspace.Spec.Keys {
		secret := &corev1.Secret{}
		name := generateResoucreName(reqInfraBlockspace.Name, reqInfraBlockspace.Spec.Region, key.KeyType)
		isExists, err := r.checkResourceExists(ctx, reqInfraBlockspace.Namespace, name, secret)
		if isExists == false {
			if err != nil {
				logger.Error(err)
				return err
			} else {
				logger.Info("start to Create Secret")
				if err := r.createSecret(ctx, reqInfraBlockspace.Namespace, name, key); err != nil {
					logger.Error(err)
					return err
				}
				logger.Info("finish to Create Secret")
			}
		} else {
			logger.Info("start to update secret")
			if err = r.updateSecret(ctx, reqInfraBlockspace.Namespace, name, key); err != nil {
				logger.Error(err)
				return err
			}
			logger.Info("finish to update secret")
		}
	}

	return nil
}
func (r *InfraBlockSpaceReconciler) checkResourceExists(ctx context.Context, namespace string, name string, obj client.Object) (bool, error) {
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj); err != nil {
		if kerrors.IsNotFound(err) { // create
			return false, nil
		} else { // error
			return false, err
		}
	} else { // update
		return true, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InfraBlockSpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrablockspacenetv1alpha1.InfraBlockSpace{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
