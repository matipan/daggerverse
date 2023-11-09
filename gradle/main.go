package main

import (
	"fmt"
)

const DefaultGradleVersion = "latest"

type Gradle struct {
	Version   string
	Container *Container
	Wrapper   bool
}

// WithDirectory mounts the directory of the application that will be potentially
// built.
func (g *Gradle) WithDirectory(src *Directory) *Gradle {
	g.checkContainer()

	g.Container = g.Container.WithMountedDirectory("/app", src)
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
	g.Container = dag.Container().
		From(fmt.Sprintf("gradle:%s", g.Version)).
		WithWorkdir("/app").
		WithMountedCache("/root/.gradle/caches", dag.CacheVolume("gradle-caches")).
		WithEntrypoint([]string{"gradle"})

	return g
}

// FromImage sets the image to be used as the base image for the gradle container.
// Keep in mind that if `WithWrapper` is not specified this image must have
// gradle installed.
func (g *Gradle) FromImage(image string) *Gradle {
	g.Container = dag.Container().
		From(image).
		WithWorkdir("/app").
		WithMountedCache("/root/.gradle/caches", dag.CacheVolume("gradle-caches")).
		WithEntrypoint([]string{"gradle"})

	return g
}

// Build runs a clean build.
func (g *Gradle) Build() *Container {
	g.checkContainer()

	return g.Container.WithExec([]string{"clean", "build", "--no-daemon"})
}

// Test runs a clean test.
func (g *Gradle) Test() *Container {
	g.checkContainer()

	return g.Container.WithExec([]string{"clean", "test", "--no-daemon"})
}

// Task allows you to run any custom gradle task you would like.
func (g *Gradle) Task(task string, args ...string) *Container {
	g.checkContainer()

	cmd := append([]string{task}, args...)
	return g.Container.WithExec(cmd)
}

// checkContainer makes sure that gradle's Container is properly
// initialized. If not, it will initialize it with the default
// gradle version
func (g *Gradle) checkContainer() {
	if g.Container != nil {
		return
	}

	if g.Version == "" {
		g.FromVersion(DefaultGradleVersion)
	} else {
		g.FromVersion(g.Version)
	}
}
