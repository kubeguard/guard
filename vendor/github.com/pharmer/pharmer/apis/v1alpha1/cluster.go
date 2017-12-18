package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1alpha1"
)

const (
	ResourceCodeCluster = ""
	ResourceKindCluster = "Cluster"
	ResourceNameCluster = "cluster"
	ResourceTypeCluster = "clusters"
)

type LightsailCloudConfig struct {
	AccessKeyID     string `json:"accessKeyID,omitempty" protobuf:"bytes,1,opt,name=accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" protobuf:"bytes,2,opt,name=secretAccessKey"`
}

type VultrCloudConfig struct {
	Token string `json:"token,omitempty" protobuf:"bytes,1,opt,name=token"`
}

type LinodeCloudConfig struct {
	Token string `json:"token,omitempty" protobuf:"bytes,1,opt,name=token"`
	Zone  string `json:"zone,omitempty" protobuf:"bytes,2,opt,name=zone"`
}

type ScalewayCloudConfig struct {
	Organization string `json:"organization,omitempty" protobuf:"bytes,1,opt,name=organization"`
	Token        string `json:"token,omitempty" protobuf:"bytes,2,opt,name=token"`
	Region       string `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`
}

type PacketCloudConfig struct {
	Project string `json:"project,omitempty" protobuf:"bytes,1,opt,name=project"`
	ApiKey  string `json:"apiKey,omitempty" protobuf:"bytes,2,opt,name=apiKey"`
	Zone    string `json:"zone,omitempty" protobuf:"bytes,3,opt,name=zone"`
}

type SoftlayerCloudConfig struct {
	UserName string `json:"username,omitempty" protobuf:"bytes,1,opt,name=username"`
	ApiKey   string `json:"apiKey,omitempty" protobuf:"bytes,2,opt,name=apiKey"`
	Zone     string `json:"zone,omitempty" protobuf:"bytes,3,opt,name=zone"`
}

// ref: https://github.com/kubernetes/kubernetes/blob/8b9f0ea5de2083589f3b9b289b90273556bc09c4/pkg/cloudprovider/providers/azure/azure.go#L56
type AzureCloudConfig struct {
	TenantID           string `json:"tenantId,omitempty" protobuf:"bytes,1,opt,name=tenantId"`
	SubscriptionID     string `json:"subscriptionId,omitempty" protobuf:"bytes,2,opt,name=subscriptionId"`
	AadClientID        string `json:"aadClientId,omitempty" protobuf:"bytes,3,opt,name=aadClientId"`
	AadClientSecret    string `json:"aadClientSecret,omitempty" protobuf:"bytes,4,opt,name=aadClientSecret"`
	ResourceGroup      string `json:"resourceGroup,omitempty" protobuf:"bytes,5,opt,name=resourceGroup"`
	Location           string `json:"location,omitempty" protobuf:"bytes,6,opt,name=location"`
	SubnetName         string `json:"subnetName,omitempty" protobuf:"bytes,7,opt,name=subnetName"`
	SecurityGroupName  string `json:"securityGroupName,omitempty" protobuf:"bytes,8,opt,name=securityGroupName"`
	VnetName           string `json:"vnetName,omitempty" protobuf:"bytes,9,opt,name=vnetName"`
	RouteTableName     string `json:"routeTableName,omitempty" protobuf:"bytes,10,opt,name=routeTableName"`
	StorageAccountName string `json:"storageAccountName,omitempty" protobuf:"bytes,11,opt,name=storageAccountName"`
}

