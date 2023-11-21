package chain

import (
	"fmt"
	"github.com/InfraBlockchain/infrablockspace-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// CreateChainSpecVolumeMount creates a volume mount for the chain spec
func CreateChainSpecVolumeMount() []corev1.VolumeMount {
	return []corev1.VolumeMount{generateVolumeMount(Spec, SpecMountPath)}
}

// CreateKeyStoreVolumeMount creates a volume mount for the key store
func CreateKeyStoreVolumeMount(keys []Key) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{generateVolumeMount(KeyStore, KeyStoreMountPath)}
	for _, key := range keys {
		mountPath := fmt.Sprintf("/var/run/secrets/%s", key.KeyType)
		volumeMounts = append(volumeMounts, generateVolumeMount(key.KeyType, mountPath))
	}
	return volumeMounts
}

// generateVolumeMount generates a volume mount for the given name and mount path
func generateVolumeMount(name, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}
}

func GetSecretVolumes(name, region, rack string, keys []Key) []corev1.Volume {
	volumes := make([]corev1.Volume, len(keys))
	for _, key := range keys {
		secretName := util.GenerateResourceName(name, region, rack, key.KeyType)
		secretVolume := getSecretVolume(key.KeyType, secretName)
		volumes = append(volumes, secretVolume)
	}
	return volumes
}

func GetPvcVolumes(name, region, rack string, chainTypes ...ChainType) []corev1.Volume {
	volumes := make([]corev1.Volume, len(chainTypes))
	for i, mode := range chainTypes {
		pvcName := util.GenerateResourceName(name, region, rack, string(mode))
		pvcVolume := getPvcVolume(string(mode), pvcName)
		volumes[i] = pvcVolume
	}
	return volumes
}

func getSecretVolume(name, secretName string) corev1.Volume {
	secretVolume := getVolume()
	secretVolume.Name = name
	secretVolume.Secret = &corev1.SecretVolumeSource{
		SecretName: secretName,
	}
	return secretVolume
}

func GetEmptyDir(name string) corev1.Volume {
	emptyDir := getVolume()
	emptyDir.Name = name
	emptyDir.EmptyDir = &corev1.EmptyDirVolumeSource{}
	return emptyDir
}

func getPvcVolume(name, claimName string) corev1.Volume {
	pvcVolume := getVolume()
	pvcVolume.Name = name + "-" + SuffixPvc
	pvcVolume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: claimName,
	}
	return pvcVolume
}

func getVolume() corev1.Volume {
	return corev1.Volume{}
}
