/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

/*
	what does this command do?

	creates a new release branch in upstream repo based off "main"

	creates a new release branch in forked repo based off newly created release branch in upstream repo

*/

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	//buildToolingRepoName = "eks-anywhere-build-tooling"
	upStreamRepoOwner = "testerIbix" // will eventually be replaced by actual upstream owner, aws
)

// createBranchCmd represents the createBranch command
var createBranchCmd = &cobra.Command{
	Use:   "create-branch",
	Short: "Creates new release branch from updated trigger file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {

		err := createAnywhereBranch()
		if err != nil {
			fmt.Printf("error calling createAnywhereBranch %s", err)
		}
	},
}

func createAnywhereBranch() error {

	latestRelease := os.Getenv("LATEST_RELEASE")

	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	// create branch in upstream repo based off main branch
	ref := "refs/heads/" + latestRelease
	baseRef :=  "main"
	

	// Get the reference for the base branch 
	baseRefObj, _, err := client.Git.GetRef(ctx, upStreamRepoOwner, EKSAnyrepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference: %v", err)
	}

	// Create a new branch
	newBranchRef, _, err := client.Git.CreateRef(ctx, upStreamRepoOwner, EKSAnyrepoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: baseRefObj.Object.SHA,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating branch: %v", err)
	}


	// branch created upstream
	fmt.Printf("New branch '%s' created successfully\n", *newBranchRef.Ref)


	// create branch in forked repo based off upstream
	ref = "refs/heads/" + latestRelease
	baseRef = latestRelease

	// Get the reference for the base branch from the upstream repository
	baseRefObj, _, err = client.Git.GetRef(ctx, upStreamRepoOwner, EKSAnyrepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference: %v", err)
	}

	// Create a new branch
	newBranchRef, _, err = client.Git.CreateRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: baseRefObj.Object.SHA,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating branch: %v", err)
	}

	// branch created upstream
	fmt.Printf("New branch '%s' created successfully\n", *newBranchRef.Ref)


	return nil
}

