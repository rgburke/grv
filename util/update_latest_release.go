package main

import (
	"context"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	githubAPITokenEnvironmentVariable = "GITHUB_API_TOKEN"
	repoUser                          = "rgburke"
	repoName                          = "grv"
	repoTagName                       = "latest"
	repoReleaseFileName               = "grv_latest_linux64"
	binaryFilePath                    = "grv"
)

var ctx = context.Background()

func main() {
	log.Printf("Starting release update")

	client := createClient()
	deleteExistingRelease(client)
	deleteExistingTag(client)
	newRelease := createNewRelease(client)
	attachBinaryToRelease(client, newRelease)

	log.Printf("Successfully completed release update")
}

func createClient() *github.Client {
	log.Printf("Creating github client")

	githubAPIToken := os.Getenv(githubAPITokenEnvironmentVariable)
	if githubAPIToken == "" {
		log.Fatalf("Environment variable %v is not set", githubAPITokenEnvironmentVariable)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAPIToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func deleteExistingRelease(client *github.Client) {
	log.Printf("Fetching existing release")

	oldRelease, resp, err := client.Repositories.GetReleaseByTag(ctx, repoUser, repoName, repoTagName)
	if resp != nil && resp.StatusCode == 404 {
		log.Printf("Release doesn't exist")
	} else if err != nil {
		log.Fatalf("Unable to fetch release: %v", err)
	} else if oldRelease != nil {
		log.Printf("Deleting release %v", oldRelease.GetName())

		_, err = client.Repositories.DeleteRelease(ctx, repoUser, repoName, oldRelease.GetID())
		if err != nil {
			log.Fatalf("Failed to delete release: %v", err)
		}
	}
}

func deleteExistingTag(client *github.Client) {
	log.Printf("Fetching existing tag")

	ref, resp, err := client.Git.GetRef(ctx, repoUser, repoName, "tags/"+repoTagName)
	if resp != nil && resp.StatusCode == 404 {
		log.Printf("Tag doesn't exist")
	} else if err != nil {
		log.Fatalf("Unable to fetch tag: %v", err)
	} else if ref != nil {
		log.Printf("Deleting tag %v", ref.GetRef())

		_, err = client.Git.DeleteRef(ctx, repoUser, repoName, ref.GetRef())
		if err != nil {
			log.Fatalf("Failed to delete tag: %v", err)
		}
	}
}

func createNewRelease(client *github.Client) *github.RepositoryRelease {
	log.Printf("Creating new release")

	repoReleaseTag := repoTagName
	repoReleaseBranch := "master"
	repoReleaseName := "Latest Build"
	repoReleaseBody := "Auto-generated build of the latest code on master.\n**Note:** This is a pre-release binary and no guarantees are made that it's in a usable or stable state"
	repoReleaseDraft := false
	repoReleasePreRelease := true

	newRelease := &github.RepositoryRelease{
		TagName:         &repoReleaseTag,
		TargetCommitish: &repoReleaseBranch,
		Name:            &repoReleaseName,
		Body:            &repoReleaseBody,
		Draft:           &repoReleaseDraft,
		Prerelease:      &repoReleasePreRelease,
	}

	newRelease, _, err := client.Repositories.CreateRelease(ctx, repoUser, repoName, newRelease)
	if err != nil {
		log.Fatalf("Failed to create new release: %v", err)
	}

	log.Printf("Created release")

	return newRelease
}

func attachBinaryToRelease(client *github.Client, newRelease *github.RepositoryRelease) {
	log.Printf("Attaching binary to release")

	uploadOptions := &github.UploadOptions{
		Name: repoReleaseFileName,
	}

	file, err := os.Open(binaryFilePath)
	if err != nil {
		log.Fatalf("Unable to open binary: %v", err)
	}

	_, _, err = client.Repositories.UploadReleaseAsset(ctx, repoUser, repoName, newRelease.GetID(), uploadOptions, file)
	if err != nil {
		log.Fatalf("Failed to upload binary: %v", err)
	}

	log.Printf("Binary attached to release")
}