// ref: https://github.com/kubernetes/kubernetes/blob/8b9f0ea5de2083589f3b9b289b90273556bc09c4/pkg/cloudprovider/providers/gce/gce.go#L228
type GCECloudConfig struct {
	TokenURL           string   `gcfg:"token-url" ini:"token-url,omitempty" protobuf:"bytes,1,opt,name=tokenURL"`
	TokenBody          string   `gcfg:"token-body" ini:"token-body,omitempty" protobuf:"bytes,2,opt,name=tokenBody"`
	ProjectID          string   `gcfg:"project-id" ini:"project-id,omitempty" protobuf:"bytes,3,opt,name=projectID"`
	NetworkName        string   `gcfg:"network-name" ini:"network-name,omitempty" protobuf:"bytes,4,opt,name=networkName"`
	NodeTags           []string `gcfg:"node-tags" ini:"node-tags,omitempty,omitempty" protobuf:"bytes,5,rep,name=nodeTags"`
	NodeInstancePrefix string   `gcfg:"node-instance-prefix" ini:"node-instance-prefix,omitempty,omitempty" protobuf:"bytes,6,opt,name=nodeInstancePrefix"`
	Multizone          bool     `gcfg:"multizone" ini:"multizone,omitempty" protobuf:"varint,7,opt,name=multizone"`
}

type OVHCloudConfig struct {
	AuthUrl  string `gcfg:"auth-url" ini:"auth-url,omitempty" protobuf:"bytes,1,opt,name=authUrl"`
	Username string `gcfg:"username" ini:"username,omitempty" protobuf:"bytes,2,opt,name=username"`
	Password string `gcfg:"password" ini:"password,omitempty" protobuf:"bytes,3,opt,name=password"`
	TenantId string `gcfg:"tenant-id" ini:"tenant-id,omitempty" protobuf:"bytes,4,opt,name=tenantId"`
	Region   string `gcfg:"region" ini:"region,omitempty" protobuf:"bytes,5,opt,name=region"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cluster struct {
	metav1.TypeMeta   `json:",inline,omitempty,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ClusterSpec   `json:"spec,omitempty,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            ClusterStatus `json:"status,omitempty,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type Networking struct {
	NetworkProvider string `json:"networkProvider,omitempty" protobuf:"bytes,1,opt,name=networkProvider"` // kubenet, flannel, calico, opencontrail
	PodSubnet       string `json:"podSubnet,omitempty" protobuf:"bytes,2,opt,name=podSubnet"`
	ServiceSubnet   string `json:"serviceSubnet,omitempty" protobuf:"bytes,3,opt,name=serviceSubnet"`
	DNSDomain       string `json:"dnsDomain,omitempty" protobuf:"bytes,4,opt,name=dnsDomain"`
	// NEW
	// Replacing API_SERVERS https://github.com/kubernetes/kubernetes/blob/62898319dff291843e53b7839c6cde14ee5d2aa4/cluster/aws/util.sh#L1004
	DNSServerIP       string `json:"dnsServerIP,omitempty" protobuf:"bytes,5,opt,name=dnsServerIP"`
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty" protobuf:"bytes,6,opt,name=nonMasqueradeCIDR"`
	MasterSubnet      string `json:"masterSubnet,omitempty" protobuf:"bytes,7,opt,name=masterSubnet"` // delete ?
}

func (n *Networking) SetDefaults() {
	if n.ServiceSubnet == "" {
		n.ServiceSubnet = kubeadmapi.DefaultServicesSubnet
	}
	if n.DNSDomain == "" {
		n.DNSDomain = kubeadmapi.DefaultServiceDNSDomain
	}
	if n.PodSubnet == "" {
		// https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/#pod-network
		switch n.NetworkProvider {
		case "calico":
			n.PodSubnet = "192.168.0.0/16"
		case "flannel":
			n.PodSubnet = "10.244.0.0/16"
		}
	}
}

type AWSSpec struct {
	// aws:TAG KubernetesCluster => clusterid
	IAMProfileMaster string `json:"iamProfileMaster,omitempty" protobuf:"bytes,1,opt,name=iamProfileMaster"`
	IAMProfileNode   string `json:"iamProfileNode,omitempty" protobuf:"bytes,2,opt,name=iamProfileNode"`
	MasterSGName     string `json:"masterSGName,omitempty" protobuf:"bytes,3,opt,name=masterSGName"`
	NodeSGName       string `json:"nodeSGName,omitempty" protobuf:"bytes,4,opt,name=nodeSGName"`
	VpcCIDR          string `json:"vpcCIDR,omitempty" protobuf:"bytes,5,opt,name=vpcCIDR"`
	VpcCIDRBase      string `json:"vpcCIDRBase,omitempty" protobuf:"bytes,6,opt,name=vpcCIDRBase"`
	MasterIPSuffix   string `json:"masterIPSuffix,omitempty" protobuf:"bytes,7,opt,name=masterIPSuffix"`
	SubnetCIDR       string `json:"subnetCidr,omitempty" protobuf:"bytes,8,opt,name=subnetCidr"`
}

