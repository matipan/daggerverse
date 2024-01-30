package main

import (
	"context"
	"fmt"
)

type Eksctl struct {
	Cluster   *File
	Container *Container
}

func New(version Optional[string], awsCreds *File, awsProfile string, cluster *File) *Eksctl {
	return &Eksctl{
		Cluster:   cluster,
		Container: eksctl(version.GetOr("latest"), awsCreds, awsProfile, cluster),
	}
}

// With allows you to modify the container to do things such as
// mount additional config, credentials, cache volumes, etc.
func (m *Eksctl) With(ctrFunc func(c *Container) *Container) *Eksctl {
	m.Container = ctrFunc(m.Container)
	return m
}

// Exec executes the eksctl command.
func (m *Eksctl) Exec(ctx context.Context, command []string) (string, error) {
	return m.Container.WithExec(command).Stdout(ctx)
}

// CreateCluster calls `eksctl create` with the cluster config. Additional
// flags can be provided in `exec` form.
func (m *Eksctl) CreateCluster(ctx context.Context, flags ...string) (string, error) {
	cmd := append([]string{"create", "cluster", "-f", "cluster.yaml"}, flags...)
	return m.Container.
		WithWorkdir("/cluster").
		WithFile("cluster.yaml", m.Cluster).
		WithExec(cmd).
		Stdout(ctx)
}

// DeleteCluster calls `eksctl delete` on the cluster config. Additional
// flags can be provided in `exec` form.
func (m *Eksctl) DeleteCluster(ctx context.Context, flags ...string) (string, error) {
	cmd := append([]string{"delete", "cluster", "-f", "cluster.yaml"}, flags...)
	return m.Container.
		WithWorkdir("/cluster").
		WithFile("cluster.yaml", m.Cluster).
		WithExec(cmd).
		Stdout(ctx)
}

// Kubeconfig returns the kubeconfig of the cluster. To download it using Dagger's
// CLI you can call `dagger download`.
func (m *Eksctl) Kubeconfig(ctx context.Context, cluster *File) *File {
	return m.Container.
		WithFile("/cluster/cluster.yaml", cluster).
		WithExec([]string{"utils", "write-kubeconfig", "-f", "/cluster/cluster.yaml", "--kubeconfig", "/cluster/kubeconfig.yaml"}).
		File("/cluster/kubeconfig.yaml")
}

func eksctl(version string, awsCreds *File, awsProfile string, cluster *File) *Container {
	return dag.Container().
		From("alpine:3.19").
		WithExec([]string{"apk", "add", "--no-cache", "--update", "curl", "tar"}).
		WithWorkdir("/").
		WithExec([]string{"curl", "-sL", "-o", "eksctl.tar.gz", fmt.Sprintf("https://github.com/eksctl-io/eksctl/releases/%s/download/eksctl_Linux_amd64.tar.gz", version)}).
		WithExec([]string{"tar", "-xzf", "eksctl.tar.gz", "-C", "/bin"}).
		WithExec([]string{"rm", "eksctl.tar.gz"}).
		WithFile("/root/.aws/credentials", awsCreds).
		WithEnvVariable("AWS_PROFILE", awsProfile).
		WithFile("/cluster.yaml", cluster).
		WithEntrypoint([]string{"/bin/eksctl"})
}
