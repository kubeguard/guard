package installer

import (
	"bytes"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

var labels = map[string]string{
	"app": "guard",
}

func Generate(opts Options) ([]byte, error) {
	var objects []runtime.Object

	if opts.namespace != metav1.NamespaceSystem && opts.namespace != metav1.NamespaceDefault {
		objects = append(objects, newNamespace(opts.namespace))
	}
	objects = append(objects, newServiceAccount(opts.namespace))
	objects = append(objects, newClusterRole(opts.namespace))
	objects = append(objects, newClusterRoleBinding(opts.namespace))
	if deployObjects, err := newDeployment(opts); err != nil {
		return nil, err
	} else {
		objects = append(objects, deployObjects...)
	}
	if svc, err := newService(opts.namespace, opts.addr); err != nil {
		return nil, err
	} else {
		objects = append(objects, svc)
	}

	mediaType := "application/yaml"
	info, ok := runtime.SerializerInfoForMediaType(clientsetscheme.Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, errors.Errorf("unsupported media type %q", mediaType)
	}
	codec := versioning.NewCodecForScheme(clientsetscheme.Scheme, info.Serializer, info.Serializer, nil, nil)

	var buf bytes.Buffer
	for i, obj := range objects {
		if i > 0 {
			buf.WriteString("---\n")
		}
		err := codec.Encode(obj, &buf)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
