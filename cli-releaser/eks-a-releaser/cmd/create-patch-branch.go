/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

/*
	what does this command do?

	creates a new patch release branch in forked repo based off latest release branch in upstream repo
*/

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

// createPatchBranchCmd represents the createPatchBranch command
var createPatchBranchCmd = &cobra.Command{
	Use:   "create-patch-branch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := createPatchBranch()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}



func createPatchBranch()error{

	latestRelease := os.Getenv("LATEST_RELEASE")

	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	// create branch in forked repo based off upstream
	ref := "refs/heads/" + latestRelease + "-releaser-patch"
	baseRef :=  latestRelease
	

	// Get the reference for the base branch from upstream
	baseRefObj, _, err := client.Git.GetRef(ctx, upStreamRepoOwner, EKSAnyrepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference: %v", err)
	}

	// Create a new branch in fork
	newBranchRef, _, err := client.Git.CreateRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, &github.Reference{
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