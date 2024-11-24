package main

import (
	"context"
	"dagger/java/internal/dagger"
)

type Java struct{}

// Build using gradle-wrapper
func (m *Java) GradleBuild(
	ctx context.Context,

	// +optional
	// +default="21"
	javaVersion string,

	// provide gradle-tasks (eg: --gradle-tasks='downloadRepos,installDist')
	// -these will be executed after ./gradlew []
	gradleTasks []string,

	// source code directory path containing gradle-wrapper(eg: --src=.)
	src *dagger.Directory,
) *dagger.Directory {

	// build with gradle-wrapper
	builddir := dag.Container().
		From("eclipse-temurin:"+javaVersion).
		WithWorkdir("/app").
		WithDirectory("/app", src).
		WithExec([]string{"chmod", "+x", "gradlew"})

	// Execute each gradle task
	for _, task := range gradleTasks {
		builddir = builddir.WithExec([]string{"./gradlew", task})
	}

	return builddir.Directory("/app")
}

// Publish the gradle-build
func (m *Java) GradlePublish(
	ctx context.Context,

	// +optional
	// +default="21"
	javaVersion string,

	// provide gradle-tasks (eg: --gradle-tasks='downloadRepos,installDist')
	// -these will be executed after ./gradlew []
	gradleTasks []string,

	// +optional
	// +default="docker.io"
	// oci-registry the image to be published to
	ociRegistry string,

	// username to authenticate with the oci registry
	ociUsername string,

	// password/token to authenticate with the oci registry
	ociPassword *dagger.Secret,

	//  repository in the oci-registry, the image to be published to (usually oci-username for docker.io)
	ociRegistryRepository string,

	// image name (eg: --image-name=myapp)
	imageName string,

	// image tag (eg: --image-tag=v1a1)
	imageTag string,

	// source code directory path containing gradle-wrapper(eg: --src=.)
	src *dagger.Directory,
) (string, error) {
	// get gradle-build directory
	builtDir := m.GradleBuild(ctx, javaVersion, gradleTasks, src)

	// copy "/app" from gradle-build to jre-image
	jreImage := dag.Container().
		From("eclipse-temurin:"+javaVersion+"-jre").
		WithDirectory("/app", builtDir)

	//  publish jre image
	publishJre, err := jreImage.WithRegistryAuth(ociRegistry, ociUsername, ociPassword).Publish(ctx, ociRegistry+"/"+ociRegistryRepository+"/"+imageName+":"+imageTag)
	if err != nil {
		return "", err
	}
	return publishJre, nil
}
