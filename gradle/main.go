package main

import (
	"fmt"
	"main/internal/dagger"
)

const DefaultGradleVersion = "latest"

type Gradle struct {
	Version   string
	Image     string
	Directory *dagger.Directory
	Wrapper   bool
}

// WithDirectory mounts the directory of the application that will be potentially
// built.
func (g *Gradle) WithDirectory(src *dagger.Directory) *Gradle {
	g.Directory = src
	return g
}

// WithWrapper enables the use of `gradlew` instead of using the gradle installed
// in the image. If `WithWrapper` is called, it is not necessary to set a specific
// version or image.
func (g *Gradle) WithWrapper() *Gradle {
	g.Wrapper = true
	return g
}

// FromVersion sets the gradle version to be used. If not set, the default
// version will be used specified by the `DefaultGradleVersion` constant.
func (g *Gradle) FromVersion(version string) *Gradle {
	g.Version = version
	return g
}

// FromImage sets the image to be used as the base image for the gradle container.
// Keep in mind that if `WithWrapper` is not specified this image must have
// gradle installed.
func (g *Gradle) FromImage(image string) *Gradle {
	g.Image = image
	return g
}

// Container returns a container with gradle, caching and directories mounted
// and ready to be used. You can use this if for any reason the available functions
// are not enough.
func (g *Gradle) Container() *dagger.Container {
	return g.buildContainer()
}

// Build runs a clean build.
func (g *Gradle) Build() *dagger.Container {
	return g.buildContainer().WithExec([]string{"clean", "build", "--no-daemon"})
}

// Test runs a clean test.
func (g *Gradle) Test() *dagger.Container {
	return g.buildContainer().WithExec([]string{"clean", "test", "--no-daemon"})
}

// Task allows you to run any custom gradle task you would like.
func (g *Gradle) Task(task string, args ...string) *dagger.Container {
	return g.buildContainer().WithExec(append([]string{task}, args...))
}

// buildContainer builds a gradle container with the specified version or
// image and the directory that was mounted.
func (g *Gradle) buildContainer() *dagger.Container {
	image := g.Image
	if image == "" {
		version := g.Version
		if version == "" {
			version = DefaultGradleVersion
		}
		image = fmt.Sprintf("gradle:%s", version)
	}

	container := dag.Container().
		From(image).
		WithWorkdir("/app").
		WithMountedCache("/root/.gradle/caches", dag.CacheVolume("gradle-caches"))
	if g.Directory != nil {
		container = container.WithMountedDirectory("/app", g.Directory)
	}

	if g.Wrapper {
		container = container.WithMountedCache("/root/.gradle/wrapper", dag.CacheVolume("gradle-wrapper"))
		return container.WithEntrypoint([]string{"./gradlew"})
	}

	return container.WithEntrypoint([]string{"gradle"})
}
