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

func TryDumpGitRepo(baseUrl string) error {
	canBeDownloaded, err := checkIfGitFolderCanBeDownloaded(baseUrl)
	if !canBeDownloaded {
		return err
	}

	println("Git folder found, downloading files")
	println("Initializing git repo")

	err = InitializeGitRepo()
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

	treeHashes, err := GetTreeHashesFromCommits(commitHashes)
	if err != nil {
		return err
	}

	err = downloadAndSaveGitHubObjectFiles(baseUrl, treeHashes)
	if err != nil {
		return err
	}

	blobHashesWithName, err := GetBlobHashesAndNamesFromTrees(treeHashes)

	if err != nil {
		return err
	}

	blobHashes := helpers.ConvertMapToArray(blobHashesWithName)
	err = downloadAndSaveGitHubObjectFiles(baseUrl, blobHashes)
	if err != nil {
		return err
	}

	for hash, name := range blobHashesWithName {
		content, err := GetFileContentFromBlob(hash)
		if err != nil {
			continue
		}

		fileName := fmt.Sprintf("repoDump/%d-%s", rand.Int(), name)
		err = helpers.SaveStringToDisc(content, fileName)
		if err != nil {
			continue
		}
	}

	err = getFilesRecursively(baseUrl, treeHashes)
	if err != nil {
		return err
	}

	return nil
}

func getFilesRecursively(baseUrl string, treeHashes []string) error {
	treeHashesWithName, err := GetTreeHashesWithNameFromTrees(treeHashes)
	if err != nil {
		return err
	}

	treeHashesArr := helpers.ConvertMapToArray(treeHashesWithName)
	err = downloadAndSaveGitHubObjectFiles(baseUrl, treeHashesArr)
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
		blobHashesWithNameTemp, err := GetBlobHashesAndNamesFromTrees([]string{treeHash})
		if err != nil {
			return err
		}

		blobHashesTemp := helpers.ConvertMapToArray(blobHashesWithNameTemp)
		err = downloadAndSaveGitHubObjectFiles(baseUrl, blobHashesTemp)
		if err != nil {
			return err
		}

		for hash, name := range blobHashesWithNameTemp {
			content, err := GetFileContentFromBlob(hash)
			if err != nil {
				continue
			}

			fileName := fmt.Sprintf("repoDump/%s/%d-%s", treeName, rand.Int(), name)
			err = helpers.SaveStringToDisc(content, fileName)
			if err != nil {
				continue
			}
		}

		blobHashes := helpers.ConvertMapToArray(blobHashesWithNameTemp)
		err = downloadAndSaveGitHubObjectFiles(baseUrl, blobHashes)
		if err != nil {
			return err
		}

		treeHashesWithNameTemp, err := GetTreeHashesWithNameFromTrees([]string{treeHash})
		if err != nil {
			return err
		}

		for hash, name := range treeHashesWithNameTemp {
			treeHashesWithName[hash] = path.Join(treeName, name)
		}

		treeHashesArrTemp := helpers.ConvertMapToArray(treeHashesWithNameTemp)
		err = downloadAndSaveGitHubObjectFiles(baseUrl, treeHashesArr)
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

func downloadAndSaveGitHubObjectFiles(baseUrl string, hashes []string) error {
	for _, hash := range hashes {
		resp, err := helpers.GetFile(baseUrl, fmt.Sprintf("/.git/objects/%s/%s", hash[:2], hash[2:]))
		if err != nil {
			return err
		}

		err = helpers.SaveToDisc(resp.Body, fmt.Sprintf("repoDump/.git/objects/%s/%s", hash[:2], hash[2:]))
		if err != nil {
			return err
		}
	}

	return nil
}
