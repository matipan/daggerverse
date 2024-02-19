# image-updater
`image-updater` is a dagger module that updates, commits and pushes a kubernetes deployment file with a new image-url.

There are many alternatives to doing GitOps with kubernetes now a days, to name a few:
- Automatically update the deployment with the new image using Kubectl or hitting the kubernetes API directly. No specific trace of the image deployed in the git repository
- Use tools such as Flux or ArgoCD to automatically watch a registry and deploy new images when they appear
- Use Flux or ArgoCD as well but instead have them look for changes on specific manifests in a repository

This tool is useful for the last alternative. When you have CD tools and that are watching kubernetes manifests on your repository you would need to change them explicitly. If you use Github or Gitlab there are actions that you can use to make this changes (for example, there is a `yq` action and a `git-auto-commit` action), but the problem is that those workflows cannot be tested locally and they become complicated. In the case of github actions, if you run your action as part of a workflow that takes a long time to run, it might happen that a new commit showed up and your push will fail. Solving this is possible, but it requires adding even more untestable bash. This is why this module exists. With `image-updater` you can implement this logic in a single step that is reproducible locally.
