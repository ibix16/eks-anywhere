/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

/*
	what does this command do?

	this command is responsible for staging prod cli release

	three functions have been created, two of the three commit updates to the forked repo

	A PR is then created originating from the forked repo targeting the upstream repo latest release branch

	changes are committed into "eks-a-releaser" branch of forked repo
*/

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	prodCliReleaseNumPath = "release/triggers/eks-a-release/production/RELEASE_NUMBER"
	prodCliReleaseVerPath = "release/triggers/eks-a-release/production/RELEASE_VERSION"
)

// prodCliCmd represents the prodCli command
var prodCliCmd = &cobra.Command{
	Use:   "prod-cli",
	Short: "creates a PR containing 2 commits, each updating the contents of a singular file intended for prod cli release",
	Long: `Retrieves updated content for production : release_number and release_version. 
	Writes the updated changes to the two files and raises a PR with the two commits.`,

	Run: func(cmd *cobra.Command, args []string) {
		updateAllProdCliFiles()
	},
}

// runs both updates functions
func updateAllProdCliFiles() {

	errOne := updateProdReleaseNumber()
	if errOne != nil {
		log.Panic(errOne)
	}

	errTwo := updateProdReleaseVersion()
	if errTwo != nil {
		log.Panic(errTwo)
	}
}

// updates release number + creates PR
func updateProdReleaseNumber() error {

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)
	releaseNumber := os.Getenv("RELEASE_NUMBER")
	latestRelease := os.Getenv("LATEST_RELEASE")

	// get latest commit sha from branch "eks-a-releaser"
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliReleaseNumPath, "/")), Type: github.String("blob"), Content: github.String(string(releaseNumber)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, forkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit, update email address
	author := &github.CommitAuthor{
		Name:  github.String("ibix16"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update release number file"),
		Tree:    &github.Tree{SHA: github.String(newTreeSHA)},
		Author:  author,
		Parents: []*github.Commit{{SHA: github.String(latestCommitSha)}},
	}

	commitOP := &github.CreateCommitOptions{}
	newCommit, _, err := client.Git.CreateCommit(ctx, forkedRepoAccount, EKSAnyrepoName, commit, commitOP)
	if err != nil {
		return fmt.Errorf("creating commit %s", err)
	}
	newCommitSHA := newCommit.GetSHA()

	// update branch reference
	ref.Object.SHA = github.String(newCommitSHA)

	_, _, err = client.Git.UpdateRef(ctx, forkedRepoAccount, EKSAnyrepoName, ref, false)
	if err != nil {
		return fmt.Errorf("error updating ref %s", err)
	}

	// create pull request
	base := latestRelease
	head := fmt.Sprintf("%s:%s", forkedRepoAccount, latestRelease)
	title := "Update version files to stage prod cli release"
	body := "This pull request is responsible for updating the contents of 2 seperate files in order to trigger the prod cli release pipeline"

	newPR := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}

	pr, _, err := client.PullRequests.Create(ctx, upStreamRepoOwner, EKSAnyrepoName, newPR)
	if err != nil {
		return fmt.Errorf("error creating PR %s", err)
	}

	log.Printf("Pull request created: %s\n", pr.GetHTMLURL())
	return nil

}

// updates release version + commits
func updateProdReleaseVersion() error {

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	latestVersion := os.Getenv("LATEST_VERSION")
	latestRelease := os.Getenv("LATEST_RELEASE")

	// get latest commit sha from branch "eks-a-releaser"
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliReleaseVerPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, forkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit, update email address
	author := &github.CommitAuthor{
		Name:  github.String("ibix16"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update version number file"),
		Tree:    &github.Tree{SHA: github.String(newTreeSHA)},
		Author:  author,
		Parents: []*github.Commit{{SHA: github.String(latestCommitSha)}},
	}

	commitOP := &github.CreateCommitOptions{}
	newCommit, _, err := client.Git.CreateCommit(ctx, forkedRepoAccount, EKSAnyrepoName, commit, commitOP)
	if err != nil {
		return fmt.Errorf("creating commit %s", err)
	}
	newCommitSHA := newCommit.GetSHA()

	// update branch reference
	ref.Object.SHA = github.String(newCommitSHA)

	_, _, err = client.Git.UpdateRef(ctx, forkedRepoAccount, EKSAnyrepoName, ref, false)
	if err != nil {
		return fmt.Errorf("error updating ref %s", err)
	}

	return nil
}




