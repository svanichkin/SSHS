package main

import (
	"fmt"
	"main/conf"
	"main/file"
	"main/menu"
	"main/ssh"
	"os"
)

const sampleArgs = "sshs device_name user@ip_address:port password"

func main() {

	config, err := conf.Init()
	if err != nil {
		fmt.Println("Failed to initialize config:", err)
		return
	}

	if len(os.Args) > 1 {
		if len(os.Args) != 3 {
			fmt.Println("Waiting for: " + sampleArgs)
			os.Exit(1)
		}

		deviceName := os.Args[1]
		sshAddress := os.Args[2]
		password := os.Args[3]

		user, _, _, err := ssh.ParseAddress(sshAddress)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = ssh.AddDevice(deviceName, sshAddress, config.Devices, user, password, config.Passwords)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	sshFiles, err := file.FindSSHFiles(config.Devices)
	if err != nil {
		fmt.Println("Please add new device: " + sampleArgs)
		return
	}

	passwordFiles, err := file.FindPasswordFiles(config.Passwords, getKeys(sshFiles))
	if err != nil {
		fmt.Println("Please add new device: " + sampleArgs)
		return
	}

	selectedDevice := menu.ShowMenu(sshFiles)
	if selectedDevice != "" {
		selectedAddress := sshFiles[selectedDevice]

		user, _, _, err := ssh.ParseAddress(selectedAddress)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ssh.ConnectSSH(selectedAddress, passwordFiles[selectedDevice][user])
	}

}

func getKeys(m map[string]string) []string {

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys

}
