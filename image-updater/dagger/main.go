package main

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const YqVersion = "4.40.7"

type ImageUpdater struct{}

// Update updates the kubernetes deployment file in the specified repository
// with the new image URL.
// NOTE: this pushes a commit to your repository so make sure that you either
// don't have a cyclic workflow trigger or that you use a token that prevents
// this from happening.
// +optional forceWithLease
func (m *ImageUpdater) Update(ctx context.Context, repo, branch, deployFilepath, imageUrl, gitUser, gitEmail string, gitPassword *Secret, forceWithLease bool) error {
	githubPassword, err := gitPassword.Plaintext(ctx)
	if err != nil {
		return err
	}

	repoAuth := &http.BasicAuth{
		Username: gitUser,
		Password: githubPassword,
	}

	repository, err := git.PlainClone("/tmp/repo", false, &git.CloneOptions{
		URL:           repo,
		Auth:          repoAuth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Depth:         1,
		SingleBranch:  true,
	})
	if err != nil {
		return err
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}

	if err := m.updateFile(ctx, worktree, deployFilepath, imageUrl); err != nil {
		return err
	}

	// Commit the changes of the deployment file and push them to the branch
	if _, err := worktree.Add(deployFilepath); err != nil {
		return err
	}

	if _, err := worktree.Commit(fmt.Sprintf("Updating deployment with image: %s", imageUrl), &git.CommitOptions{
		Author: &object.Signature{
			Name:  gitUser,
			Email: gitEmail,
		},
	}); err != nil {
		return err
	}

	refName := plumbing.NewBranchReferenceName(branch)
	pushOptions := &git.PushOptions{
		RemoteName: "origin",
		Auth:       repoAuth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(refName + ":" + refName),
		},
	}
	if forceWithLease {
		pushOptions.ForceWithLease = &git.ForceWithLease{}
	}

	return repository.PushContext(ctx, pushOptions)
}

// updateFile opens the file at the specified filepath, edits the image spec setting
// the new image URL that was specified and writes the file back to the worktree
func (m *ImageUpdater) updateFile(ctx context.Context, worktree *git.Worktree, deployFilepath, imageUrl string) error {
	file, err := worktree.Filesystem.Open(deployFilepath)
	if err != nil {
		return err
	}
	defer file.Close()

	deployment, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	updated, err := dag.Container().
		From("mikefarah/yq:"+YqVersion).
		WithNewFile("deployment.yaml", ContainerWithNewFileOpts{
			Contents:    string(deployment),
			Permissions: 0o666,
		}).
		WithoutEntrypoint().
		WithExec([]string{"sh", "-c", "yq -i '.spec.template.spec.containers[0].image = \"" + imageUrl + "\"' deployment.yaml"}).
		File("deployment.yaml").
		Contents(ctx)
	if err != nil {
		return err
	}

	f, err := worktree.Filesystem.Create(deployFilepath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(updated)); err != nil {
		return err
	}
	return nil
}
