// A generated module for Eksctl functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/eksctl/internal/dagger"
	"fmt"
)

type Eksctl struct {
	Cluster   *dagger.File
	Container *dagger.Container
}

func New(
	// +optional
	// +default="latest"
	version string,
	// +default="~/.aws/credentials"
	awsCreds *dagger.Secret,
	awsProfile string,
	// +optional
	awsConfig *dagger.File,
	cluster *dagger.File,
) *Eksctl {
	return &Eksctl{
		Cluster:   cluster,
		Container: eksctl(version, awsCreds, awsProfile, awsConfig, cluster),
	}
}

// WithContainer allows you to modify the container used to run eksctl.
// You should always use the existing `eksctl.Container` and add things on
// top of it. This is the unsafe alternative to something like accepting a
// function as a parameter that modifies the existing container.
// See https://github.com/dagger/dagger/issues/6213 for more details.
func (m *Eksctl) WithContainer(ctr *dagger.Container) *Eksctl {
	m.Container = ctr
	return m
}

// Exec executes the eksctl command.
func (m *Eksctl) Exec(ctx context.Context, command []string) (string, error) {
	return m.Container.WithExec(command).Stdout(ctx)
}

// Create calls `eksctl create` with the cluster config. Additional
// flags can be provided in `exec` form.
func (m *Eksctl) Create(ctx context.Context,
	// +optional
	flags []string) (string, error) {
	return m.Exec(ctx, append([]string{"create", "cluster", "-f", "/cluster.yaml"}, flags...))
}

// DeleteCluster calls `eksctl delete` on the cluster config. Additional
// flags can be provided in `exec` form.
func (m *Eksctl) Delete(ctx context.Context,
	// +optional
	flags []string) (string, error) {
	return m.Exec(ctx, append([]string{"delete", "cluster", "-f", "/cluster.yaml"}, flags...))
}

// Kubeconfig returns the kubeconfig of the cluster. To download it using Dagger's
// CLI you can call `dagger download`.
func (m *Eksctl) Kubeconfig(ctx context.Context) *dagger.File {
	return m.Container.
		WithExec([]string{"utils", "write-kubeconfig", "-f", "/cluster.yaml", "--kubeconfig", "/kubeconfig.yaml"}).
		File("/kubeconfig.yaml")
}

func eksctl(version string, awsCreds *dagger.Secret, awsProfile string, awsConfig *dagger.File, cluster *dagger.File) *dagger.Container {
	var asset string
	if version == "latest" {
		asset = "https://github.com/eksctl-io/eksctl/releases/latest/download/eksctl_Linux_amd64.tar.gz"
	} else {
		asset = fmt.Sprintf("https://github.com/eksctl-io/eksctl/releases/download/%s/eksctl_Linux_amd64.tar.gz", version)
	}

	return dag.Container().
		From("alpine:3.19").
		WithExec([]string{"apk", "add", "--no-cache", "--update", "curl", "tar"}).
		WithWorkdir("/").
		WithExec([]string{"curl", "-sL", "-o", "eksctl.tar.gz", asset}).
		WithExec([]string{"tar", "-xzf", "eksctl.tar.gz", "-C", "/bin"}).
		WithExec([]string{"rm", "eksctl.tar.gz"}).
		WithExec([]string{"curl", "-sL", "-o", "/bin/aws-iam-authenticator", "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v0.6.14/aws-iam-authenticator_0.6.14_linux_amd64"}).
		WithExec([]string{"chmod", "+x", "/bin/aws-iam-authenticator"}).
		WithMountedSecret("/root/.aws/credentials", awsCreds).
		With(func(c *dagger.Container) *dagger.Container {
			if awsConfig != nil {
				return c.WithFile("/root/.aws/config", awsConfig)
			}
			return c
		}).
		WithEnvVariable("AWS_PROFILE", awsProfile).
		WithFile("/cluster.yaml", cluster).
		WithEntrypoint([]string{"/bin/eksctl"})
}
