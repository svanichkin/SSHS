package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/riywo/loginshell"
	"golang.org/x/crypto/ssh/terminal"
)

type Options struct {
	ExpectedPassword            string
	ExpectedKeyPass             string
	ExpectedFailure             string
	Timeout                     time.Duration
	AutoConfirmHostAuthenticity bool
	Shell                       string
}

var DefaultOptions = &Options{
	ExpectedPassword:            "word:",
	ExpectedKeyPass:             "':",
	ExpectedFailure:             "denied",
	Timeout:                     time.Second * 10,
	AutoConfirmHostAuthenticity: true,
	Shell:                       "",
}

func ParseAddress(address string) (user string, ip string, port string, err error) {

	ipv6Pattern := regexp.MustCompile(`^([a-zA-Z0-9\-]+)@\[([0-9a-fA-F:]+)\](?::(\d+))?$`)
	ipv6NoBracketPattern := regexp.MustCompile(`^([a-zA-Z0-9\-]+)@([0-9a-fA-F:]+)(?::(\d+))?$`)
	ipv4Pattern := regexp.MustCompile(`^([a-zA-Z0-9\-]+)@([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)(?::(\d+))?$`)
	domainPattern := regexp.MustCompile(`^([a-zA-Z0-9\-]+)@([a-zA-Z0-9\-\.]+)(?::(\d+))?$`)

	if matches := ipv6Pattern.FindStringSubmatch(address); matches != nil {
		user = matches[1]
		ip = matches[2]
		port = matches[3]
		if port == "" {
			port = "22"
		}
		return
	}

	if matches := ipv6NoBracketPattern.FindStringSubmatch(address); matches != nil {
		user = matches[1]
		ip = matches[2]
		port = matches[3]
		if port == "" {
			port = "22"
		}
		return
	}

	// Проверяем адрес на IPv4
	if matches := ipv4Pattern.FindStringSubmatch(address); matches != nil {
		user = matches[1]
		ip = matches[2]
		port = matches[3]
		if port == "" {
			port = "22"
		}
		return
	}

	if matches := domainPattern.FindStringSubmatch(address); matches != nil {
		user = matches[1]
		ip = matches[2]
		port = matches[3]
		if port == "" {
			port = "22"
		}
		return
	}

	err = fmt.Errorf("bad address: %v", address)
	return
}

func AddDevice(deviceName, sshAddress, configDevicePath, user, password, configPasswordPath string) error {

	deviceDir := filepath.Join(configDevicePath, deviceName)

	if _, err := os.Stat(deviceDir); os.IsNotExist(err) {
		err := os.MkdirAll(deviceDir, 0755)
		if err != nil {
			return fmt.Errorf("error create folder: %v", err)
		}
	}

	passwordDir := filepath.Join(configPasswordPath, deviceName)

	if _, err := os.Stat(passwordDir); os.IsNotExist(err) {
		err := os.MkdirAll(passwordDir, 0755)
		if err != nil {
			return fmt.Errorf("error create folder: %v", err)
		}
	}

	sshFilePath := filepath.Join(deviceDir, "ssh.txt")

	err := os.WriteFile(sshFilePath, []byte(sshAddress), 0644)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}

	fmt.Printf("SSH saved %s\n", sshFilePath)

	passwordFilePath := filepath.Join(deviceDir, user)

	err = os.WriteFile(passwordFilePath, []byte(password), 0644)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}

	fmt.Printf("Password saved %s\n", passwordFilePath)

	return nil

}

func ConnectSSH(address, password string) {

	options := DefaultOptions

	user, ip, port, err := ParseAddress(address)
	if err != nil {
		fmt.Println("Error parsing address:", err)
		return
	}

	cmdStr := fmt.Sprintf("ssh -p %s %s@%s", port, user, ip)
	fmt.Println("Executing command:", cmdStr)

	err = run(cmdStr, password, options)
	if err != nil {
		fmt.Println("Failed to execute SSH command:", err)
		return
	}

	fmt.Println("Password entered, authentication successful.")
}

func enterPassword(pt *os.File, password string, options *Options, redirectPipes bool) (string, error) {

	errChan := make(chan error)
	readyChan := make(chan string)

	go func() {
		buf := make([]byte, 4096)
		confirmed := false
		entered := false
		var data string
		for {
			n, err := pt.Read(buf)
			if err != nil {
				errChan <- err
				break
			}
			if n == 0 {
				continue
			}
			data += string(buf[:n])
			if !confirmed && strings.Contains(data, "The authenticity of host ") {
				if options.AutoConfirmHostAuthenticity {
					confirmed = true
					data = ""
					pt.Write([]byte("yes\n"))
				} else {
					errChan <- fmt.Errorf("host authenticity confirmation required, but it was disabled")
					break
				}
			} else if !entered && (strings.Contains(data, options.ExpectedPassword) || strings.Contains(data, options.ExpectedKeyPass)) {
				entered = true
				data = ""
				pt.Write([]byte(password + "\n"))
			} else if entered && len(data) > 5 {
				if strings.Contains(data, options.ExpectedPassword) || strings.Contains(data, options.ExpectedKeyPass) || strings.Contains(data, options.ExpectedFailure) {
					errChan <- fmt.Errorf("authentication failure")
					break
				}
				readyChan <- data
				break
			}
		}
	}()

	timer := time.NewTimer(options.Timeout)
	defer timer.Stop()

	select {
	case newBuffered := <-readyChan:
		if redirectPipes {
			os.Stdout.WriteString(newBuffered)
			go func() { _, _ = io.Copy(pt, os.Stdin) }()
			_, _ = io.Copy(os.Stdout, pt)
			return "", nil
		}
		return newBuffered, nil
	case err := <-errChan:
		return "", err
	case <-timer.C:
		return "", fmt.Errorf("timed out waiting for prompt")
	}

}

func run(cmd string, password string, options *Options) error {

	if options == nil {
		options = DefaultOptions
	}

	shell := options.Shell

	if shell == "" {
		var err error
		shell, err = loginshell.Shell()
		if err != nil {
			shell = "/bin/bash"
		}
	}

	c := exec.Command(shell)

	pt, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer func() { _ = pt.Close() }()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, pt); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGHUP

	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }()

	if _, err := pt.Write([]byte(cmd + "; exit\n")); err != nil {
		return err
	}

	_, err = enterPassword(pt, password, options, true)
	if err != nil {
		return err
	}

	return nil

}
