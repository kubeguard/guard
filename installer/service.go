package installer

import (
	"net"
	"strconv"

	"github.com/appscode/guard/server"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newService(namespace, addr string) (runtime.Object, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errors.Wrap(err, "Guard server address is invalid.")
	}
	svcPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.Wrap(err, "Guard server port is invalid.")
	}

	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: core.ServiceSpec{
			Type:      core.ServiceTypeClusterIP,
			ClusterIP: host,
			Ports: []core.ServicePort{
				{
					Name:       "api",
					Port:       int32(svcPort),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromInt(server.ServingPort),
				},
			},
			Selector: labels,
		},
	}, nil
}
