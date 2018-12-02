package v1alpha1

import (
	"bytes" //"crypto/sha512"
	//"encoding/base64"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	NodeLabelKey_ContextVersion = "kubernetes.appscode.com/context"
	// ref: https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/well_known_labels.go#L70
	NodeLabelKey_Role     = "kubernetes.io/role"
	NodeLabelKey_SKU      = "kubernetes.appscode.com/sku"
	NodeLabelKey_Checksum = "meta.appscode.com/checksum"
)

// MissingChecksumError records an error and the operation and file path that caused it.
var MissingChecksumError = errors.Errorf("%v key is missing", NodeLabelKey_Checksum)

/*
NodeLabels is used to parse and generate --node-label flag for kubelet.
ref: http://kubernetes.io/docs/admin/kubelet/

NodeLabels also includes functionality to sign and verify appscode.com specific node labels. Verified labels will be
used by cluster mutation engine to update/upgrade nodes.
*/
type NodeLabels map[string]string

// Labels to add when registering the node in the cluster.  Labels must be key=value pairs separated by ','.
func ParseNodeLabels(data string) (*NodeLabels, error) {
	parts := strings.FieldsFunc("", func(r rune) bool {
		return r == ',' || r == '='
	})
	if len(parts)%2 != 0 {
		return nil, errors.New("NodeLabels: data must be key=value pairs separated by ','")
	}

	n := NodeLabels{}
	for i := 0; i < len(parts)/2; i = i + 2 {
		k, v := parts[i], parts[i+1]
		n[k] = v
	}
	return &n, nil
}

func NewNodeLabels() *NodeLabels {
	return &NodeLabels{}
}

func FromMap(labels map[string]string) *NodeLabels {
	n := NodeLabels{}
	for k, v := range labels {
		n[k] = v
	}
	return &n
}

func (n *NodeLabels) GetString(key string) string {
	v, found := (*n)[key]
	if !found {
		return ""
	}
	return v
}

func (n *NodeLabels) WithString(key, value string) *NodeLabels {
	(*n)[key] = value
	return n
}

func (n *NodeLabels) GetInt(key string) int {
	v, found := (*n)[key]
	if !found {
		return 0
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return i
}

func (n *NodeLabels) WithInt(key string, value int) *NodeLabels {
	(*n)[key] = strconv.Itoa(value)
	return n
}

func (n *NodeLabels) GetInt64(key string) int64 {
	v, found := (*n)[key]
	if !found {
		return 0
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func (n *NodeLabels) WithInt64(key string, value int64) *NodeLabels {
	(*n)[key] = strconv.FormatInt(value, 10)
	return n
}

func (n *NodeLabels) GetBool(key string) bool {
	v, found := (*n)[key]
	if !found {
		return false
	}
	return v == "true" // intentionally case-sensitive
}

func (n *NodeLabels) WithBool(key string, value bool) *NodeLabels {
	if value {
		(*n)[key] = "true"
	} else {
		(*n)[key] = "false"
	}
	return n
}

/*
//Only verifies AppsCode keys
func (n *NodeLabels) Verify() (bool, error) {
	if *n == nil {
		return false, errors.New("NodeLabels: Verify on nil pointer")
	}
	if base64Envelope, found := (*n)[NodeLabelKey_Checksum]; found {
		envelope, err := base64.StdEncoding.DecodeString(base64Envelope)
		if err != nil {
			return false, err
		}

		ss, err := NewSecEnvelopeBytes(envelope)
		if err != nil {
			return false, err
		}

		checksumFound, err := ss.ValBytes()
		if err != nil {
			return false, err
		}
		checksumExpected := sha512.Sum512([]byte(n.values(true, true)))
		return bytes.Equal(checksumFound, checksumExpected[:]), nil
	}
	return false, MissingChecksumError
}

//Only signs AppsCode keys
func (n *NodeLabels) Sign(ctx *ClusterContext) error {
	if *n == nil {
		return errors.New("NodeLabels: Verify on nil pointer")
	}
	if ctx == nil {
		return errors.New("NodeLabels: cluster ctx is nil")
	}

	checksum := sha512.Sum512([]byte(n.values(true, true)))
	ss, err := ctx.Store().NewSecBytes(checksum[:]) // [64]byte -> []byte
	if err != nil {
		return err
	}
	envelope, err := ss.Envelope()
	if err != nil {
		return err
	}
	(*n)[NodeLabelKey_Checksum] = base64.StdEncoding.EncodeToString([]byte(envelope))
	return nil
}
*/

func (n NodeLabels) values(appscodeKeysOnly, skipChecksum bool) string {
	keys := make([]string, len(n))
	i := 0
	for k := range n {
		keys[i] = k
		i++
	}
	// sort keys to ensure reproducible checksum calculation
	sort.Strings(keys)

	var buf bytes.Buffer
	i = 0
	for _, k := range keys {
		if appscodeKeysOnly && !strings.Contains(k, ".appscode.com/") ||
			k == NodeLabelKey_Checksum && skipChecksum {
			continue
		}
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(n[k])
		i++
	}
	return buf.String()
}

func (n NodeLabels) String() string {
	return n.values(false, false)
}