type GoogleSpec struct {
	NetworkName string   `gcfg:"network-name" ini:"network-name,omitempty" protobuf:"bytes,1,opt,name=networkName"`
	NodeTags    []string `gcfg:"node-tags" ini:"node-tags,omitempty,omitempty" protobuf:"bytes,2,rep,name=nodeTags"`
	// gce
	// NODE_SCOPES="${NODE_SCOPES:-compute-rw,monitoring,logging-write,storage-ro}"
	NodeScopes []string `json:"nodeScopes,omitempty" protobuf:"bytes,3,rep,name=nodeScopes"`
}

type AzureSpec struct {
	InstanceImageVersion string `json:"instanceImageVersion,omitempty" protobuf:"bytes,1,opt,name=instanceImageVersion"`
	RootPassword         string `json:"rootPassword,omitempty" protobuf:"bytes,2,opt,name=rootPassword"`
	SubnetCIDR           string `json:"subnetCidr,omitempty" protobuf:"bytes,3,opt,name=subnetCidr"`
	ResourceGroup        string `json:"resourceGroup,omitempty" protobuf:"bytes,4,opt,name=resourceGroup"`
	SubnetName           string `json:"subnetName,omitempty" protobuf:"bytes,5,opt,name=subnetName"`
	SecurityGroupName    string `json:"securityGroupName,omitempty" protobuf:"bytes,6,opt,name=securityGroupName"`
	VnetName             string `json:"vnetName,omitempty" protobuf:"bytes,7,opt,name=vnetName"`
	RouteTableName       string `json:"routeTableName,omitempty" protobuf:"bytes,8,opt,name=routeTableName"`
	StorageAccountName   string `json:"azureStorageAccountName,omitempty" protobuf:"bytes,9,opt,name=azureStorageAccountName"`
}

type LinodeSpec struct {
	// Linode
	RootPassword string `json:"rootPassword,omitempty" protobuf:"bytes,1,opt,name=rootPassword"`
	KernelId     int64  `json:"kernelId,omitempty" protobuf:"varint,2,opt,name=kernelId"`
}

type CloudSpec struct {
	CloudProvider        string      `json:"cloudProvider,omitempty" protobuf:"bytes,1,opt,name=cloudProvider"`
	Project              string      `json:"project,omitempty" protobuf:"bytes,2,opt,name=project"`
	Region               string      `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`
	Zone                 string      `json:"zone,omitempty" protobuf:"bytes,4,opt,name=zone"` // master needs it for ossec
	InstanceImage        string      `json:"instanceImage,omitempty" protobuf:"bytes,5,opt,name=instanceImage"`
	OS                   string      `json:"os,omitempty" protobuf:"bytes,6,opt,name=os"`
	InstanceImageProject string      `json:"instanceImageProject,omitempty" protobuf:"bytes,7,opt,name=instanceImageProject"`
	CCMCredentialName    string      `json:"ccmCredentialName,omitempty" protobuf:"bytes,8,opt,name=ccmCredentialName"`
	SSHKeyName           string      `json:"sshKeyName,omitempty" protobuf:"bytes,9,opt,name=sshKeyName"`
	AWS                  *AWSSpec    `json:"aws,omitempty" protobuf:"bytes,10,opt,name=aws"`
	GCE                  *GoogleSpec `json:"gce,omitempty" protobuf:"bytes,11,opt,name=gce"`
	Azure                *AzureSpec  `json:"azure,omitempty" protobuf:"bytes,12,opt,name=azure"`
	Linode               *LinodeSpec `json:"linode,omitempty" protobuf:"bytes,13,opt,name=linode"`
}

type API struct {
	// AdvertiseAddress sets the address for the API server to advertise.
	AdvertiseAddress string `json:"advertiseAddress" protobuf:"bytes,1,opt,name=advertiseAddress"`
	// BindPort sets the secure port for the API Server to bind to
	BindPort int32 `json:"bindPort" protobuf:"varint,2,opt,name=bindPort"`
}

