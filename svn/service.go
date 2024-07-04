package svn

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/lczerniawski/LOPEX/helpers"
)

func TryDownloadSvnRepository(baseUrl, outputFolder string) error {
	canBeDownloaded, err := helpers.CheckIfFolderExists(baseUrl, ".svn")
	if !canBeDownloaded {
		return err
	}
	println("SVN folder found.")

	err = downloadSvnDB(baseUrl, outputFolder)
	if err != nil {
		return err
	}

	db, err := openRepositoryDb(outputFolder)
	if err != nil {
		return err
	}
	defer db.Close()

	err = readRepositoryInfo(db)
	if err != nil {
		return err
	}

	println("Download of SVN repository started")
	filesMap, err := readRepositoryFiles(db)
	if err != nil {
		return err
	}

	for path, fileHash := range filesMap {
		if path == "" {
			continue
		}

		if fileHash == "" {
			os.MkdirAll(outputFolder+"/"+path, os.ModePerm)
			continue
		}

		hashWithoutSha := strings.Replace(fileHash, "$sha1$", "", -1)
		fileUrl := "/.svn/pristine/" + hashWithoutSha[0:2] + "/" + hashWithoutSha + ".svn-base"
		fileContent, err := helpers.GetFile(baseUrl, fileUrl)
		if err != nil {
			return err
		}

		err = helpers.SaveToDisc(fileContent.Body, outputFolder+"/"+path)
		if err != nil {
			return nil
		}
	}

	println("Download succeeded")
	fmt.Printf("All files can be found under %s", outputFolder)

	return nil
}

func downloadSvnDB(baseUrl, outputFolder string) error {
	file, err := helpers.GetFile(baseUrl, "/.svn/wc.db")
	if err != nil {
		return err
	}

	err = helpers.SaveToDisc(file.Body, outputFolder+"/wc.db")

	if err != nil {
		return err
	}

	return nil
}

func openRepositoryDb(outputFolder string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", outputFolder+"/wc.db")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readRepositoryInfo(db *sql.DB) error {
	rows, err := db.Query(`SELECT root FROM REPOSITORY`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var root string
		err = rows.Scan(&root)
		if err != nil {
			return err
		}
		println("SVN Repository server URL:", root)
	}

	return nil
}

func readRepositoryFiles(db *sql.DB) (map[string]string, error) {
	result := map[string]string{}

	rows, err := db.Query(`SELECT local_relpath, checksum FROM NODES`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var local_relpath string
		var checksum sql.NullString
		err = rows.Scan(&local_relpath, &checksum)
		if err != nil {
			return nil, err
		}

		result[local_relpath] = checksum.String
	}

	return result, nil
}
