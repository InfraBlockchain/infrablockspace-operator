package chain

import (
	infrablockspacenetv1alpha1 "github.com/InfraBlockchain/infrablockspace-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetServicePorts(reqInfraBlockSpace *infrablockspacenetv1alpha1.InfraBlockSpace) []int32 {
	var ports []int32
	if reqInfraBlockSpace.Spec.Port.WSPort == 0 {
		ports = getDefaultRelayPorts()
	} else {
		ports = getCustomRelayPorts(reqInfraBlockSpace.Spec.Port)
	}
	return ports
}

func getDefaultRelayPorts() []int32 {
	return []int32{DefaultChainWSPort, DefaultChainRPCPort, DefaultChainP2PPort}
}

func getCustomRelayPorts(port Port) []int32 {
	return []int32{port.WSPort, port.RPCPort, port.P2PPort}
}

func GenerateClusterIpServiceObject(name, namespace string, ports []corev1.ServicePort, selector map[string]string) *corev1.Service {
	clusterIPService := generateServiceObject(name, namespace, corev1.ServiceTypeClusterIP, ports, selector)
	return clusterIPService
}

func GenerateHeadlessServiceObject(name, namespace string, ports []corev1.ServicePort, selector map[string]string) *corev1.Service {
	headlessService := generateServiceObject(name, namespace, corev1.ServiceTypeClusterIP, ports, selector)
	headlessService.Spec.ClusterIP = corev1.ClusterIPNone
	return headlessService
}

func GenerateServicePorts(ports ...int32) []corev1.ServicePort {
	servicePorts := make([]corev1.ServicePort, len(ports))
	for _, port := range ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Port:       port,
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: port},
			Protocol:   corev1.ProtocolTCP,
		})
	}
	return servicePorts
}

func generateServiceObject(name, namespace string, serviceType corev1.ServiceType, ports []corev1.ServicePort, selector map[string]string) *corev1.Service {
	return &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Ports:    ports,
			Selector: selector,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
