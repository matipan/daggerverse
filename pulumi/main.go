package main

import (
	"context"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Pulumi struct {
	AwsAccessKey *Secret
	AwsSecretKey *Secret
	PulumiToken  *Secret
	Version      string
}

// FromVersion is an optional function that users can use to specify
// the version of pulumi's docker image to use as base.
func (m *Pulumi) FromVersion(version string) *Pulumi {
	m.Version = version
	return m
}

// WithAwsCredentials sets the AWS credentials to be used by Pulumi.
// Call this function if you want pulumi to point your changes to AWS.
func (m *Pulumi) WithAwsCredentials(awsAccessKey, awsSecretKey *Secret) *Pulumi {
	m.AwsAccessKey = awsAccessKey
	m.AwsSecretKey = awsSecretKey
	return m
}

// WithPulumiToken sets the Pulumi token to be used by Pulumi.
func (m *Pulumi) WithPulumiToken(pulumiToken *Secret) *Pulumi {
	m.PulumiToken = pulumiToken
	return m
}

// Up runs the `pulumi up` command for the given stack and directory.
// NOTE: This command will perform changes in your cloud.
func (m *Pulumi) Up(ctx context.Context, src *Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi up --stack %s --yes --non-interactive", stack))
}

// Preview runs the `pulumi preview` command for the given stack and directory
// returning the output of the diff that was generated.
func (m *Pulumi) Preview(ctx context.Context, src *Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi preview --stack %s --non-interactive --diff", stack))
}

// Refresh runs the `pulumi refresh` command for the given stack and directory
// returning the output of the diff if there was any.
func (m *Pulumi) Refresh(ctx context.Context, src *Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi refresh --stack %s --non-interactive --diff", stack))
}

// Destroy runs the `pulumi destroy` command for the given stack and directory.
// NOTE: This command will destroy all the resources created by the stack.
func (m *Pulumi) Destroy(ctx context.Context, src *Directory, stack string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi destroy --stack %s --non-interactive --yes", stack))
}

// Run runs the specified pulumi command. For example: preview --diff.
func (m *Pulumi) Run(ctx context.Context, src *Directory, command string) (string, error) {
	return m.commandOutput(ctx, src, fmt.Sprintf("pulumi %s", command))
}

// commandOutput runs the given command in the pulumi container and returns its output.
func (m *Pulumi) commandOutput(ctx context.Context, src *Directory, command string) (string, error) {
	ct, err := m.authenticatedContainer(ctx, src)
	if err != nil {
		return "", err
	}

	return ct.
		WithExec([]string{"-c", command}).
		Stdout(ctx)
}

// authenticatedContainer returns a pulumi container with the required credentials.
// Users have to set credentials for their cloud provider by using the `With<Provider>Credentials`
// function.
func (m *Pulumi) authenticatedContainer(ctx context.Context, src *Directory) (*Container, error) {
	if m.PulumiToken == nil {
		return nil, errors.New("pulumi token is required. Use `with-pulumi-token` to set it")
	}

	ct, err := container(ctx, src, m.PulumiToken, m.Version)
	if err != nil {
		return nil, err
	}

	switch {
	case m.AwsAccessKey != nil && m.AwsSecretKey != nil:
		ct = ct.WithSecretVariable("AWS_ACCESS_KEY_ID", m.AwsAccessKey).
			WithSecretVariable("AWS_SECRET_ACCESS_KEY", m.AwsSecretKey)
	default:
		return nil, errors.New("no cloud provider credentails was provided")
	}

	return ct, nil
}

// container obtains a base container with pulumi's CLI installed.
func container(ctx context.Context, src *Directory, pulumiToken *Secret, version string) (*Container, error) {
	f := src.File("Pulumi.yaml")
	if f == nil {
		return nil, errors.New("Pulumi.yaml file not found")
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

	return dag.
		Container().
		From(fmt.Sprintf("pulumi/pulumi-%s:%s", project.Runtime, version)).
		WithSecretVariable("PULUMI_ACCESS_TOKEN", pulumiToken).
		WithMountedDirectory("/infra", src).
		WithWorkdir("/infra").
		WithEntrypoint([]string{"/bin/bash"}).
		WithExec([]string{"-c", depCmd}), nil
}
