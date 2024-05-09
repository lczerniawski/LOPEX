package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "boom",
		Usage:  "make an explosive entrance",
		Action: mainAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Value:    "",
				Usage:    "Url to look for the files",
				Required: true,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) error {
	var urlFlag = c.String("url")

	println("Downloading .git files")
	err := tryDownloadDotGitFiles(urlFlag)
	if err != nil {
		println(err.Error())
	}

	return nil
}

func tryDownloadDotGitFiles(baseUrl string) error {
	canBeDownloaded, err := checkIfGitFolderCanBeDownloaded(baseUrl)
	if !canBeDownloaded {
		return err
	}

	println("Git folder found, downloading files")
	println("Initializing git repo")

	err = initializeGitRepo()
	if err != nil {
		return err
	}

	commitHashes, err := getCommitHashes(baseUrl)
	if err != nil {
		return err
	}

	err = downloadAndSaveGitHubObjectFiles(baseUrl, commitHashes)
	if err != nil {
		return err
	}

	treeHashes, err := getTreeHashesFromCommits(commitHashes)
	if err != nil {
		return err
	}

	err = downloadAndSaveGitHubObjectFiles(baseUrl, treeHashes)
	if err != nil {
		return err
	}

	blobHashesWithName, err := getBlobHashesAndNamesFromTrees(treeHashes)
	if err != nil {
		return err
	}

	blobHashes := convertMapToArray(blobHashesWithName)
	err = downloadAndSaveGitHubObjectFiles(baseUrl, blobHashes)
	if err != nil {
		return err
	}

	// Call for getting getTreeHashesFromTrees to get all the files inside folders
	// Needs to be called in recursive way until there are no more folders

	for hash, name := range blobHashesWithName {
		content, err := getFileContentFromBlob(hash)
		if err != nil {
			continue
		}

		fileName := fmt.Sprintf("repoDump/%d-%s", rand.Int(), name)
		err = saveStringToDisc(content, fileName)
		if err != nil {
			continue
		}
	}

	return nil
}

func getCommitHashes(baseUrl string) ([]string, error) {
	resp, err := getFile(baseUrl, "/.git/logs/head")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyString := string(bodyBytes)
	lines := strings.Split(bodyString, "\n")
	hashes := make([]string, 0)

	for _, line := range lines {
		lineParts := strings.Split(line, " ")
		if len(lineParts) < 2 {
			continue
		}

		hashes = append(hashes, lineParts[1])
	}

	return hashes, nil
}

func getFile(baseUrl, urlPath string) (*http.Response, error) {
	resp, err := http.Get(baseUrl + urlPath)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func saveToDisc(body io.ReadCloser, pathToSave string) error {
	defer body.Close()

	dir := path.Dir(pathToSave)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(pathToSave)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	return err
}

func checkIfGitFolderCanBeDownloaded(baseUrl string) (bool, error) {
	resp, err := http.Head(baseUrl + "/.git")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 404 {
		return true, nil
	}

	return false, errors.New(".git Not found")
}

func initializeGitRepo() error {
	err := os.MkdirAll("repoDump", 0755)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = "repoDump"
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func getTreeHashesFromCommits(commitHashes []string) ([]string, error) {
	treeHashes := make([]string, 0)

	for _, commitHash := range commitHashes {
		cmd := exec.Command("git", "cat-file", "-p", commitHash)
		cmd.Dir = "repoDump"
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "tree") {
				treeHash := strings.Split(line, " ")[1]
				treeHashes = append(treeHashes, treeHash)
			}
		}
	}

	return treeHashes, nil
}

func getTreeHashesWithNameFromTrees(treeHashes []string) (map[string]string, error) {
	newTreeHashes := make(map[string]string)

	for _, treeHash := range treeHashes {
		cmd := exec.Command("git", "cat-file", "-p", treeHash)
		cmd.Dir = "repoDump"
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			re := regexp.MustCompile(`\s+`)
			splitedLine := re.Split(line, -1)
			if len(splitedLine) < 4 {
				continue
			}

			if splitedLine[1] == "tree" {
				newTreeHash := splitedLine[2]
				treeName := splitedLine[3]
				newTreeHashes[treeName] = newTreeHash
			}
		}
	}

	return newTreeHashes, nil
}

func getBlobHashesAndNamesFromTrees(treeHashes []string) (map[string]string, error) {
	blobHashesWithNames := make(map[string]string)

	for _, treeHash := range treeHashes {
		cmd := exec.Command("git", "cat-file", "-p", treeHash)
		cmd.Dir = "repoDump"
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			re := regexp.MustCompile(`\s+`)
			splitedLine := re.Split(line, -1)
			if len(splitedLine) < 4 {
				continue
			}

			if splitedLine[1] == "blob" {
				blobHash := splitedLine[2]
				blobName := splitedLine[3]
				blobHashesWithNames[blobHash] = blobName
			}
		}
	}

	return blobHashesWithNames, nil
}

func getFileContentFromBlob(blobHash string) (string, error) {
	cmd := exec.Command("git", "cat-file", "-p", blobHash)
	cmd.Dir = "repoDump"
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func saveStringToDisc(content, pathToSave string) error {
	dir := path.Dir(pathToSave)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(pathToSave)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func downloadAndSaveGitHubObjectFiles(baseUrl string, hashes []string) error {
	for _, hash := range hashes {
		resp, err := getFile(baseUrl, fmt.Sprintf("/.git/objects/%s/%s", hash[:2], hash[2:]))
		if err != nil {
			return err
		}

		err = saveToDisc(resp.Body, fmt.Sprintf("repoDump/.git/objects/%s/%s", hash[:2], hash[2:]))
		if err != nil {
			return err
		}
	}

	return nil
}

func convertMapToArray(hashes map[string]string) []string {
	result := make([]string, 0)

	for hash := range hashes {
		result = append(result, hash)
	}

	return result
}
