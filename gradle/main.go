package main

import (
	"fmt"
)

const DefaultGradleVersion = "latest"

type Gradle struct {
	Version   string
	Container *Container
}

func (g *Gradle) WithVersion(version string) *Gradle {
	g.Version = version
	g.Container = dag.Container().
		From(fmt.Sprintf("gradle:%s", g.Version)).
		WithWorkdir("/app").
		WithMountedCache("/root/.gradle/caches", dag.CacheVolume("gradle-caches")).
		WithMountedCache("/root/.gradle/wrapper", dag.CacheVolume("gradle-wrapper")).
		WithEntrypoint([]string{"gradle"})

	return g
}

func (g *Gradle) WithSource(src *Directory) *Gradle {
	g.checkContainer()

	g.Container = g.Container.WithMountedDirectory("/app", src)
	return g
}

func (g *Gradle) Build() *Container {
	g.checkContainer()

	return g.Container.WithExec([]string{"clean", "build", "--no-daemon"})
}

func (g *Gradle) Test() *Container {
	g.checkContainer()

	return g.Container.WithExec([]string{"clean", "test", "--no-daemon"})
}

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
		g.WithVersion(DefaultGradleVersion)
	} else {
		g.WithVersion(g.Version)
	}
}
