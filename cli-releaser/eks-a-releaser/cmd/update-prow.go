/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

/*
	what does this command do?

	currently :

	updateTemplaterFile() - accessess the templater file from the provided path on "main" branch, forked repo
	retrieves content from templater file

	updates file content to point to new release, stores updated file content in a variable
	creates new file path/name by altering previous file & updating "release-0.00.yaml" portion

	deletes previously exisiting file using previous file path/name ~ templaterFilePath
	creates a new file using the updated file path/name and the updated file content

	commits changes to prow-jobs repo latest release branch, forked repo

	raises PR with commits targeting upstream repo "main" branch
*/
import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	prowRepoName      = "eks-anywhere-prow-jobs"
)

// upProwCmd represents the upProw command
var updateProwCmd = &cobra.Command{
	Use:   "update-prow",
	Short: "accesses prow-jobs repo and updates version files",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {
		updateTemplaterFile()
	},
}


func updateTemplaterFile() {

	// var holds latest release retrieved from trigger file
	latestRelease := os.Getenv("LATEST_RELEASE")

	// var holds latest file name
	latestFileName, err := FetchFileName(forkedRepoAccount, prowRepoName, "templater/jobs/periodic/eks-anywhere-build-tooling", "eks-a-releaser")
	if err != nil {
		fmt.Print("error fetching file names", err)
	}

	// var holds updated full file path
	templaterFilePath := "/templater/jobs/periodic/eks-anywhere-build-tooling/" + latestFileName

	// create client 
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	opts := &github.RepositoryContentGetOptions{
		Ref: "main", // branch to access
	}

	// access templater file on main branch and retrieve entire file contents
	templaterFileContent, _, _, err := client.Repositories.GetContents(ctx, forkedRepoAccount, prowRepoName, templaterFilePath, opts)
	if err != nil {
		fmt.Print("first breakpoint", err)
	}

	// var "content" holds entire string of templater file
	content, err := templaterFileContent.GetContent()
	if err != nil {
		fmt.Print("second breakpoint", err)
	}

	// update jobName field, isolate line
	nameSnippetStartIdentifier := "jobName: "
	Firstlines := strings.Split(content, "\n")
	startIndex := -1
	endIndex := -1

	for i, line := range Firstlines {
		if strings.Contains(line, nameSnippetStartIdentifier) {
			startIndex = i
			endIndex = i // Set endIndex to the same line as startIndex
			break
		}
	}
	if startIndex == -1 && endIndex == -1 {
		// return fmt.Errorf("snippet not found", nil)  // Snippet not found
		log.Panic("snippet not found...")
	}

	// holds string - name: eks-anywhere-attribution-periodic-release-0-19
	nameLine := Firstlines[startIndex]

	jobNameLineParts := strings.Split(nameLine, "release-")

	// holds string 0-19
	jobNameLineReleasePortion := jobNameLineParts[1]

	// latestRelease var holds release-0.00 from trigger file
	// we want to isolate the numerical portion
	// and convert it from 0.00 ---> 0-00
	splitLatestRelease := strings.Split(latestRelease, ".")
	targetLatestReleaseValue := splitLatestRelease[1]

	// var holds 0-21
	convertedTargetLatestReleaseValue := "0-" + targetLatestReleaseValue

	firstUpdatedFile := strings.ReplaceAll(content, jobNameLineReleasePortion, convertedTargetLatestReleaseValue)

	// update jobName field, isolate line
	nameSnippetStartIdentifier = "value: "
	Firstlines = strings.Split(content, "\n")
	startIndex = -1
	endIndex = -1

	for i, line := range Firstlines {
		if strings.Contains(line, nameSnippetStartIdentifier) {
			startIndex = i
			endIndex = i // Set endIndex to the same line as startIndex
			break
		}
	}
	if startIndex == -1 && endIndex == -1 {
		// return fmt.Errorf("snippet not found", nil)  // Snippet not found
		log.Panic("snippet not found...")
	}

	// holds value: release-0.00 from templater file
	nameLine = Firstlines[startIndex]

	// isolates release-0.00 portion from templater file
	valueLine := strings.Split(nameLine, ": ")
	valueLinePortion := valueLine[1]

	// replaces all instances of "release-0.00" with var valueLinePortion, updating both value: line and baseRef: line
	secondUpdatedFileContent := strings.ReplaceAll(firstUpdatedFile, valueLinePortion, latestRelease)


	// all required fields successfully get updated


	// variable holds temp file path, removing the leading "/"
	prevFileName := "templater/jobs/periodic/eks-anywhere-build-tooling/" + latestFileName

	parts := strings.Split(prevFileName, "periodics-")

	// index 1 : release-0.00.yaml
	// index 0 : /templater/jobs/periodic/eks-anywhere-build-tooling/eks-anywhere-attribution-
	fmt.Print(parts[0])

	newFilePathString := parts[0] + "periodics-" + latestRelease + ".yaml"

	// by the end of this function we have : the updated content for the file ~ in a string variable : secondUpdatedFile
	// the updated file path including the file name for the new file that needs to be created ~ in a string variable : newString


	err = deleteFile(ctx, client, forkedRepoAccount, prowRepoName, prevFileName, "eks-a-releaser")
	if err != nil {
		fmt.Printf("error:  %s", err)
	}

	err = createFile(forkedRepoAccount, prowRepoName, newFilePathString, secondUpdatedFileContent)
	if err != nil {
		fmt.Printf("error:  %s", err)
	}


	// insert logic here to execute make command 


	err = createPullRequest(ctx, client, "main", "Update Templater File", "This PR updates the templater file for the new release.")
	if err != nil {
		fmt.Printf("error:  %s", err)
	}
}

