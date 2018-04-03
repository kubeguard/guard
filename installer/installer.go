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

	if opts.Namespace != metav1.NamespaceSystem && opts.Namespace != metav1.NamespaceDefault {
		objects = append(objects, newNamespace(opts.Namespace))
	}
	objects = append(objects, newServiceAccount(opts.Namespace))
	objects = append(objects, newClusterRole(opts.Namespace))
	objects = append(objects, newClusterRoleBinding(opts.Namespace))
	if deployObjects, err := newDeployment(opts); err != nil {
		return nil, err
	} else {
		objects = append(objects, deployObjects...)
	}
	if svc, err := newService(opts.Namespace, opts.Addr); err != nil {
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
