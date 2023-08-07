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
	"context"
	"errors"
	infrablockspacenetv1alpha1 "github.com/InfraBlockchain/infrablockspace-operator/api/v1alpha1"
	"github.com/InfraBlockchain/infrablockspace-operator/pkg/chain"
	"github.com/InfraBlockchain/infrablockspace-operator/pkg/util"
	"github.com/tae2089/bob-logging/logger"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
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
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

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

	reqInfraBlockSpace := &infrablockspacenetv1alpha1.InfraBlockSpace{}
	err := r.Get(ctx, req.NamespacedName, reqInfraBlockSpace)
	if err != nil {
		logger.Error(err)
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	err = r.ensureChainSecrets(ctx, reqInfraBlockSpace)

	if err != nil {
		return ctrl.Result{}, err
	}

	result, err := r.ensureChainPVC(ctx, reqInfraBlockSpace)
	if err != nil || result.Requeue {
		return result, err
	}

	_, err = r.ensureService(ctx, reqInfraBlockSpace)
	if err != nil {
		return result, err
	}

	result, err = r.ensureStatefulSet(ctx, reqInfraBlockSpace)
	if err != nil || result.Requeue {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *InfraBlockSpaceReconciler) ensureChainSecrets(ctx context.Context, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) error {

	for _, key := range *reqInfraBlockSpace.Spec.Keys {
		secret := &corev1.Secret{}
		name := util.GenerateResourceName(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Region, reqInfraBlockSpace.Spec.Rack, key.KeyType)
		isExists, err := r.checkResourceExists(ctx, reqInfraBlockSpace.Namespace, name, secret)
		if err != nil {
			return err
		}
		if !(isExists) {
			if err := r.createSecret(ctx, reqInfraBlockSpace.Namespace, name, key); err != nil {
				return err
			}
		} else {
			if err = r.updateSecret(ctx, reqInfraBlockSpace.Namespace, name, key); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *InfraBlockSpaceReconciler) ensureChainPVC(ctx context.Context, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) (ctrl.Result, error) {
	name := util.GenerateResourceName(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Region, reqInfraBlockSpace.Spec.Rack, chain.SuffixPvc)
	isExists, err := r.checkResourceExists(ctx, reqInfraBlockSpace.Namespace, name, &corev1.PersistentVolumeClaim{})
	if !(isExists) {
		if err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
		// create
		return r.createChainPVC(ctx, name, reqInfraBlockSpace.Namespace, reqInfraBlockSpace.Spec.Size, reqInfraBlockSpace.Spec.StorageClassName)
	} else {
		// update
		return r.updateChainPVC(ctx, name, reqInfraBlockSpace.Namespace, reqInfraBlockSpace.Spec.Size)
	}

}
func (r *InfraBlockSpaceReconciler) checkResourceExists(ctx context.Context, namespace string, name string, obj client.Object) (bool, error) {
	if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj); err != nil {
		logger.Error(err)
		if kerrors.IsNotFound(err) { // create
			return false, nil
		} else { // error
			return false, err
		}
	} else { // update
		return true, nil
	}
}

func (r *InfraBlockSpaceReconciler) createSecret(ctx context.Context, namespace, name string, key chain.Key) error {

	if err := r.validateKey(key); err != nil {
		return err
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
		logger.Error(err)
		return err
	}
	logger.Info("created secrets", zapcore.Field{
		Key:    "key",
		Type:   zapcore.StringType,
		String: key.KeyType,
	})
	return nil
}

func (r *InfraBlockSpaceReconciler) updateSecret(ctx context.Context, namespace, name string, key chain.Key) error {
	if err := r.validateKey(key); err != nil {
		logger.Error(err)
		return err
	}

	foundSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, foundSecret); err != nil {
		logger.Error(err)
		return err
	}
	existingKeyType := util.DecodingBase64(foundSecret.Data["type"])
	existingKeySeed := util.DecodingBase64(foundSecret.Data["seed"])
	existingKeyScheme := util.DecodingBase64(foundSecret.Data["scheme"])

	if !(existingKeyType == key.KeyType &&
		existingKeySeed == key.Seed &&
		existingKeyScheme == key.Scheme) {
		foundSecret.StringData = make(map[string]string)
		foundSecret.StringData["type"] = key.KeyType
		foundSecret.StringData["seed"] = key.Seed
		foundSecret.StringData["scheme"] = key.Scheme
		if err := r.Update(ctx, foundSecret); err != nil {
			logger.Error(err)
			return err
		}
		logger.Info("updated secrets", zapcore.Field{
			Key:    "key",
			Type:   zapcore.StringType,
			String: key.KeyType,
		})
	}

	return nil
}

func (r *InfraBlockSpaceReconciler) validateKey(key chain.Key) error {
	if key.KeyType == "" || key.Scheme == "" || key.Seed == "" {
		err := errors.New("key type, scheme and seed are required")
		logger.Error(err)
		return err
	}
	return nil
}

func (r *InfraBlockSpaceReconciler) createChainPVC(ctx context.Context, name, namespace, size, scName string) (ctrl.Result, error) {
	if size == "" {
		size = chain.VolumeSize100Gi
	}
	pvc := chain.CreateChainPVC(name, namespace, size, scName)
	if err := r.Create(ctx, pvc); err != nil {
		logger.Error(err)
		return ctrl.Result{}, err
	}
	logger.Info("created pvc", zapcore.Field{
		Key:    "Name",
		Type:   zapcore.StringType,
		String: name,
	})
	return ctrl.Result{Requeue: true}, nil
}
func (r *InfraBlockSpaceReconciler) updateChainPVC(ctx context.Context, name, namespace, size string) (ctrl.Result, error) {
	if size == "" {
		return ctrl.Result{}, errors.New("size is required")
	}
	foundPVC := &corev1.PersistentVolumeClaim{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, foundPVC); err != nil {
		logger.Error(err)
		return ctrl.Result{}, err
	}
	if !chain.IsSamePvcSize(*foundPVC.Spec.Resources.Requests.Storage(), resource.MustParse(size)) {
		*foundPVC.Spec.Resources.Requests.Storage() = resource.MustParse(size)
		if err := r.Update(ctx, foundPVC); err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
	}
	logger.Info("updated pvc", zapcore.Field{
		Key:    "Name",
		Type:   zapcore.StringType,
		String: name,
	})
	return ctrl.Result{}, nil
}
func (r *InfraBlockSpaceReconciler) ensureStatefulSet(ctx context.Context, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) (ctrl.Result, error) {
	name := util.GenerateResourceName(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Region, reqInfraBlockSpace.Spec.Rack)
	isExists, err := r.checkResourceExists(ctx, reqInfraBlockSpace.Namespace, name, &corev1.PersistentVolumeClaim{})
	if err != nil {
		logger.Error(err)
		return ctrl.Result{}, err
	}
	if !(isExists) {
		return r.createStatefulSet(ctx, name, reqInfraBlockSpace)
	} else {
		return r.updateStatefulSet(ctx, name, reqInfraBlockSpace)
	}
}

func (r *InfraBlockSpaceReconciler) createStatefulSet(ctx context.Context, name string, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) (ctrl.Result, error) {

	initContainers := r.getInitContainers(reqInfraBlockSpace)
	mainContainers := r.getMainContainers(reqInfraBlockSpace)
	volumes := r.getVolumes(reqInfraBlockSpace)

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: reqInfraBlockSpace.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: name + "-" + chain.SuffixHeadlessService,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers:     mainContainers,
					Volumes:        volumes,
				},
			},
		},
	}

	if reqInfraBlockSpace.Spec.Resources != nil {
		statefulSet.Spec.Template.Spec.Containers[0].Resources = *reqInfraBlockSpace.Spec.Resources
	}

	if reqInfraBlockSpace.Spec.ReadinessProbe != nil {
		statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe = reqInfraBlockSpace.Spec.ReadinessProbe
	}

	if reqInfraBlockSpace.Spec.LivenessProbe != nil {
		statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe = reqInfraBlockSpace.Spec.LivenessProbe
	}

	if reqInfraBlockSpace.Spec.Lifecycle != nil {
		statefulSet.Spec.Template.Spec.Containers[0].Lifecycle = reqInfraBlockSpace.Spec.Lifecycle
	}

	if err := r.Create(ctx, statefulSet); err != nil {
		logger.Error(err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *InfraBlockSpaceReconciler) updateStatefulSet(ctx context.Context, name string, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) (ctrl.Result, error) {
	foundStatefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: reqInfraBlockSpace.Namespace}, foundStatefulSet); err != nil {
		logger.Error(err)
		return ctrl.Result{}, err
	}

	if foundStatefulSet.Spec.Replicas != pointer.Int32(reqInfraBlockSpace.Spec.Replicas) {
		foundStatefulSet.Spec.Replicas = pointer.Int32(reqInfraBlockSpace.Spec.Replicas)
		if err := r.Update(ctx, foundStatefulSet); err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if foundStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe != reqInfraBlockSpace.Spec.ReadinessProbe {
		foundStatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe = reqInfraBlockSpace.Spec.ReadinessProbe
		if err := r.Update(ctx, foundStatefulSet); err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if foundStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe != reqInfraBlockSpace.Spec.LivenessProbe {
		foundStatefulSet.Spec.Template.Spec.Containers[0].LivenessProbe = reqInfraBlockSpace.Spec.LivenessProbe
		if err := r.Update(ctx, foundStatefulSet); err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if foundStatefulSet.Spec.Template.Spec.Containers[0].Lifecycle != reqInfraBlockSpace.Spec.Lifecycle {
		foundStatefulSet.Spec.Template.Spec.Containers[0].Lifecycle = reqInfraBlockSpace.Spec.Lifecycle
		if err := r.Update(ctx, foundStatefulSet); err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func (r *InfraBlockSpaceReconciler) ensureService(ctx context.Context, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) (ctrl.Result, error) {
	name := util.GenerateResourceName(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Region, reqInfraBlockSpace.Spec.Rack)
	isExists, err := r.checkResourceExists(ctx, reqInfraBlockSpace.Namespace, name, &corev1.Service{})
	if err != nil {
		logger.Error(err)
		return ctrl.Result{}, err
	}
	if !(isExists) {
		if err := r.createServices(ctx, name, reqInfraBlockSpace); err != nil {
			logger.Error(err)
			return ctrl.Result{}, err
		}
	}
	err = r.updateServices(ctx, name, reqInfraBlockSpace)
	return ctrl.Result{}, nil
}

func (r *InfraBlockSpaceReconciler) createServices(ctx context.Context, name string, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) error {
	if err := r.createHeadlessService(ctx, name, reqInfraBlockSpace); err != nil {
		return err
	}
	if err := r.createHeadlessService(ctx, name, reqInfraBlockSpace); err != nil {
		return err
	}
	return nil
}

func (r *InfraBlockSpaceReconciler) createHeadlessService(ctx context.Context, name string, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) error {
	ports := chain.GetServicePorts(reqInfraBlockSpace)
	servicePorts := chain.GenerateServicePorts(ports...)
	service := chain.GenerateHeadlessServiceObject(name+"-"+chain.SuffixHeadlessService, reqInfraBlockSpace.Namespace, servicePorts, nil)
	err := r.createService(ctx, service, reqInfraBlockSpace)
	return err
}

func (r *InfraBlockSpaceReconciler) createClusterIPService(ctx context.Context, name string, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) error {
	ports := chain.GetServicePorts(reqInfraBlockSpace)
	servicePorts := chain.GenerateServicePorts(ports...)
	service := chain.GenerateClusterIpServiceObject(name+"-"+chain.SuffixService, reqInfraBlockSpace.Namespace, servicePorts, nil)
	err := r.createService(ctx, service, reqInfraBlockSpace)
	return err
}

func (r *InfraBlockSpaceReconciler) createService(ctx context.Context, service *corev1.Service, reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) error {
	if err := r.Create(ctx, service); err != nil {
		return err
	}
	_ = ctrl.SetControllerReference(reqInfraBlockSpace, service, r.Scheme)
	logger.Info("created service", zapcore.Field{
		Key:    "Name",
		Type:   zapcore.StringType,
		String: service.Name,
	})
	return nil
}

func (r *InfraBlockSpaceReconciler) getMainContainers(reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) []corev1.Container {
	isBootNode := reqInfraBlockSpace.Spec.BootNodes == nil
	args := chain.GetRelayChainArgs(reqInfraBlockSpace.Spec.Port, isBootNode, reqInfraBlockSpace.Spec.BootNodes)
	chainContainer := chain.CreateChainContainer(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Image, nil, args, nil)
	return []corev1.Container{chainContainer}
}

func (r *InfraBlockSpaceReconciler) getInitContainers(reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) []corev1.Container {
	downloadRelayChainSpecContainer := r.createDownloadChainSpecInitContainer(reqInfraBlockSpace.Spec.ChainSpec)
	injectKeysContainer := r.createInjectKeysInitContainer(reqInfraBlockSpace)
	return []corev1.Container{downloadRelayChainSpecContainer, injectKeysContainer}
}

func (r *InfraBlockSpaceReconciler) getVolumes(reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) []corev1.Volume {
	var volumes []corev1.Volume
	secretVolumes := chain.GetSecretVolumes(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Region, reqInfraBlockSpace.Spec.Rack, reqInfraBlockSpace.Spec.Keys)
	pvcVolumes := chain.GetPvcVolumes(reqInfraBlockSpace.Name, reqInfraBlockSpace.Spec.Region, reqInfraBlockSpace.Spec.Rack, chain.RelayChain)
	volumes = append(volumes, secretVolumes...)
	volumes = append(volumes, pvcVolumes...)
	return volumes
}

func (r *InfraBlockSpaceReconciler) createDownloadChainSpecInitContainer(chainSpecUrl string) corev1.Container {
	commands := chain.GetDownloadSpecCommand(chainSpecUrl, chain.RelayChainSpecFileName)
	volumeMounts := chain.CreateChainSpecVolumeMount()
	return chain.CreateInitContainer(chain.DownloadRelayChainSpecContainer, chain.DownloadChainSpecImage, commands, volumeMounts)
}

func (r *InfraBlockSpaceReconciler) createInjectKeysInitContainer(reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) corev1.Container {
	commands, args := chain.GetInjectKeyCommandAndArgs(reqInfraBlockSpace.Spec.Keys)
	volumeMounts := chain.CreateKeyStoreVolumeMount(reqInfraBlockSpace.Spec.Keys)
	container := chain.CreateChainContainer(chain.InjectKeysContainer, reqInfraBlockSpace.Spec.Image, commands, args, volumeMounts)
	return container
}

// SetupWithManager sets up the controller with the Manager.
func (r *InfraBlockSpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrablockspacenetv1alpha1.InfraBlockSpace{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
