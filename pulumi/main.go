package main

import (
	"context"
	"errors"
	"fmt"
	"main/internal/dagger"

	"gopkg.in/yaml.v3"
)

type Pulumi struct {
	// +private
	AwsAccessKey *dagger.Secret
	// +private
	AwsSecretKey *dagger.Secret
	// +private
	PulumiToken  *dagger.Secret
	// The Pulumi ESC environment used for AWS OIDC
	EscEnv       string
	// The version of the Pulumi base image
	Version      string
	// Whether a Docker Engine will be bound to the Pulumi container
	Docker       bool
}

// Optional function to specify the version of Pulumi's docker image to use as base
func (m *Pulumi) FromVersion(version string) *Pulumi {
	m.Version = version
	return m
}

// Sets the AWS credentials to be used by Pulumi
// Call this function if you want pulumi to point your changes to AWS
func (m *Pulumi) WithAwsCredentials(awsAccessKey, awsSecretKey *dagger.Secret) *Pulumi {
	m.AwsAccessKey = awsAccessKey
	m.AwsSecretKey = awsSecretKey
	return m
}

// Use Pulumi ESC as the provider of AWS OIDC credentials
func (m *Pulumi) WithEsc(env string) *Pulumi {
	m.EscEnv = env
	return m
}

// Sets the Pulumi token to be used by Pulumi
func (m *Pulumi) WithPulumiToken(pulumiToken *dagger.Secret) *Pulumi {
	m.PulumiToken = pulumiToken
	return m
}

// Sets up the Pulumi container with a Docker Engine Service container
func (m *Pulumi) WithDocker() *Pulumi {
	m.Docker = true
	return m
}

// Runs the `pulumi up` command for the given stack and directory
// NOTE: This command will perform changes in your cloud
func (m *Pulumi) Up(ctx context.Context, src *dagger.Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi up --stack %s --yes --non-interactive", stack))
}

// Runs the `pulumi preview` command for the given stack and directory
// returning the output of the diff that was generated.
func (m *Pulumi) Preview(ctx context.Context, src *dagger.Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi preview --stack %s --non-interactive --diff", stack))
}

// Runs the `pulumi refresh` command for the given stack and directory
// returning the output of the diff if there was any
func (m *Pulumi) Refresh(ctx context.Context, src *dagger.Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi refresh --stack %s --non-interactive --diff", stack))
}

// Destroy runs the `pulumi destroy` command for the given stack and directory.
// NOTE: This command will destroy all the resources created by the stack.
func (m *Pulumi) Destroy(ctx context.Context, src *dagger.Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi destroy --stack %s --non-interactive --yes", stack))
}

// Runs the specified pulumi command. For example: preview --diff.
func (m *Pulumi) Run(ctx context.Context, src *dagger.Directory, command string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi %s", command))
}

// Gets the output value from the stack.
func (m *Pulumi) Output(ctx context.Context, src *dagger.Directory,  property string, stack string) (string, error) {
	selectCmd := fmt.Sprintf("pulumi stack select %s", stack)
	outputCmd := fmt.Sprintf("pulumi stack output %s", property)
	return m.commandOutput(ctx, src, fmt.Sprintf("%s && %s", selectCmd, outputCmd))
}

// commandOutput runs the given command in the pulumi container and returns its output.
func (m *Pulumi) commandOutput(ctx context.Context, src *dagger.Directory, command string) (string, error) {
	ct, err := m.authenticatedContainer(ctx, src)
	if err != nil {
		return "", err
	}

	return ct.
		WithExec([]string{"/bin/bash", "-c", command}).
		Stdout(ctx)
}

// Pulumi container with the required credentials
// Users have to set credentials for their cloud provider by using the `With<Provider>Credentials`
// function or `WithEsc` function for Pulumi AWS OIDC
func (m *Pulumi) authenticatedContainer(ctx context.Context, src *dagger.Directory) (*dagger.Container, error) {
	if m.PulumiToken == nil {
		return nil, errors.New("pulumi token is required. Use `with-pulumi-token` to set it")
	}

	ct, err := m.container(ctx, src, m.PulumiToken, m.Version)
	if err != nil {
		return nil, err
	}

	if m.EscEnv == "" {
		switch {
		case m.AwsAccessKey != nil && m.AwsSecretKey != nil:
			ct = ct.WithSecretVariable("AWS_ACCESS_KEY_ID", m.AwsAccessKey).
				WithSecretVariable("AWS_SECRET_ACCESS_KEY", m.AwsSecretKey)
		default:
			return nil, errors.New("no cloud provider credentails was provided")
		}
	}
	return ct, nil
}

// Base container with Pulumi's CLI installed.
func (m *Pulumi) container(ctx context.Context, src *dagger.Directory, pulumiToken *dagger.Secret, version string) (*dagger.Container, error) {
	f := src.File("Pulumi.yaml")
	if f == nil {
		return nil, errors.New("a Pulumi.yaml file not found")
	}

	b, err := f.Contents(ctx)
	if err != nil {
		return nil, err
	}

	project := struct {
		Runtime string `yaml:"runtime"`
	}{}
	if err := yaml.Unmarshal([]byte(b), &project); err != nil {
		return nil, err
	}

	if version == "" {
		version = "latest"
	}

	depCmd := ""
	switch project.Runtime {
	case "go":
		depCmd = "go mod tidy"
	case "nodejs":
		depCmd = "npm install"
	case "python":
		depCmd = "pip install -r requirements.txt"
	case "dotnet":
	default:
		return nil, fmt.Errorf("unsupported pulumi runtime: %s", project.Runtime)
	}

	escInstallCmd := "curl -fsSL https://get.pulumi.com/esc/install.sh | sh"
	escOpenCmd := fmt.Sprintf("$HOME/.pulumi/bin/esc env open %s", m.EscEnv)
	ct := dag.
		Container().
		From(fmt.Sprintf("pulumi/pulumi-%s:%s", project.Runtime, version)).
		WithSecretVariable("PULUMI_ACCESS_TOKEN", pulumiToken).
		WithMountedDirectory("/infra", src).
		WithWorkdir("/infra").
		WithExec([]string{"/bin/bash", "-c", depCmd}).
		WithExec([]string{"/bin/bash", "-c", escInstallCmd}).
		WithExec([]string{"/bin/bash", "-c", escOpenCmd})
	if m.Docker {
		ct = ct.
			WithEnvVariable("DOCKER_HOST", "tcp://docker:2375").
			WithServiceBinding("docker", dag.Docker().Engine())
	}
	return ct, nil
}