type ClusterSpec struct {
	Cloud                      CloudSpec         `json:"cloud" protobuf:"bytes,1,opt,name=cloud"`
	API                        API               `json:"api" protobuf:"bytes,2,opt,name=api"`
	Networking                 Networking        `json:"networking" protobuf:"bytes,3,opt,name=networking"`
	KubernetesVersion          string            `json:"kubernetesVersion,omitempty" protobuf:"bytes,4,opt,name=kubernetesVersion"`
	KubeletVersion             string            `json:"kubeletVersion,omitempty" protobuf:"bytes,5,opt,name=kubeletVersion"`
	KubeadmVersion             string            `json:"kubeadmVersion,omitempty" protobuf:"bytes,6,opt,name=kubeadmVersion"`
	Locked                     bool              `json:"locked,omitempty" protobuf:"varint,7,opt,name=locked"`
	CACertName                 string            `json:"caCertName,omitempty" protobuf:"bytes,8,opt,name=caCertName"`
	FrontProxyCACertName       string            `json:"frontProxyCACertName,omitempty" protobuf:"bytes,9,opt,name=frontProxyCACertName"`
	CredentialName             string            `json:"credentialName,omitempty" protobuf:"bytes,10,opt,name=credentialName"`
	KubeletExtraArgs           map[string]string `json:"kubeletExtraArgs,omitempty" protobuf:"bytes,11,rep,name=kubeletExtraArgs"`
	APIServerExtraArgs         map[string]string `json:"apiServerExtraArgs,omitempty" protobuf:"bytes,12,rep,name=apiServerExtraArgs"`
	ControllerManagerExtraArgs map[string]string `json:"controllerManagerExtraArgs,omitempty" protobuf:"bytes,13,rep,name=controllerManagerExtraArgs"`
	SchedulerExtraArgs         map[string]string `json:"schedulerExtraArgs,omitempty" protobuf:"bytes,14,rep,name=schedulerExtraArgs"`
	AuthorizationModes         []string          `json:"authorizationModes,omitempty" protobuf:"bytes,15,rep,name=authorizationModes"`
	APIServerCertSANs          []string          `json:"apiServerCertSANs,omitempty" protobuf:"bytes,16,rep,name=apiServerCertSANs"`

	// Deprecated
	MasterInternalIP string `json:"-"`
	// the master root ebs volume size (typically does not need to be very large)
	// Deprecated
	MasterDiskId string `json:"-"`

	// Delete since moved to NodeGroup / Instance
	// Deprecated
	MasterDiskType string `json:"-"`
	// If set to Elasticsearch IP, master instance will be associated with this IP.
	// If set to auto, a new Elasticsearch IP will be acquired
	// Otherwise amazon-given public ip will be used (it'll change with reboot).
	// Deprecated
	MasterReservedIP string `json:"-"`
}

type AWSStatus struct {
	MasterSGId string `json:"masterSGID,omitempty" protobuf:"bytes,1,opt,name=masterSGID"`
	NodeSGId   string `json:"nodeSGID,omitempty" protobuf:"bytes,2,opt,name=nodeSGID"`

	VpcId         string `json:"vpcID,omitempty" protobuf:"bytes,3,opt,name=vpcID"`
	SubnetId      string `json:"subnetID,omitempty" protobuf:"bytes,4,opt,name=subnetID"`
	RouteTableId  string `json:"routeTableID,omitempty" protobuf:"bytes,5,opt,name=routeTableID"`
	IGWId         string `json:"igwID,omitempty" protobuf:"bytes,6,opt,name=igwID"`
	DHCPOptionsId string `json:"dhcpOptionsID,omitempty" protobuf:"bytes,7,opt,name=dhcpOptionsID"`
	VolumeId      string `json:"volumeID,omitempty" protobuf:"bytes,8,opt,name=volumeID"`

	// Deprecated
	RootDeviceName string `json:"-"`
}

