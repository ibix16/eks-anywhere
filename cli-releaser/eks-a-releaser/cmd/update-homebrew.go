/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

/*
	what does this command do?
	this command is responsible for updating the homebrew release version file

	retrieves the current file from forked repo and updates it with the new value, committing the changes

	PR is then raised from latest release branch, forked repo, targgetting new release branch on upstream repo
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
	homebrewPath = "release/triggers/brew-version-release/CLI_RELEASE_VERSION"
)

// updateHomebrewCmd represents the updateHomebrew command
var updateHomebrewCmd = &cobra.Command{
	Use:   "update-homebrew",
	Short: "Updates homebrew with latest version in eks-a-releaser branch, PR targets release branch",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {
		runAllHomebrew()
	},
}

func runAllHomebrew(){

	RELEASE_TYPE := os.Getenv("RELEASE_TYPE")	

	errOne := updateHomebrew(RELEASE_TYPE)
	if errOne != nil {
		log.Panic(errOne)
	}

	errTwo := createPullRequestHomebrew(RELEASE_TYPE)
	if errTwo != nil {
		log.Panic(errTwo)
	}
}



func updateHomebrew(releaseType string)error{

	if releaseType == "minor"{

		// fetch latest "v0.xx.xx" from env
	latestVersionValue := os.Getenv("LATEST_VERSION")

	latestRelease := os.Getenv("LATEST_RELEASE")

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	opts := &github.RepositoryContentGetOptions{
		Ref: "main", // specific branch to check for homebrew file
	}

	// access homebrew file
	triggerFileContentBundleNumber, _, _, err := client.Repositories.GetContents(ctx, forkedRepoAccount, EKSAnyrepoName, homebrewPath, opts)
	if err != nil {
		fmt.Print("first breakpoint", err)
	}

	// holds content of homebrew cli version file
	content, err := triggerFileContentBundleNumber.GetContent()
	if err != nil {
		fmt.Print("second breakpoint", err)
	}

	// update instances of previous release with new
	updatedFile := strings.ReplaceAll(content, content, latestVersionValue)


	// get latest commit sha from branch
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(homebrewPath, "/")), Type: github.String("blob"), Content: github.String(string(updatedFile)), Mode: github.String("100644")})
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
		Message: github.String("Update brew-version value to point to new release"),
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

	// else

	// fetch latest "v0.xx.xx" from env
	latestVersionValue := os.Getenv("LATEST_VERSION")

	latestRelease := os.Getenv("LATEST_RELEASE")

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	opts := &github.RepositoryContentGetOptions{
		Ref: "main", // specific branch to check for homebrew file
	}

	// access homebrew file
	triggerFileContentBundleNumber, _, _, err := client.Repositories.GetContents(ctx, forkedRepoAccount, EKSAnyrepoName, homebrewPath, opts)
	if err != nil {
		fmt.Print("first breakpoint", err)
	}

	// holds content of homebrew cli version file
	content, err := triggerFileContentBundleNumber.GetContent()
	if err != nil {
		fmt.Print("second breakpoint", err)
	}

	// update instances of previous release with new
	updatedFile := strings.ReplaceAll(content, content, latestVersionValue)


	// get latest commit sha from branch
	ref, _, err := client.Git.GetRef(ctx, forkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease+"-releaser-patch")
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(homebrewPath, "/")), Type: github.String("blob"), Content: github.String(string(updatedFile)), Mode: github.String("100644")})
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
		Message: github.String("Update brew-version value to point to new release"),
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




func createPullRequestHomebrew(releaseType string)error{

	if releaseType == "minor"{
		
	// fetch latest release "release-0.xx" from env
	latestRelease := os.Getenv("LATEST_RELEASE")

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	
	base := latestRelease // branch PR will be merged into
	head := fmt.Sprintf("%s:%s", forkedRepoAccount, latestRelease)
	title := "Update homebrew cli version value to point to new release"
	body := "This pull request is responsible for updating the contents of the home brew cli version file"

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

	// else 
	// fetch latest release "release-0.xx" from env
	latestRelease := os.Getenv("LATEST_RELEASE")

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	
	base := latestRelease // branch PR will be merged into
	head := fmt.Sprintf("%s:%s", forkedRepoAccount, latestRelease+"-releaser-patch")
	title := "Update homebrew cli version value to point to new release"
	body := "This pull request is responsible for updating the contents of the home brew cli version file"

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