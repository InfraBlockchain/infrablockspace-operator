package chain

import (
	"fmt"
	"github.com/InfraBlockchain/infrablockspace-operator/pkg/render"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestRenderInjectKeyTemplate(t *testing.T) {

	keys := []Key{
		{KeyType: "gran", Scheme: "test2", Seed: "test2"},
		{KeyType: "test", Scheme: "test", Seed: "test"},
	}
	t.Run("create key", func(t *testing.T) {
		got := render.RenderingInTemplate(InjectKeyScript, keys)
		fmt.Print(got)
	})
}

func TestRenderInjectKeyRange(t *testing.T) {

	keys := []*Key{
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
