package installer

const (
	DefaultAzureEntraSDKImage  = "mcr.microsoft.com/entra-sdk/auth-sidecar:1.0.0-azurelinux3.0-distroless"
	azureEntraSDKContainerName = "entra-sdk"
	azureEntraSDKPort          = 8080
	azureLinuxBaseCoreImage    = "mcr.microsoft.com/azurelinux/base/core:3.0"
	entraSDKCertsVolumeName    = "entra-sdk-ssl-certs"
)
