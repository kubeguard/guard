package client

import (
	attic "appscode.com/api/attic/v1beta1"
	auth "appscode.com/api/auth/v1beta1"
	ci "appscode.com/api/ci/v1beta1"
	cloud_v1alpha1 "appscode.com/api/cloud/v1alpha1"
	k8s_v1beta1 "appscode.com/api/kubernetes/v1beta1"
	namespace "appscode.com/api/namespace/v1beta1"
	"google.golang.org/grpc"
)

// multi client services are grouped by there main client. the api service
// clients are wrapped around with sub-service.
type multiClientInterface interface {
	Attic() *atticService
	Authentication() *authenticationService
	CI() *ciService
	Namespace() *nsService
	Cloud() *versionedClusterService
	Kubernetes() *versionedKubernetesService
}

type multiClientServices struct {
	atticClient               *atticService
	authenticationClient      *authenticationService
	ciClient                  *ciService
	nsClient                  *nsService
	versionedClusterClient    *versionedClusterService
	versionedKubernetesClient *versionedKubernetesService
}

func newMultiClientService(conn *grpc.ClientConn) multiClientInterface {
	return &multiClientServices{
		atticClient: &atticService{
			artifactClient: attic.NewArtifactsClient(conn),
			versionClient:  attic.NewVersionsClient(conn),
		},
		authenticationClient: &authenticationService{
			authenticationClient: auth.NewAuthenticationClient(conn),
			conduitClient:        auth.NewConduitClient(conn),
			projectClient:        auth.NewProjectsClient(conn),
		},
		ciClient: &ciService{
			agentsClient:   ci.NewAgentsClient(conn),
			metadataClient: ci.NewMetadataClient(conn),
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
				incidentClient: k8s_v1beta1.NewIncidentsClient(conn),
				clientsClient:  k8s_v1beta1.NewClientsClient(conn),
				diskClient:     k8s_v1beta1.NewDisksClient(conn),
			},
		},
		nsClient: &nsService{
			teamClient: namespace.NewTeamsClient(conn),
		},
	}
}

func (s *multiClientServices) Attic() *atticService {
	return s.atticClient
}

func (s *multiClientServices) Authentication() *authenticationService {
	return s.authenticationClient
}

func (s *multiClientServices) Namespace() *nsService {
	return s.nsClient
}

func (s *multiClientServices) CI() *ciService {
	return s.ciClient
}

func (s *multiClientServices) Cloud() *versionedClusterService {
	return s.versionedClusterClient
}

func (s *multiClientServices) Kubernetes() *versionedKubernetesService {
	return s.versionedKubernetesClient
}

// original service clients that needs to exposed under grouped wrapper
// services.
type atticService struct {
	artifactClient attic.ArtifactsClient
	versionClient  attic.VersionsClient
}

func (a *atticService) Artifacts() attic.ArtifactsClient {
	return a.artifactClient
}

func (a *atticService) Versions() attic.VersionsClient {
	return a.versionClient
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

type ciService struct {
	agentsClient   ci.AgentsClient
	metadataClient ci.MetadataClient
}

func (a *ciService) Agents() ci.AgentsClient {
	return a.agentsClient
}

func (a *ciService) Metadata() ci.MetadataClient {
	return a.metadataClient
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
	incidentClient k8s_v1beta1.IncidentsClient
	clientsClient  k8s_v1beta1.ClientsClient
	diskClient     k8s_v1beta1.DisksClient
}

func (a *kubernetesV1beta1Service) Incident() k8s_v1beta1.IncidentsClient {
	return a.incidentClient
}

func (k *kubernetesV1beta1Service) Client() k8s_v1beta1.ClientsClient {
	return k.clientsClient
}

func (k *kubernetesV1beta1Service) Disk() k8s_v1beta1.DisksClient {
	return k.diskClient
}
