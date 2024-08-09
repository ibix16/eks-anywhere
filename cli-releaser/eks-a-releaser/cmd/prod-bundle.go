/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

/*
	what does this command do?

	this command is responsible for staging prod bundle release

	four seperate functions have been created, three out of the four update a file and commit the changes to the forked repo

	A PR is then created from the forked repo to the main upstream eks-a repo
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
	prodBundleNumPath     = "release/triggers/bundle-release/production/BUNDLE_NUMBER"
	prodCliMaxVersionPath = "release/triggers/bundle-release/production/CLI_MAX_VERSION"
	prodCliMinVersionPath = "release/triggers/bundle-release/production/CLI_MIN_VERSION"
)

// prodBundleCmd represents the prodBundle command
var prodBundleCmd = &cobra.Command{
	Use:   "prod-bundle",
	Short: "creates a PR containing 3 commits, each updating the contents of a singular file intended for prod bundle release",
	Long: `Retrieves updated content for production : bundle number, cli max version, and cli min version. 
	Writes the updated changes to the 3 files and raises a PR with the 3 commits.`,

	Run: func(cmd *cobra.Command, args []string) {
		err := updateAllProdBundleFiles()
		if err != nil {
			log.Panic(err)
		}
	},
}

func updateAllProdBundleFiles() error {

	errOne := updateProdBundleNumber()
	if errOne != nil {
		return errOne
	}

	errTwo := updateProdMaxVersion()
	if errTwo != nil {
		return errTwo
	}

	errThree := updateProdMinVersion()
	if errThree != nil {
		return errThree
	}

	return nil
}

func updateProdBundleNumber() error {

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// fetch latest release number from env 
	bundleNumber := os.Getenv("RELEASE_NUMBER")
	latestRelease := os.Getenv("LATEST_RELEASE")


	// get latest commit sha from "eks-a-releaser" branch
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodBundleNumPath, "/")), Type: github.String("blob"), Content: github.String(string(bundleNumber)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, forkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return fmt.Errorf("error creating tree %s", err)
	}


	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit, update email address associated with commit
	author := &github.CommitAuthor{
		Name:  github.String("ibix16"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update production bundle number file"),
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
	title := "Update version files to stage production bundle release"
	body := "This pull request is responsible for updating the contents of 3 seperate files in order to trigger the production bundle release pipeline"

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

func updateProdMaxVersion() error {

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	latestVersion := os.Getenv("LATEST_VERSION")
	latestRelease := os.Getenv("LATEST_RELEASE")
	

	// get latest commit sha from "eks-a-releaser" branch
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliMaxVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, forkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit
	author := &github.CommitAuthor{
		Name:  github.String("ibix16"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update production max version file"),
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

func updateProdMinVersion() error {

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	latestVersion := os.Getenv("LATEST_VERSION")
	latestRelease := os.Getenv("LATEST_RELEASE")

	// get latest commit sha from "eks-a-releaser" branch
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliMinVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, forkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit
	author := &github.CommitAuthor{
		Name:  github.String("ibix16"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update production min version file"),
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
