/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

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
	cliReleaseNumPath  = "release/triggers/eks-a-release/development/RELEASE_NUMBER"
	cliReleaseVerPath  = "release/triggers/eks-a-release/development/RELEASE_VERSION"
)

// stageCliCmd represents the stageCli command
var stageCliCmd = &cobra.Command{
	Use:   "stage-cli",
	Short: "creates a PR containing 2 commits, each updating the contents of a singular file intended for staging cli release",
	Long: `Retrieves updated content for development : release_number and release_version. 
	Writes the updated changes to the two files and raises a PR with the two commits.`,

	Run: func(cmd *cobra.Command, args []string) {
		updateAllStageCliFiles()
	},
}



// runs both updates functions
func updateAllStageCliFiles(){

	errOne := updateReleaseNumber()
	if errOne != nil{
		log.Panic(errOne)
	}

	errTwo := updateReleaseVersion()
	if errTwo != nil{
		log.Panic(errTwo)
	}
}


// updates release number + creates PR
func updateReleaseNumber()(error){
	//create client
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN2")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// access trigger file and retrieve contents
	triggerFileContentBundleNumber,_,_, err := client.Repositories.GetContents(ctx, PersonalforkedRepoOwner, repoName, triggerFilePath, nil)
	if err != nil {
		fmt.Print("first breakpoint", err)
	}
	content, err := triggerFileContentBundleNumber.GetContent()
	if err != nil {
		fmt.Print("second breakpoint", err)
	}

	// Find the line containing the identifier
	snippetStartIdentifierB := "RELEASE_NUMBER: "
	lines := strings.Split(content, "\n")
	startIndex := -1
	endIndex := -1

	for i, line := range lines {
		if strings.Contains(line, snippetStartIdentifierB) {
			startIndex = i
			endIndex = i // Set endIndex to the same line as startIndex
			break
		}
	}
	if startIndex == -1 && endIndex == -1 {
		//return fmt.Errorf("snippet not found", nil)  // Snippet not found
		log.Panic("snippet not found...")
	}

	// holds full string 
	bundleNumberLine := lines[startIndex]

	// split string to isolate bundle number
	parts := strings.Split(bundleNumberLine, ": ")

	// holds bundle number value as string 
	desiredPart := parts[1]

	
	// get latest commit sha
	ref, _, err := client.Git.GetRef(ctx, PersonalforkedRepoOwner, repoName, "heads/eks-a-releaser")
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliReleaseNumPath, "/")), Type: github.String("blob"), Content: github.String(string(desiredPart)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx,PersonalforkedRepoOwner, repoName, *ref.Object.SHA, entries)
	if err != nil {
		 return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit
	author := &github.CommitAuthor{
	Name:  github.String("ibix16"),
	Email: github.String("ibixrivera16@gmail.com"),
	}

	commit := &github.Commit{
	Message: github.String("Update release number file"),
	Tree:    &github.Tree{SHA: github.String(newTreeSHA)},
	Author:  author,
	Parents: []*github.Commit{{SHA: github.String(latestCommitSha)}},
	}

	commitOP := &github.CreateCommitOptions{}
	newCommit, _, err := client.Git.CreateCommit(ctx, PersonalforkedRepoOwner, repoName, commit, commitOP)
	if err != nil {
	return fmt.Errorf("creating commit %s", err)
	}
	newCommitSHA := newCommit.GetSHA()
	
	// update branch reference
	ref.Object.SHA = github.String(newCommitSHA)

	_, _, err = client.Git.UpdateRef(ctx, PersonalforkedRepoOwner, repoName, ref, false)
	if err != nil {
	return fmt.Errorf("error updating ref %s", err)
	}

	// create pull request
	base := "main"
	head := fmt.Sprintf("%s:%s", PersonalforkedRepoOwner, "eks-a-releaser")
	title := "Update version files to stage cli release"
	body := "This pull request is responsible for updating the contents of 3 seperate files in order to trigger the staging bundle release pipeline"

	newPR := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}

	pr, _, err := client.PullRequests.Create(ctx, PersonalforkedRepoOwner, repoName, newPR)
	if err != nil {
		return fmt.Errorf("error creating PR %s", err)
	}

	log.Printf("Pull request created: %s\n", pr.GetHTMLURL())
	return nil

}



// updates release version + commits
func updateReleaseVersion()(error){

	//create client
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN2")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// access trigger file and retrieve contents
	triggerFileContentBundleNumber,_,_, err := client.Repositories.GetContents(ctx, PersonalforkedRepoOwner, repoName, triggerFilePath, nil)
	if err != nil {
		fmt.Print("first breakpoint", err)
	}
	content, err := triggerFileContentBundleNumber.GetContent()
	if err != nil {
		fmt.Print("second breakpoint", err)
	}

	// Find the line containing the identifier
	snippetStartIdentifierB := "RELEASE_VERSION: "
	lines := strings.Split(content, "\n")
	startIndex := -1
	endIndex := -1

	for i, line := range lines {
		if strings.Contains(line, snippetStartIdentifierB) {
			startIndex = i
			endIndex = i 
			break
		}
	}
	if startIndex == -1 && endIndex == -1 {
		log.Panic("snippet not found...")
	}

	// holds full string 
	bundleNumberLine := lines[startIndex]

	// split string to isolate bundle number
	parts := strings.Split(bundleNumberLine, ": ")

	// holds bundle number value as string 
	desiredPart := parts[1]

	
	// get latest commit sha
	ref, _, err := client.Git.GetRef(ctx, PersonalforkedRepoOwner, repoName, "heads/eks-a-releaser")
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliReleaseVerPath, "/")), Type: github.String("blob"), Content: github.String(string(desiredPart)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx,PersonalforkedRepoOwner, repoName, *ref.Object.SHA, entries)
	if err != nil {
		 return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit
	author := &github.CommitAuthor{
	Name:  github.String("ibix16"),
	Email: github.String("ibixrivera16@gmail.com"),
	}

	commit := &github.Commit{
	Message: github.String("Update version number file"),
	Tree:    &github.Tree{SHA: github.String(newTreeSHA)},
	Author:  author,
	Parents: []*github.Commit{{SHA: github.String(latestCommitSha)}},
	}

	commitOP := &github.CreateCommitOptions{}
	newCommit, _, err := client.Git.CreateCommit(ctx, PersonalforkedRepoOwner, repoName, commit, commitOP)
	if err != nil {
	return fmt.Errorf("creating commit %s", err)
	}
	newCommitSHA := newCommit.GetSHA()
	
	// update branch reference
	ref.Object.SHA = github.String(newCommitSHA)

	_, _, err = client.Git.UpdateRef(ctx, PersonalforkedRepoOwner, repoName, ref, false)
	if err != nil {
	return fmt.Errorf("error updating ref %s", err)
	}
	
	return nil

}