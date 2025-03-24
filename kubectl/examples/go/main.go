package main

import (
	"context"
	"dagger/examples/internal/dagger"
)

type Examples struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Examples) KubectlKubectlEks(ctx context.Context, kubeConfig, awsCreds *dagger.Secret, awsProfile string) (string, error) {
	k := dag.Kubectl(kubeConfig).
		KubectlEks(awsCreds, awsProfile)

	return k.
		Exec(ctx, []string{"get", "pods", "-n", "kube-system"})
}

// Get the logs of a pod from a deployment using the deployment name
func (m *Examples) Kubectl_CreatePod(ctx context.Context, kubeConfig, awsCreds *dagger.Secret, awsProfile string) (string, error) {
	k := dag.Kubectl(kubeConfig).
		KubectlEks(awsCreds, awsProfile)

	// use `k` to get the list of pods from a deployment, getting only the pod
	// name as output
	out, err := k.Exec(ctx, []string{"get", "pods", "-n", "kube-system", "-o", "jsonpath='{.items[0].metadata.name}'"})
	if err != nil {
		return "", err
	}

	// return the logs for the pod that was fetched
	return k.Exec(ctx, []string{"logs", "-n", "kube-system", out})
}
