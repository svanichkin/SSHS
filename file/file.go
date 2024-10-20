package file

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func FindSSHFiles(rootPath string) (map[string]string, error) {

	sshFiles := make(map[string]string)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			sshFile := filepath.Join(path, "ssh")
			if _, err := os.Stat(sshFile); err == nil {
				file, err := os.Open(sshFile)
				if err != nil {
					return err
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				if scanner.Scan() {
					ip := scanner.Text()
					folderName := filepath.Base(path)
					sshFiles[folderName] = ip
				}
			}
		}
		return nil
	})

	return sshFiles, err

}

func FindPasswordFiles(rootPath string, devices []string) (map[string]map[string]string, error) {
	passwordFiles := make(map[string]map[string]string)

	contains := func(device string, devices []string) bool {
		for _, d := range devices {
			if d == device {
				return true
			}
		}
		return false
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			folderName := filepath.Base(filepath.Dir(path))

			if !contains(folderName, devices) {
				return nil
			}

			fileName := filepath.Base(path)

			if strings.HasPrefix(fileName, ".") {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			if scanner.Scan() {
				password := scanner.Text()
				if _, exists := passwordFiles[folderName]; !exists {
					passwordFiles[folderName] = make(map[string]string)
				}
				passwordFiles[folderName][fileName] = password
			}
		}
		return nil
	})

	return passwordFiles, err

}
