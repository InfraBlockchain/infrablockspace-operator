package render

import (
	"fmt"
	"testing"

	"github.com/InfraBlockchain/infrablockspace-operator/pkg/chain"
	corev1 "k8s.io/api/core/v1"
)

func TestRenderInjectKeyTemplate(t *testing.T) {

	keys := []chain.Key{
		{KeyType: "gran", Scheme: "test2", Seed: "test2"},
		{KeyType: "test", Scheme: "test", Seed: "test"},
	}
	t.Run("create key", func(t *testing.T) {
		got := RenderingInTemplate(keys)
		fmt.Print(got)
	})
}

func TestRenderInjectKeyRange(t *testing.T) {

	keys := []*chain.Key{
		{KeyType: "gran", Scheme: "test2", Seed: "test2"},
		{KeyType: "test", Scheme: "test", Seed: "test"},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "chain-keystore",
			MountPath: "/keystore",
		},
	}
	for _, key := range keys {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      key.KeyType,
			MountPath: fmt.Sprintf("/var/run/secrets/%s", key.KeyType),
		})
	}
	fmt.Println(volumeMounts)
}
