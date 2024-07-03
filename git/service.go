package git

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path"
	"strings"

	"github.com/lczerniawski/LOPEX/helpers"
)

func TryDownloadGitRepository(baseUrl, outputFolder string) error {
	canBeDownloaded, err := checkIfGitFolderExists(baseUrl)
	if !canBeDownloaded {
		return err
	}
	println("Git folder found.")

	println("Initializing git repository.")
	err = InitializeGitRepository(outputFolder)
	if err != nil {
		return err
	}

	println("Downloading of git repository started.")
	commitHashes, err := getCommitHashes(baseUrl)
	if err != nil {
		return err
	}

	err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, commitHashes)
	if err != nil {
		return err
	}

	treeHashes, err := GetTreeHashesFromCommits(commitHashes, outputFolder)
	if err != nil {
		return err
	}

	err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, treeHashes)
	if err != nil {
		return err
	}

	blobHashesWithName, err := GetBlobHashesAndNamesFromTrees(treeHashes, outputFolder)

	if err != nil {
		return err
	}

	blobHashes := helpers.ConvertMapToArray(blobHashesWithName)
	err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, blobHashes)
	if err != nil {
		return err
	}

	for hash, name := range blobHashesWithName {
		content, err := GetFileContentFromBlob(hash, outputFolder)
		if err != nil {
			continue
		}

		fileName := fmt.Sprintf("%s/%d-%s", outputFolder, rand.Int(), name)
		err = helpers.SaveStringToDisc(content, fileName)
		if err != nil {
			continue
		}
	}

	err = getFilesRecursively(baseUrl, outputFolder, treeHashes)
	if err != nil {
		return err
	}

	println("Download succeeded")
	fmt.Printf("All files can be found under %s", outputFolder)
	return nil
}

func getFilesRecursively(baseUrl, outputFolder string, treeHashes []string) error {
	treeHashesWithName, err := GetTreeHashesWithNameFromTrees(treeHashes, outputFolder)
	if err != nil {
		return err
	}

	treeHashesArr := helpers.ConvertMapToArray(treeHashesWithName)
	err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, treeHashesArr)
	if err != nil {
		return err
	}

	for {
		// Check if there are any more folders to download
		if len(treeHashesArr) == 0 {
			break
		}

		treeHash := treeHashesArr[0]
		treeName := treeHashesWithName[treeHash]
		treeHashesArr = treeHashesArr[1:]

		// Get all the files inside the folders
		blobHashesWithNameTemp, err := GetBlobHashesAndNamesFromTrees([]string{treeHash}, outputFolder)
		if err != nil {
			return err
		}

		blobHashesTemp := helpers.ConvertMapToArray(blobHashesWithNameTemp)
		err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, blobHashesTemp)
		if err != nil {
			return err
		}

		for hash, name := range blobHashesWithNameTemp {
			content, err := GetFileContentFromBlob(hash, outputFolder)
			if err != nil {
				continue
			}

			fileName := fmt.Sprintf("%s/%s/%d-%s", outputFolder, treeName, rand.Int(), name)
			err = helpers.SaveStringToDisc(content, fileName)
			if err != nil {
				continue
			}
		}

		blobHashes := helpers.ConvertMapToArray(blobHashesWithNameTemp)
		err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, blobHashes)
		if err != nil {
			return err
		}

		treeHashesWithNameTemp, err := GetTreeHashesWithNameFromTrees([]string{treeHash}, outputFolder)
		if err != nil {
			return err
		}

		for hash, name := range treeHashesWithNameTemp {
			treeHashesWithName[hash] = path.Join(treeName, name)
		}

		treeHashesArrTemp := helpers.ConvertMapToArray(treeHashesWithNameTemp)
		err = downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder, treeHashesArr)
		if err != nil {
			return err
		}

		treeHashesArr = append(treeHashesArr, treeHashesArrTemp...)
	}

	return nil
}

func getCommitHashes(baseUrl string) ([]string, error) {
	resp, err := helpers.GetFile(baseUrl, "/.git/logs/head")
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

func checkIfGitFolderExists(baseUrl string) (bool, error) {
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

func downloadAndSaveGitHubObjectFiles(baseUrl, outputFolder string, hashes []string) error {
	for _, hash := range hashes {
		resp, err := helpers.GetFile(baseUrl, fmt.Sprintf("/.git/objects/%s/%s", hash[:2], hash[2:]))
		if err != nil {
			return err
		}

		err = helpers.SaveToDisc(resp.Body, fmt.Sprintf("%s/.git/objects/%s/%s", outputFolder, hash[:2], hash[2:]))
		if err != nil {
			return err
		}
	}

	return nil
}
