// Package domain contains provider-nodeadm's configuration types.
package domain

import (
	"github.com/aws/eks-hybrid/api/v1alpha1"
)

// Provider option keys
const (
	// ClusterRootPathKey is the key for the cluster's root path.
	ClusterRootPathKey = "cluster_root_path"

	// CredentialProviderKey is the key for the AWS credential provider type.
	CredentialProviderKey string = "credentialProvider"

	// KubernetesVersionKey is the key for the target version of the hybrid edge node.
	KubernetesVersionKey string = "kubernetesVersion"

	// NetworkConfigurationKey is the key for the network configuration of the cluster.
	NetworkConfigurationKey string = "networkConfiguration"

	// NodeConfigurationKey is the key for the configuration of the hybrid edge node.
	NodeConfigurationKey string = "nodeConfiguration"

	// HandleDependenciesKey is the key for configuring dependency installation/upgrade.
	// If set to "true", nodeadm install and upgrade will be invoked. Otherwise, only nodeadm init will be invoked.
	HandleDependenciesKey string = "handleDependencies"
)

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

// NodeadmConfig contains all configuration required by nodeadm to bootstrap an EKS hybrid node.
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

// NetworkConfig contains the EKS hybrid node's network configuration.
type NetworkConfig struct {
	// ServiceCIDR is the cluster's service CIDR.
	ServiceCIDR string `json:"serviceCidr" yaml:"serviceCidr"`

	// PodCIDR is the cluster's pod CIDR.
	PodCIDR string `json:"podCidr" yaml:"podCidr"`
}

// NodeConfig contains the EKS hybrid node's local configuration.
type NodeConfig struct {
	// ClusterName is the name of hybrid nodes-enabled EKS cluster.
	ClusterName string `json:"clusterName" yaml:"clusterName"`

	// Region is the AWS region of the hybrid nodes-enabled EKS cluster.
	Region string `json:"region" yaml:"region"`

	// IAMRolesAnywhere contains IAM Roles Anywhere authentication configuration.
	IAMRolesAnywhere *IRAConfig `json:"iamRolesAnywhere,omitempty" yaml:"iamRolesAnywhere,omitempty"`

	// SSM contains Systems Manager authentication configuration.
	SSM *SSMConfig `json:"ssm,omitempty" yaml:"ssm,omitempty"`

	// UserConfig is the user-provided configuration for the hybrid edge node.
	UserConfig *UserNodeConfig `json:"userNodeConfig,omitempty" yaml:"userNodeConfig,omitempty"`
}

// UserNodeConfig contains user-provided configuration for the hybrid edge node.
type UserNodeConfig struct {
	// ContainerdOptions are additional parameters passed to `containerd`.
	Containerd v1alpha1.ContainerdOptions `json:"containerd,omitempty"`

	// KubeletOptions are additional parameters passed to `kubelet`.
	Kubelet v1alpha1.KubeletOptions `json:"kubelet,omitempty"`
}

// IRAConfig contains IAM Roles Anywhere authentication configuration.
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

// SSMConfig contains Systems Manager authentication configuration.
type SSMConfig struct {
	// ActivationCode is the Systems Manager activation code.
	ActivationCode string `json:"activationCode" yaml:"activationCode"`

	// ActivationID is the Systems Manager activation ID.
	ActivationID string `json:"activationId" yaml:"activationId"`
}
