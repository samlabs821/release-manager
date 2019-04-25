package flow

import (
	"context"
	"fmt"
	"os"
	"path"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/lunarway/release-manager/internal/git"
	"github.com/lunarway/release-manager/internal/log"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
)

// Promote promotes a specific service to environment env.
//
// By convention, promotion means:
//
//    Move released version of the previous environment into this environment
//
// Promotion follows this convention
//
//   master -> dev -> staging -> prod
//
// Flow
//
// Checkout the current kubernetes configuration status and find the
// artifact.json spec for the service and previous environment.
// Use the artifact ID as a key for locating the artifacts.
//
// Find the commit with the artifact ID and checkout the config repository at
// this point.
//
// Copy artifacts from the current release into the new environment and commit
// the changes
func (s *Service) Promote(ctx context.Context, actor Actor, environment, service string) (string, error) {
	sourceConfigRepoPath, closeSource, err := tempDir("k8s-config-promote-source")
	if err != nil {
		return "", err
	}
	defer closeSource()
	destinationConfigRepoPath, closeDestination, err := tempDir("k8s-config-promote-destination")
	if err != nil {
		return "", err
	}
	defer closeDestination()
	// find current released artifact.json for service in env - 1 (dev for staging, staging for prod)
	log.Debugf("Cloning source config repo %s into %s", s.ConfigRepoURL, sourceConfigRepoPath)
	sourceRepo, err := git.Clone(ctx, s.ConfigRepoURL, sourceConfigRepoPath, s.SSHPrivateKeyPath)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("clone '%s' into '%s'", s.ConfigRepoURL, sourceConfigRepoPath))
	}
	sourceSpec, err := sourceSpec(sourceConfigRepoPath, s.ArtifactFileName, service, environment)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("locate source spec"))
	}

	// find release identifier in artifact.json
	release := sourceSpec.ID
	// ckechout commit of release
	var hash plumbing.Hash
	// when promoting to dev we use should look for the artifact instead of
	// release as the artifact have never been released.
	if environment == "dev" {
		hash, err = git.LocateArtifact(sourceRepo, release)
	} else {
		hash, err = git.LocateRelease(sourceRepo, release)
	}
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("locate release '%s' from '%s'", release, s.ConfigRepoURL))
	}
	log.Debugf("internal/flow: Promote: release hash '%v'", hash)
	err = git.Checkout(sourceRepo, hash)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("checkout release hash '%s' from '%s'", hash, s.ConfigRepoURL))
	}

	destinationRepo, err := git.Clone(ctx, s.ConfigRepoURL, destinationConfigRepoPath, s.SSHPrivateKeyPath)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("clone destination repo '%s' into '%s'", s.ConfigRepoURL, destinationConfigRepoPath))
	}

	// release service to env from original release
	sourcePath := srcPath(sourceConfigRepoPath, service, "master", environment)
	destinationPath := releasePath(destinationConfigRepoPath, service, environment)
	log.Debugf("Copy resources from: %s to %s", sourcePath, destinationPath)

	// empty existing resources in destination
	err = os.RemoveAll(destinationPath)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("remove destination path '%s'", destinationPath))
	}
	err = os.MkdirAll(destinationPath, os.ModePerm)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("create destination dir '%s'", destinationPath))
	}
	// copy previous env. files into destination
	err = copy.Copy(sourcePath, destinationPath)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("copy resources from '%s' to '%s'", sourcePath, destinationPath))
	}
	// copy artifact spec
	artifactSourcePath := srcPath(sourceConfigRepoPath, service, "master", s.ArtifactFileName)
	artifactDestinationPath := path.Join(releasePath(destinationConfigRepoPath, service, environment), s.ArtifactFileName)
	log.Debugf("Copy artifact from: %s to %s", artifactSourcePath, artifactDestinationPath)
	err = copy.Copy(artifactSourcePath, artifactDestinationPath)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("copy artifact spec from '%s' to '%s'", artifactSourcePath, artifactDestinationPath))
	}

	authorName := sourceSpec.Application.AuthorName
	authorEmail := sourceSpec.Application.AuthorEmail
	releaseMessage := git.ReleaseCommitMessage(environment, service, release)
	log.Debugf("Committing release: %s, Author: %s <%s>, Committer: %s <%s>", releaseMessage, authorName, authorEmail, actor.Name, actor.Email)
	err = git.Commit(ctx, destinationRepo, releasePath(".", service, environment), authorName, authorEmail, actor.Name, actor.Email, releaseMessage, s.SSHPrivateKeyPath)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("commit changes from path '%s'", destinationPath))
	}
	err = s.notifyRelease(NotifyReleaseOptions{
		Service:       service,
		Environment:   environment,
		ArtifactID:    sourceSpec.ID,
		CommitAuthor:  sourceSpec.Application.AuthorName,
		CommitMessage: sourceSpec.Application.Message,
		CommitSHA:     sourceSpec.Application.SHA,
		CommitLink:    sourceSpec.Application.URL,
		Releaser:      actor.Name,
	})
	if err != nil {
		log.Errorf("flow: Promote: error notifying release: %v", err)
	}

	return release, nil
}