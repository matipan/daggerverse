// A generated module for Checker functions
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
	"dagger/checker/internal/dagger"
)

type Checker struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Checker) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Checker) CheckDefaultArgs(ctx context.Context,
	// +defaultPath="check-file.txt"
	checkFile *dagger.File,
) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithFile("/check-file", checkFile).
		WithExec([]string{"cat", "/check-file"}).
		Stdout(ctx)
}

func (m *Checker) CheckRequired(ctx context.Context,
	checkFile *dagger.File,
) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithFile("/check-file", checkFile).
		WithExec([]string{"cat", "/check-file"}).
		Stdout(ctx)
}

func (m *Checker) CheckNoArgs(ctx context.Context) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"echo", "hello"}).
		Stdout(ctx)
}