func createFile(repoOwner, repoName, filePath, content string) error {

	accessToken := os.Getenv("SECRET_PAT")

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", repoOwner, repoName, filePath)

	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
	data := map[string]interface{}{
		"message": "Create file",
		"content": encodedContent,
		"branch":  "eks-a-releaser", // Ensure the changes are made on the eks-a-releaser branch
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create file: %s", body)
	}

	return nil
}

func deleteFile(ctx context.Context, client *github.Client, repoOwner, repoName, filePath, branch string) error {
	opts := &github.RepositoryContentGetOptions{Ref: branch}

	// Get the file information to retrieve the SHA
	fileContent, _, _, err := client.Repositories.GetContents(ctx, repoOwner, repoName, filePath, opts)
	if err != nil {
		return fmt.Errorf("failed to get file information: %v", err)
	}

	sha := fileContent.GetSHA()
	message := "Delete outdated file"
	options := &github.RepositoryContentFileOptions{
		Message: &message,
		SHA:     &sha,
		Branch:  &branch,
	}

	_, _, err = client.Repositories.DeleteFile(ctx, repoOwner, repoName, filePath, options)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

func createPullRequest(ctx context.Context, client *github.Client, baseBranch, title, body string) error {

	head := fmt.Sprintf("%s:%s", forkedRepoAccount, "eks-a-releaser")

	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  &head,
		Base:  github.String(baseBranch),
		Body:  github.String(body),
	}


	pr, _, err := client.PullRequests.Create(ctx, upStreamRepoOwner, prowRepoName, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("Created PR: %s\n", pr.GetHTMLURL())
	return nil
}




func FetchFileName(owner, repo, dir, branch string)(string, error){

	// create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	opts := &github.RepositoryContentGetOptions{
		Ref: branch,
	}

	_, dirContents, _, err := client.Repositories.GetContents(ctx, owner, repo, dir, opts)
	if err != nil {
		return "error fetching files from repo", err
	}



	// extract file names
	var fileNames []string
	for _, file := range dirContents {
		fileNames = append(fileNames, *file.Name)
	}


	// filters to return "release" file name only
	for _, name := range fileNames{
		if strings.Contains(name, "release-"){
			return name, nil
		}
	}

	return "file not found", nil
}

// successfully deletes old file and creates new file on forked repo, eks-a-releaser branch 
// successfully creates a PR targetting upstream prow jobs repo with the new commits, merges into "main" branch

// partially incomplete, the templater file correctly gets updated with the new release contents 
// A PR is created right away after the templater file is updated 
// this is incorrect as a command needs to be executed after the templater file has been updated but prior to the PR creation