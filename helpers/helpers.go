package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/s00500/env_logger"
)

// ##########################
// #### Helper Functions ####
// ##########################

// Create a folder on disk if it dows not already exsist
func CreateFolder(path string, folder string) string {
	folderPath := strings.Join([]string{path, folder}, "/")

	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(folderPath, 0755)
		if errDir != nil {
			log.Should(err)
		}
	}
	return folder
}

// Delete all folders and files in a folder
func DeleteFolder(path string, folder string) (err error) {
	folderPath := strings.Join([]string{path, folder}, "/")

	err = os.RemoveAll(folderPath)
	if err != nil {
		return err
	}

	return nil
}

// Delete a saved file
func DeleteFile(path string, fileName string) (err error) {
	filePath := strings.Join([]string{path, fileName}, "/")

	err = os.Remove(filePath)
	if err == nil {
		return err
	}

	return nil
}

// updates the Json
func UpdateJson(data interface{}, path string, fileName string) {
	filePath := strings.Join([]string{path, fileName}, "/")

	file, err := json.Marshal(data)
	if !log.Should(err) {
		err = ioutil.WriteFile(filePath, file, 0755)
		log.Should(err)
	}
}

// reads the Json and creates a new if not found
func ReadJson(data interface{}, path string, fileName string) {
	filePath := strings.Join([]string{path, fileName}, "/")

	file, err := os.ReadFile(filePath) // Read File
	log.Debug(err)

	if err == nil {
		err = json.Unmarshal(file, data)
		log.Debug(err)
	} else {
		// Write a new file
		file, err := json.Marshal(data)
		log.Debug(err)
		if err == nil {
			log.Should(ioutil.WriteFile(filePath, file, 0644))
		}
	}
}
