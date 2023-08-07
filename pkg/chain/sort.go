package chain

import corev1 "k8s.io/api/core/v1"

type ServicePortSort []corev1.ServicePort

func (s ServicePortSort) Len() int {
	return len(s)
}

func (s ServicePortSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ServicePortSort) Less(i, j int) bool {
	return s[i].Port < s[j].Port
}