type CloudStatus struct {
	SShKeyExternalID string     `json:"sshKeyExternalID,omitempty" protobuf:"bytes,1,opt,name=sshKeyExternalID"`
	AWS              *AWSStatus `json:"aws,omitempty" protobuf:"bytes,2,opt,name=aws"`
}

/*
+---------------------------------+
|                                 |
|  +---------+     +---------+    |     +--------+
|  | PENDING +-----> FAILING +----------> FAILED |
|  +----+----+     +---------+    |     +--------+
|       |                         |
|       |                         |
|  +----v----+                    |
|  |  READY  |                    |
|  +----+----+                    |
|       |                         |
|       |                         |
|  +----v-----+                   |
|  | DELETING |                   |
|  +----+-----+                   |
|       |                         |
+---------------------------------+
        |
        |
   +----v----+
   | DELETED |
   +---------+
*/

// ClusterPhase is a label for the condition of a Cluster at the current time.
type ClusterPhase string

// These are the valid statuses of Cluster.
const (
	ClusterPending   ClusterPhase = "Pending"
	ClusterReady     ClusterPhase = "Ready"
	ClusterDeleting  ClusterPhase = "Deleting"
	ClusterDeleted   ClusterPhase = "Deleted"
	ClusterUpgrading ClusterPhase = "Upgrading"
)

type ClusterStatus struct {
	Phase        ClusterPhase       `json:"phase,omitempty,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=ClusterPhase"`
	Reason       string             `json:"reason,omitempty,omitempty" protobuf:"bytes,2,opt,name=reason"`
	Cloud        CloudStatus        `json:"cloud,omitempty" protobuf:"bytes,4,opt,name=cloud"`
	APIAddresses []core.NodeAddress `json:"apiServer,omitempty" protobuf:"bytes,5,rep,name=apiServer"`
	ReservedIPs  []ReservedIP       `json:"reservedIP,omitempty" protobuf:"bytes,6,rep,name=reservedIP"`
}

type ReservedIP struct {
	IP   string `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`
	ID   string `json:"id,omitempty" protobuf:"bytes,2,opt,name=id"`
	Name string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
}

func (c *Cluster) clusterIP(seq int64) string {
	octets := strings.Split(c.Spec.Networking.ServiceSubnet, ".")
	p, _ := strconv.ParseInt(octets[3], 10, 64)
	p = p + seq
	octets[3] = strconv.FormatInt(p, 10)
	return strings.Join(octets, ".")
}

func (c *Cluster) KubernetesClusterIP() string {
	return c.clusterIP(1)
}

func (c Cluster) APIServerURL() string {
	m := map[core.NodeAddressType]string{}
	for _, addr := range c.Status.APIAddresses {
		m[addr.Type] = fmt.Sprintf("https://%s:%d", addr.Address, c.Spec.API.BindPort)
	}
	if u, found := m[core.NodeExternalIP]; found {
		return u
	}
	if u, found := m[core.NodeExternalDNS]; found {
		return u
	}
	return ""
}

// ref: https://github.com/digitalocean/digitalocean-cloud-controller-manager#kubernetes-node-names-must-match-the-droplet-name
func (c *Cluster) APIServerAddress() string {
	m := map[core.NodeAddressType]string{}
	for _, addr := range c.Status.APIAddresses {
		m[addr.Type] = fmt.Sprintf("%s:%d", addr.Address, c.Spec.API.BindPort)
	}

	// ref: https://github.com/kubernetes/kubernetes/blob/d595003e0dc1b94455d1367e96e15ff67fc920fa/cmd/kube-apiserver/app/options/options.go#L99
	addrTypes := []core.NodeAddressType{
		core.NodeInternalDNS,
		core.NodeInternalIP,
		core.NodeExternalDNS,
		core.NodeExternalIP,
	}
	if pat, found := c.Spec.APIServerExtraArgs["kubelet-preferred-address-types"]; found {
		ats := strings.Split(pat, ",")
		addrTypes = make([]core.NodeAddressType, len(ats))
		for i, at := range ats {
			addrTypes[i] = core.NodeAddressType(at)
		}
	}

	for _, at := range addrTypes {
		if u, found := m[at]; found {
			return u
		}
	}
	return ""
}
