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
package main

import (
	"context"
	"dagger/image-updater/internal/dagger"
	"fmt"
	"io"
	"time"

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
func (m *ImageUpdater) Update(ctx context.Context,
	// name of the application that is being updated. appName is used on the commit message
	// if no name is provided then a generic message is committed.
	// +optional
	appName string,
	// repository to clone
	repo string,
	// branch to checkout
	branch string,
	// list of files that should be updated
	files []string,
	// full URL of the image to set
	imageUrl string,
	// username for the author of the commit
	gitUser string,
	// email used for both the commit and the authentication
	gitEmail string,
	// password to authenticate against git server.
	gitPassword *dagger.Secret,
	// if specified then the push is made with --force-with-lease
	// +optional
	forceWithLease bool,
	// list of container IDs to update on each of the files
	// +optional
	containers []int,
) error {
	if len(containers) == 0 {
		containers = []int{0}
	}

	githubPassword, err := gitPassword.Plaintext(ctx)
	if err != nil {
		return err
	}

	yqContainer, err := dag.Container().
		From("mikefarah/yq:" + YqVersion).
		Sync(ctx)
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

	if err := m.updateFiles(ctx, yqContainer, worktree, imageUrl, files, containers); err != nil {
		return err
	}

	// Commit the changes of the deployment file and push them to the branch
	for _, file := range files {
		if _, err := worktree.Add(file); err != nil {
			return err
		}
	}

	msg := fmt.Sprintf("Updating resource with image: %s", imageUrl)
	if appName != "" {
		msg = fmt.Sprintf("Updating %s resource with image: %s", appName, imageUrl)
	}
	if _, err := worktree.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  gitUser,
			Email: gitEmail,
			When:  time.Now(),
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

// updateFiles opens each file at the specified filepath, edits the image spec setting
// for each of the containers with the new image URL that was specified and writes
// the file back to the worktree
func (m *ImageUpdater) updateFiles(ctx context.Context, yq *dagger.Container, worktree *git.Worktree, imageUrl string, files []string, containers []int) error {
	for _, filePath := range files {
		file, err := worktree.Filesystem.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		deployment, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		updated, err := yq.
			WithNewFile("deployment.yaml", string(deployment), dagger.ContainerWithNewFileOpts{
				Permissions: 0o666,
			}).
			WithoutEntrypoint().
			With(func(c *dagger.Container) *dagger.Container {
				for _, cid := range containers {
					c = c.WithExec([]string{"sh", "-c", fmt.Sprintf("yq -i '.spec.template.spec.containers[%d].image = \"%s\"' deployment.yaml", cid, imageUrl)})
				}
				return c
			}).
			File("deployment.yaml").
			Contents(ctx)
		if err != nil {
			return err
		}

		f, err := worktree.Filesystem.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Write([]byte(updated)); err != nil {
			return err
		}
	}

	return nil
}
