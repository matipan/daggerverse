// `kubectl` that provides the kubectl through many authentication methods.
// The goal of this module is to be the one stop shop for interacting with a
// kubernetes cluster. The main challenge when doing this is how authentication
// is done against the cluster. Eventually this module should support all most
// used methods.
// Each top level method is in charge of creating a container with all the tools
// and credentials ready to go for kubectl commands to be executed.
// For an example on how this is implemented you can check out the `KubectlEks`
// method.
package main

import (
	"context"
	"fmt"
)

const (
	KubectlVersion             = "v1.29.1"
	AWSIamAuthenticatorVersion = "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v0.6.14/aws-iam-authenticator_0.6.14_linux_amd64"
	BaseContainerImage         = "debian:trixie-slim"
)

type Kubectl struct {
	Kubeconfig *File
}

// New creates a new instance of the Kubectl module with an already configured
// kubeconfig file. Kubectl is the top level module that provides functions setting
// up the authentication for a specific k8s setup.
func New(kubeconfig *File) *Kubectl {
	return &Kubectl{
		Kubeconfig: kubeconfig,
	}
}

// KubectlCLI is a child module that holds a Container that should already
// be configured to talk to a k8s cluster.
type KubectlCLI struct {
	Container *Container
}

// Exec runs the specified kubectl command.
// NOTE: `kubectl` should be specified as part of the cmd variable.
// For example, to list pods: ["get", "pods", "-n", "namespace"]
func (k *KubectlCLI) Exec(ctx context.Context, cmd []string) (string, error) {
	return k.Container.WithExec(cmd).Stdout(ctx)
}

// DebugSh is a helper function that developers can use to get a terminal
// into the container where the commands are run and troubleshoot potential
// misconfigurations.
// For example:
// dagger call --kubeconfig kubeconfig.yaml kubectl-eks --aws-creds ~/.aws/credentials --aws-profile "example" --aws-config ~/.aws/config debug-sh terminal
func (k *KubectlCLI) DebugSh() *Container {
	return k.Container.WithoutEntrypoint()
}

// KubectlEks returns a KubectlCLI with aws-iam-authenticator and AWS credentials
// configured to communicate with an EKS cluster.
func (m *Kubectl) KubectlEks(ctx context.Context,
	// +optional
	awsConfig *File,
	awsCreds *File,
	awsProfile string,
) *KubectlCLI {
	kubectl := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/amd64/kubectl", KubectlVersion)
	c := dag.Container().
		From(BaseContainerImage).
		// WithExec([]string{"apk", "add", "--no-cache", "--update", "ca-certificates", "curl"}).
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "curl", "gettext-base"}).
		WithExec([]string{"curl", "-sL", "-o", "/bin/kubectl", kubectl}).
		WithExec([]string{"chmod", "+x", "/bin/kubectl"}).
		WithExec([]string{"curl", "-sL", "-o", "/bin/aws-iam-authenticator", AWSIamAuthenticatorVersion}).
		WithExec([]string{"chmod", "+x", "/bin/aws-iam-authenticator"}).
		WithFile("/root/.kube/config", m.Kubeconfig).
		WithFile("/root/.aws/credentials", awsCreds)
	if awsConfig != nil {
		c = c.WithFile("/root/.aws/config", awsConfig)
	}
	return &KubectlCLI{
		Container: c.
			WithEnvVariable("AWS_PROFILE", awsProfile).
			WithEntrypoint([]string{"/bin/kubectl"}),
	}
}
