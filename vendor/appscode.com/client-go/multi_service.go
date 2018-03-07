package client

import (
	auth "appscode.com/api/auth/v1alpha1"
	cloud_v1alpha1 "appscode.com/api/cloud/v1alpha1"
	k8s_v1beta1 "appscode.com/api/kubernetes/v1alpha1"
	namespace "appscode.com/api/namespace/v1alpha1"
	"google.golang.org/grpc"
)

// multi client services are grouped by there main client. the api service
// clients are wrapped around with sub-service.
type multiClientInterface interface {
	Authentication() *authenticationService
	Namespace() *nsService
	Cloud() *versionedClusterService
	Kubernetes() *versionedKubernetesService
}

type multiClientServices struct {
	authenticationClient      *authenticationService
	nsClient                  *nsService
	versionedClusterClient    *versionedClusterService
	versionedKubernetesClient *versionedKubernetesService
}

func newMultiClientService(conn *grpc.ClientConn) multiClientInterface {
	return &multiClientServices{
		authenticationClient: &authenticationService{
			authenticationClient: auth.NewAuthenticationClient(conn),
			conduitClient:        auth.NewConduitClient(conn),
			projectClient:        auth.NewProjectsClient(conn),
		},
		versionedClusterClient: &versionedClusterService{
			v1alpha1Service: &clusterV1alpha1Service{
				certificateClient: cloud_v1alpha1.NewCertificatesClient(conn),
				credentialClient:  cloud_v1alpha1.NewCredentialsClient(conn),
				clusterClient:     cloud_v1alpha1.NewClustersClient(conn),
				nodeGroupClient:   cloud_v1alpha1.NewNodeGroupsClient(conn),
				sshConfigClient:   cloud_v1alpha1.NewSSHConfigClient(conn),
				metadataClient:    cloud_v1alpha1.NewMetadataClient(conn),
			},
		},
		versionedKubernetesClient: &versionedKubernetesService{
			v1beta1Service: &kubernetesV1beta1Service{
				clientsClient: k8s_v1beta1.NewClientsClient(conn),
				diskClient:    k8s_v1beta1.NewDisksClient(conn),
			},
		},
		nsClient: &nsService{
			teamClient: namespace.NewTeamsClient(conn),
		},
	}
}

func (s *multiClientServices) Authentication() *authenticationService {
	return s.authenticationClient
}

func (s *multiClientServices) Namespace() *nsService {
	return s.nsClient
}

func (s *multiClientServices) Cloud() *versionedClusterService {
	return s.versionedClusterClient
}

func (s *multiClientServices) Kubernetes() *versionedKubernetesService {
	return s.versionedKubernetesClient
}

type authenticationService struct {
	authenticationClient auth.AuthenticationClient
	conduitClient        auth.ConduitClient
	projectClient        auth.ProjectsClient
}

func (a *authenticationService) Authentication() auth.AuthenticationClient {
	return a.authenticationClient
}

func (a *authenticationService) Conduit() auth.ConduitClient {
	return a.conduitClient
}

func (a *authenticationService) Project() auth.ProjectsClient {
	return a.projectClient
}

type nsService struct {
	teamClient namespace.TeamsClient
}

func (b *nsService) Team() namespace.TeamsClient {
	return b.teamClient
}

type versionedClusterService struct {
	v1alpha1Service *clusterV1alpha1Service
}

func (v *versionedClusterService) V1alpha1() *clusterV1alpha1Service {
	return v.v1alpha1Service
}

type clusterV1alpha1Service struct {
	certificateClient cloud_v1alpha1.CertificatesClient
	credentialClient  cloud_v1alpha1.CredentialsClient
	clusterClient     cloud_v1alpha1.ClustersClient
	nodeGroupClient   cloud_v1alpha1.NodeGroupsClient
	sshConfigClient   cloud_v1alpha1.SSHConfigClient
	metadataClient    cloud_v1alpha1.MetadataClient
}

func (k *clusterV1alpha1Service) CertificatesClient() cloud_v1alpha1.CertificatesClient {
	return k.certificateClient
}

func (k *clusterV1alpha1Service) Credential() cloud_v1alpha1.CredentialsClient {
	return k.credentialClient
}

func (k *clusterV1alpha1Service) Cluster() cloud_v1alpha1.ClustersClient {
	return k.clusterClient
}

func (k *clusterV1alpha1Service) NodeGroup() cloud_v1alpha1.NodeGroupsClient {
	return k.nodeGroupClient
}

func (k *clusterV1alpha1Service) SSHConfig() cloud_v1alpha1.SSHConfigClient {
	return k.sshConfigClient
}

func (k *clusterV1alpha1Service) Metadata() cloud_v1alpha1.MetadataClient {
	return k.metadataClient
}

type versionedKubernetesService struct {
	v1beta1Service *kubernetesV1beta1Service
}

func (v *versionedKubernetesService) V1beta1() *kubernetesV1beta1Service {
	return v.v1beta1Service
}

type kubernetesV1beta1Service struct {
	clientsClient k8s_v1beta1.ClientsClient
	diskClient    k8s_v1beta1.DisksClient
}

func (k *kubernetesV1beta1Service) Client() k8s_v1beta1.ClientsClient {
	return k.clientsClient
}

func (k *kubernetesV1beta1Service) Disk() k8s_v1beta1.DisksClient {
	return k.diskClient
}
