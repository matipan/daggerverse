// NOTE: this module is not meant to be directly used.
// NeonPreviews implements a set of functions that were shown in a blog post:
// https://blog.matiaspan.dev/its-fun-to-work-on-ci.
// These functions are not meant to be used directly. They are left in this module
// as reference in case users are interested in implementing what is describe
// in the post and what a good starting point.
package main

import (
	"context"
	"dagger/neon-previews/internal/dagger"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

const (
	PreviewComputeUnits   = "0.25"
	PreviewSuspendTimeout = "300"
	PreviewType           = "read_write"
	PreviewParent         = "main"
	PreviewDatabase       = "example"
	PreviewRole           = "example"
)

type NeonPreviews struct{}

func (m *NeonPreviews) ProvisionPreviewDB(ctx context.Context,
	// branch is the name of the preview branch. By default we should be using
	// the name of the git branch
	branch string,
	// projectID is the Neon project ID
	projectID string,
	neonAPIKey *dagger.Secret,
	// awsDir is the directory where the AWS CLI configuration and credentials
	// are stored
	awsDir *dagger.Directory,
	// awsProfile used to authenticate with AWS and save the connection string
	awsProfile string,
) error {
	branch = getBranchSlug(branch)

	neon := newNeonctl(neonAPIKey, projectID)

	exists, err := isAlreadyProvisioned(ctx, neon, branch)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("preview branch %s already exists", branch)
		return nil
	}

	if _, err := neon.Exec(ctx, "branches", "create",
		"--name", branch,
		"--parent", PreviewParent,
		"--type", PreviewType,
		"--suspend-timeout", PreviewSuspendTimeout,
		"--cu", PreviewComputeUnits,
	); err != nil {
		return err
	}

	connectionString, err := neon.ConnectionString(ctx, branch)
	if err != nil {
		return nil
	}

	_, err = m.putParameter(ctx, awsDir, awsProfile, "neon-"+branch, connectionString)
	return err
}

func (m *NeonPreviews) DestroyPreviewDB(ctx context.Context,
	// branch is the name of the preview branch. By default we should be using
	// the name of the git branch
	branch string,
	// projectID is the Neon project ID
	projectID string,
	neonAPIKey *dagger.Secret,
	// awsDir is the directory where the AWS CLI configuration and credentials
	// are stored
	awsDir *dagger.Directory,
	// awsProfile used to authenticate with AWS and save the connection string
	awsProfile string,
) error {
	branch = getBranchSlug(branch)

	neon := newNeonctl(neonAPIKey, projectID)

	exists, err := isAlreadyProvisioned(ctx, neon, branch)
	if err != nil {
		return err
	}

	if !exists {
		log.Printf("preview branch %s does not exist", branch)
		return nil
	}

	if _, err = neon.Exec(ctx, "branches", "delete", branch); err != nil {
		return err
	}

	_, err = m.deleteParameter(ctx, awsDir, awsProfile, "neon-"+branch)
	return err
}

func isAlreadyProvisioned(ctx context.Context, neon *neonctl, branch string) (bool, error) {
	out, err := neon.Exec(ctx, "branches", "list", "--output", "json")
	if err != nil {
		return false, nil
	}

	res := []struct {
		Name string `json:"name"`
	}{}
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		return false, err
	}

	for _, r := range res {
		if r.Name == branch {
			return true, nil
		}
	}

	return false, nil
}

func newNeonctl(apiKey *dagger.Secret, projectID string) *neonctl {
	return &neonctl{
		ctr: dag.Container().
			From("debian:stable-20250113-slim@sha256:b5ace515e78743215a1b101a6f17e59ed74b17132139ca3af3c37e605205e973").
			WithExec([]string{"sh", "-c", "apt update && apt install -y curl"}).
			WithExec([]string{"sh", "-c", "curl -sL https://github.com/neondatabase/neonctl/releases/download/v2.6.0/neonctl-linux-x64 -o /bin/neonctl"}).
			WithExec([]string{"chmod", "+x", "/bin/neonctl"}).
			WithSecretVariable("NEON_API_KEY", apiKey).
			WithEnvVariable("CACHE_BUST", time.Now().String()),
		projectID: projectID,
	}
}

type neonctl struct {
	ctr       *dagger.Container
	projectID string
}

func (n *neonctl) Exec(ctx context.Context, args ...string) (string, error) {
	return n.ctr.WithEnvVariable("CACHE_BUST", time.Now().String()).WithExec(append([]string{"/bin/neonctl", "--project-id", n.projectID}, args...)).Stdout(ctx)
}

func (n *neonctl) ConnectionString(ctx context.Context, branch string) (*dagger.Secret, error) {
	connectionString, err := n.ctr.WithEnvVariable("CACHE_BUST", time.Now().String()).
		WithExec([]string{
			"sh", "-c",
			fmt.Sprintf("/bin/neonctl --project-id %s connection-string %s --role-name %s --database-name %s > /tmp/connection-string", n.projectID, branch, PreviewRole, PreviewDatabase)}).
		File("/tmp/connection-string").
		Contents(ctx)
	if err != nil {
		return nil, err
	}

	return dag.SetSecret("connection-string", strings.Trim(connectionString, "\n")), nil
}

func (m *NeonPreviews) putParameter(ctx context.Context, awsDir *dagger.Directory, awsProfile, name string, connectionString *dagger.Secret) (string, error) {
	return awsCli(awsDir, awsProfile).
		WithMountedSecret("/tmp/connection-string", connectionString).
		WithExec(
			[]string{"sh", "-c", fmt.Sprintf("aws ssm put-parameter --type String --name %s --overwrite --value $(cat /tmp/connection-string)", name)},
		).
		Stdout(ctx)
}

func (m *NeonPreviews) deleteParameter(ctx context.Context, awsDir *dagger.Directory, awsProfile, name string) (string, error) {
	return awsCli(awsDir, awsProfile).
		WithExec([]string{"ssm", "delete-parameter", "--name", name}, dagger.ContainerWithExecOpts{UseEntrypoint: true}).
		Stdout(ctx)
}
func awsCli(awsDir *dagger.Directory, awsProfile string) *dagger.Container {
	return dag.Container().
		From("amazon/aws-cli:latest").
		WithMountedDirectory("/root/.aws", awsDir).
		WithEnvVariable("AWS_PROFILE", awsProfile).
		WithEnvVariable("AWS_REGION", "us-east-2").
		WithEnvVariable("CACHE_BUST", time.Now().String())
}

func getBranchSlug(branch string) string {
	slug.MaxLength = 50
	slug.CustomSub = map[string]string{
		"_": "-",
	}
	return slug.Make(branch)
}
