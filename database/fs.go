package database

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJSONFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlockDBFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "block.db")
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func writeEmptyBlockDBToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}

func initDataDirIfNotExists(dataDir string) error {
	if fileExist(getGenesisJSONFilePath(dataDir)) {
		return nil
	}

	if err := os.MkdirAll(getDatabaseDirPath(dataDir), os.ModePerm); err != nil {
		return fmt.Errorf("error while creating database directory: %w", err)
	}

	if err := writeGenesisToDisk(getGenesisJSONFilePath(dataDir)); err != nil {
		return fmt.Errorf("error while writing genesis block to file: %w", err)
	}

	if err := writeEmptyBlockDBToDisk(getBlockDBFilePath(dataDir)); err != nil {
		return fmt.Errorf("error while creating block file: %w", err)
	}

	return nil
}
