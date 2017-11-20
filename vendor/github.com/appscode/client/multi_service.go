package client

import (
	attic "github.com/appscode/api/attic/v1beta1"
	auth "github.com/appscode/api/auth/v1beta1"
	ca "github.com/appscode/api/certificate/v1beta1"
	ci "github.com/appscode/api/ci/v1beta1"
	cluster_v1alpha1 "github.com/appscode/api/cluster/v1alpha1"
	k8s_v1beta1 "github.com/appscode/api/kubernetes/v1beta1"
	namespace "github.com/appscode/api/namespace/v1beta1"
	"google.golang.org/grpc"
)

// multi client services are grouped by there main client. the api service
// clients are wrapped around with sub-service.
type multiClientInterface interface {
	Attic() *atticService
	Authentication() *authenticationService
	CA() *caService
	CI() *ciService
	Namespace() *nsService
	Cluster() *versionedClusterService
	Kubernetes() *versionedKubernetesService
}

type multiClientServices struct {
	atticClient               *atticService
	authenticationClient      *authenticationService
	caClient                  *caService
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
		caClient: &caService{
			certificateClient: ca.NewCertificatesClient(conn),
		},
		ciClient: &ciService{
			agentsClient:   ci.NewAgentsClient(conn),
			metadataClient: ci.NewMetadataClient(conn),
		},
		versionedClusterClient: &versionedClusterService{
			v1alpha1Service: &clusterV1alpha1Service{
				credentialClient: cluster_v1alpha1.NewCredentialsClient(conn),
				clusterClient:    cluster_v1alpha1.NewClustersClient(conn),
				nodeGroupClient:  cluster_v1alpha1.NewNodeGroupsClient(conn),
				sshConfigClient:  cluster_v1alpha1.NewSSHConfigClient(conn),
				metdataClient:    cluster_v1alpha1.NewMetadataClient(conn),
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

func (s *multiClientServices) CA() *caService {
	return s.caClient
}

func (s *multiClientServices) CI() *ciService {
	return s.ciClient
}

func (s *multiClientServices) Cluster() *versionedClusterService {
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

type caService struct {
	certificateClient ca.CertificatesClient
}

func (c *caService) CertificatesClient() ca.CertificatesClient {
	return c.certificateClient
}

type versionedClusterService struct {
	v1alpha1Service *clusterV1alpha1Service
}

func (v *versionedClusterService) V1alpha1() *clusterV1alpha1Service {
	return v.v1alpha1Service
}

type clusterV1alpha1Service struct {
	credentialClient cluster_v1alpha1.CredentialsClient
	clusterClient    cluster_v1alpha1.ClustersClient
	nodeGroupClient  cluster_v1alpha1.NodeGroupsClient
	sshConfigClient  cluster_v1alpha1.SSHConfigClient
	metdataClient    cluster_v1alpha1.MetadataClient
}

func (k *clusterV1alpha1Service) Credential() cluster_v1alpha1.CredentialsClient {
	return k.credentialClient
}

func (k *clusterV1alpha1Service) Cluster() cluster_v1alpha1.ClustersClient {
	return k.clusterClient
}

func (k *clusterV1alpha1Service) NodeGroup() cluster_v1alpha1.NodeGroupsClient {
	return k.nodeGroupClient
}

func (k *clusterV1alpha1Service) SSHConfig() cluster_v1alpha1.SSHConfigClient {
	return k.sshConfigClient
}

func (k *clusterV1alpha1Service) Metadata() cluster_v1alpha1.MetadataClient {
	return k.metdataClient
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
