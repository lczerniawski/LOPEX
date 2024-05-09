package git

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func GetFileContentFromBlob(blobHash string) (string, error) {
	cmd := exec.Command("git", "cat-file", "-p", blobHash)
	cmd.Dir = "repoDump"
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func GetBlobHashesAndNamesFromTrees(treeHashes []string) (map[string]string, error) {
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

func GetTreeHashesFromCommits(commitHashes []string) ([]string, error) {
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

func GetTreeHashesWithNameFromTrees(treeHashes []string) (map[string]string, error) {
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
				newTreeHashes[newTreeHash] = treeName
			}
		}
	}

	return newTreeHashes, nil
}

func InitializeGitRepo() error {
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
