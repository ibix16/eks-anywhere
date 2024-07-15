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
	bundleNumPath  = "release/triggers/bundle-release/development/BUNDLE_NUMBER"
	cliMaxVersionPath  = "release/triggers/bundle-release/development/CLI_MAX_VERSION"
	cliMinVersionPath  = "release/triggers/bundle-release/development/CLI_MIN_VERSION"
	triggerFilePath = "release/triggers/eks-a-releaser-trigger"
	PersonalforkedRepoOwner = "ibix16"
)

// stageBundleCmd represents the stageBundle command
var stageBundleCmd = &cobra.Command{
	Use:   "stage-bundle",
	Short: "creates a PR containing 3 commits, each updating the contents of a singular file intended for staging bundle release",
	Long: `Retrieves updated content for development : bundle number, cli max version, and cli min version. 
	Writes the updated changes to the 3 files and raises a PR with the 3 commits.`,

	Run: func(cmd *cobra.Command, args []string) {
		err := updateAllThree()
		if err != nil{
			log.Print(err)
		}
	},
}



// this function is responsible for invoking the 3 other functions 
// will create a PR with 3 commits, from eks-a-releaser branch targetting main branch (of my forked copy)
func updateAllThree() (error){
	errOne := updateDevBundleNumber()
	if errOne != nil{
		log.Panic("error calling function A")
	}

	errTwo := updateDevCliMaxVersion()
	if errTwo != nil{
		log.Panic("error calling function B")
	}

	errThree := updateDevCliMinVersion()
	if errThree != nil{
		log.Panic("error calling function C")
	}
	
	return nil
}



// this function is responsible for updating the bundle number file in order to trigger the staging bundle release pipeline 
// the function accesses the trigger file and retrieves the value assigned to dev/bundle-number: #
// a new commit and PR is then created using the retrieved value from the trigger file
func updateDevBundleNumber() (error) {
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
	snippetStartIdentifierB := "dev/bundle-number: "
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
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(bundleNumPath, "/")), Type: github.String("blob"), Content: github.String(string(desiredPart)), Mode: github.String("100644")})
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
	Message: github.String("Increment bundle number file"),
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
    title := "Update version files to stage bundle release"
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



// this function is responsible for updating the cli max version file in order to trigger the staging bundle release pipeline 
// the function accesses the trigger file and retrieves the first line of code containg the version e.g "v0.0.0"
// a new commit and PR is then created using the retrieved value from the trigger file
func updateDevCliMaxVersion() (error) {
	//create client
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN2")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	fileContent, _, _, err := client.Repositories.GetContents(ctx, PersonalforkedRepoOwner, repoName, triggerFilePath, nil)
	if err != nil {
		fmt.Print(err)
	}

	// variable content holds content for entire file
	content, err := fileContent.GetContent()
	if err != nil {
		fmt.Print(err)
	}

	// Find the line containing the identifier
	snippetStartIdentifierB := "dev/CLI_MAX_VERSION: "
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
		log.Panic("snippet not found....")
	}

	// holds string
	cliMaxVer := lines[startIndex]

	// split string to isolate bundle number
	parts := strings.Split(cliMaxVer, ": ")

	// holds bundle number value as string
	desiredPart := parts[1]

	// get latest commit sha
	ref, _, err := client.Git.GetRef(ctx, PersonalforkedRepoOwner, repoName, "heads/eks-a-releaser")
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliMaxVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(desiredPart)), Mode: github.String("100644")})
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
	Message: github.String("Update CLI Max Version number"),
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




// this function is responsible for updating the cli min version file in order to trigger the staging bundle release pipeline 
// the function accesses the trigger file and retrieves the first line of code containg the version e.g "v0.0.0"
// a new commit and PR is then created using the retrieved value from the trigger file
func updateDevCliMinVersion() (error) {

	//create client
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN2")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	fileContent, _, _, err := client.Repositories.GetContents(ctx, PersonalforkedRepoOwner, repoName, triggerFilePath, nil)
	if err != nil {
		fmt.Print(err)
	}

	// variable content holds content for entire file
	content, err := fileContent.GetContent()
	if err != nil {
		fmt.Print(err)
	}

	// Find the line containing the identifier
	snippetStartIdentifierB := "dev/CLI_MIN_VERSION: "
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
		log.Panic("snippet not found!")  // Snippet not found
	}

	// holds string for base_ref: release-0.19
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
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliMinVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(desiredPart)), Mode: github.String("100644")})
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
	Message: github.String("Update CLI Min Version number"),
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
