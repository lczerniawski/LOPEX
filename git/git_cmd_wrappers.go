package git

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func GetFileContentFromBlob(blobHash, outputFolder string) (string, error) {
	cmd := exec.Command("git", "cat-file", "-p", blobHash)
	cmd.Dir = outputFolder
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func GetBlobHashesAndNamesFromTrees(treeHashes []string, outputFolder string) (map[string]string, error) {
	blobHashesWithNames := make(map[string]string)

	for _, treeHash := range treeHashes {
		cmd := exec.Command("git", "cat-file", "-p", treeHash)
		cmd.Dir = outputFolder
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			re := regexp.MustCompile(`\s+`)
			splitLine := re.Split(line, -1)
			if len(splitLine) < 4 {
				continue
			}

			if splitLine[1] == "blob" {
				blobHash := splitLine[2]
				blobName := splitLine[3]
				blobHashesWithNames[blobHash] = blobName
			}
		}
	}

	return blobHashesWithNames, nil
}

func GetTreeHashesFromCommits(commitHashes []string, outputFolder string) ([]string, error) {
	treeHashes := make([]string, 0)

	for _, commitHash := range commitHashes {
		cmd := exec.Command("git", "cat-file", "-p", commitHash)
		cmd.Dir = outputFolder
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

func GetTreeHashesWithNameFromTrees(treeHashes []string, outputFolder string) (map[string]string, error) {
	newTreeHashes := make(map[string]string)

	for _, treeHash := range treeHashes {
		cmd := exec.Command("git", "cat-file", "-p", treeHash)
		cmd.Dir = outputFolder
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			re := regexp.MustCompile(`\s+`)
			splitLine := re.Split(line, -1)
			if len(splitLine) < 4 {
				continue
			}

			if splitLine[1] == "tree" {
				newTreeHash := splitLine[2]
				treeName := splitLine[3]
				newTreeHashes[newTreeHash] = treeName
			}
		}
	}

	return newTreeHashes, nil
}

func InitializeGitRepository(outputFolder string) error {
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = outputFolder
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
