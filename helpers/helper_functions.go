package helpers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
)

func ConvertMapToArray(hashes map[string]string) []string {
	result := make([]string, 0)

	for hash := range hashes {
		result = append(result, hash)
	}

	return result
}

func GetFile(baseUrl, urlPath string) (*http.Response, error) {
	resp, err := http.Get(baseUrl + urlPath)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func SaveToDisc(body io.ReadCloser, pathToSave string) error {
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

func SaveStringToDisc(content, pathToSave string) error {
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

func CheckIfFolderExists(baseUrl, folderName string) (bool, error) {
	resp, err := http.Head(baseUrl + "/" + folderName)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != 404 {
		return true, nil
	}

	return false, errors.New(folderName + " Not found")
}
