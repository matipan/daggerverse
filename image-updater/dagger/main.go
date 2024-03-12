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

// image-updater is a dagger module that updates, commits and pushes a kubernetes deployment file with a new image-url.
//
// There are many alternatives to doing GitOps with kubernetes now a days, to name a few:
//
// - Automatically update the deployment with the new image using Kubectl or hitting the kubernetes API directly (no specific trace of the image deployed in the git repository)
// - Use tools such as Flux or ArgoCD to automatically watch a registry and deploy new images when they appear
// - Use Flux or ArgoCD as well but instead have them look for changes on specific manifests in a repository
//
// This module is useful for the last alternative. When you have CD tools that are watching kubernetes manifests on your
// repository you would need to change them explicitly. If you use Github or Gitlab there are actions that you can use to make
// this changes (for example, there is a yq action and a git-auto-commit action), but the problem is that those workflows
// cannot be tested locally and they become complicated. In the case of github actions, if you run your action as part of a
// workflow that takes a long time to run, it might happen that a new commit showed up and your push will fail. Solving this
// is possible, but it requires adding even more untestable bash. This is why this module exists. With image-updater you can
// implement this logic in a single step that is reproducible locally.
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
