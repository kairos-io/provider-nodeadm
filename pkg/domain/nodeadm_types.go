package domain

// CredentialProvider is the AWS credential provider type.
type CredentialProvider string

// MarshalYAML implements the yaml.Marshaler interface.
func (n *CredentialProvider) MarshalYAML() (interface{}, error) {
	return string(*n), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (n *CredentialProvider) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val string
	if err := unmarshal(&val); err != nil {
		return err
	}
	*n = CredentialProvider(val)
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (n *CredentialProvider) MarshalJSON() ([]byte, error) {
	return []byte(*n), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *CredentialProvider) UnmarshalJSON(val []byte) error {
	*n = CredentialProvider(val)
	return nil
}

const (
	// CredentialProviderIAMRolesAnywhere denotes the IAM Roles Anywhere credential provider.
	CredentialProviderIAMRolesAnywhere CredentialProvider = "iam-ra"
	// CredentialProviderSystemsManager denotes the Systems Manager credential provider.
	CredentialProviderSystemsManager CredentialProvider = "ssm"
)

type NodeadmConfig struct {
	// CredentialProvider is the AWS credential provider type. Supported values are "iam-ra" and "ssm".
	CredentialProvider CredentialProvider `json:"credentialProvider" yaml:"credentialProvider"`
	// KubernetesVersion is the target version of the hybrid edge node.
	// +optional
	KubernetesVersion string `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	// NetworkConfig is the network configuration of the cluster.
	NetworkConfiguration NetworkConfig `json:"networkConfiguration" yaml:"networkConfiguration"`
	// NodeConfig is the configuration of the hybrid edge node.
	NodeConfiguration NodeConfig `json:"nodeConfiguration" yaml:"nodeConfiguration"`
}

type NetworkConfig struct {
	// ServiceCIDR is the cluster's service CIDR.
	ServiceCIDR string `json:"serviceCidr" yaml:"serviceCidr"`
	// PodCIDR is the cluster's pod CIDR.
	PodCIDR string `json:"podCidr" yaml:"podCidr"`
}

type NodeConfig struct {
	// ClusterName is the name of hybrid nodes-enabled EKS cluster.
	ClusterName string `json:"clusterName" yaml:"clusterName"`
	// Region is the AWS region of the hybrid nodes-enabled EKS cluster.
	Region string `json:"region" yaml:"region"`
	// IAMRolesAnywhere contains IAM Roles Anywhere authentication configuration.
	IAMRolesAnywhere *IRAConfig `json:"iamRolesAnywhere,omitempty" yaml:"iamRolesAnywhere,omitempty"`
	// SSM contains Systems Manager authentication configuration.
	SSM *SSMConfig `json:"ssm,omitempty" yaml:"ssm,omitempty"`
}

type IRAConfig struct {
	// NodeName is the name of the hybrid edge node.
	NodeName string `json:"nodeName" yaml:"nodeName"`
	// AssumeRoleARN is the IAM Roles Anywhere assume role ARN.
	AssumeRoleARN string `json:"assumeRoleArn" yaml:"assumeRoleArn"`
	// ProfileARN is the IAM Roles Anywhere profile ARN.
	ProfileARN string `json:"profileArn" yaml:"profileArn"`
	// RoleARN is the IAM Roles Anywhere role ARN.
	RoleARN string `json:"roleArn" yaml:"roleArn"`
	// TrustAnchorARN is the IAM Roles Anywhere trust anchor ARN.
	TrustAnchorARN string `json:"trustAnchorArn" yaml:"trustAnchorArn"`
	// Certificate is the IAM Roles Anywhere certificate.
	Certificate string `json:"certificate" yaml:"certificate"`
	// PrivateKey is the IAM Roles Anywhere private key.
	PrivateKey string `json:"privateKey" yaml:"privateKey"`
}

type SSMConfig struct {
	// ActivationCode is the Systems Manager activation code.
	ActivationCode string `json:"activationCode" yaml:"activationCode"`
	// ActivationID is the Systems Manager activation ID.
	ActivationID string `json:"activationId" yaml:"activationId"`
}
