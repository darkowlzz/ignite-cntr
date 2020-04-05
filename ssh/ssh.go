package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSH constants.
const (
	DefaultNetwork = "tcp"
	DefaultPort    = "22"
	DefaultTimeout = uint32(1)
)

// NewSSHClient creates and returns a ssh client with an active connection.
// The consumer of the client must ensure that the connection is closed.
func NewSSHClient(ip, user, privateKeyFile string) (*ssh.Client, error) {
	// Ensure that the private key is an existing file.
	info, err := os.Stat(privateKeyFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("identity file not found: %v", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("identity file path is a directory, want to be a file: %v", err)
	}

	signer, err := newSignerForKey(privateKeyFile)
	if err != nil {
		return nil, err
	}

	config := newSSHConfig(user, signer, DefaultTimeout)

	// Start a ssh connection.
	sshClient, err := ssh.Dial(DefaultNetwork, net.JoinHostPort(ip, DefaultPort), config)
	if err != nil {
		return nil, err
	}
	return sshClient, nil
}

// newSignerForKey takes a file path of a private key and returns a signer for
// for it.
func newSignerForKey(keyPath string) (ssh.Signer, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	return ssh.ParsePrivateKey(key)
}

// newSSHConfig takes ssh configurations and returns a new ssh config.
func newSSHConfig(user string, publicKey ssh.Signer, timeout uint32) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(publicKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: use ssh.FixedPublicKey instead
		Timeout:         time.Second * time.Duration(timeout),
	}
}

// RunSSHCommand runs command using a given sh client and returns the output and
// error from the command execution.
func RunSSHCommand(ip, user, privateKeyFile, command string) ([]byte, []byte, error) {
	// Create a new SSH Client.
	client, err := NewSSHClient(ip, user, privateKeyFile)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()

	var cmdOut, cmdErr bytes.Buffer

	// Create a session for the command
	session, err := client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = &cmdOut
	session.Stderr = &cmdErr

	if err := session.Run(command); err != nil {
		return nil, nil, err
	}

	return cmdOut.Bytes(), cmdErr.Bytes(), nil
}

// StartSSHCommand runs command and doesn't wait for the command to complete.
func StartSSHCommand(ip, user, privateKeyFile, command string) error {
	// Create a new SSH Client.
	client, err := NewSSHClient(ip, user, privateKeyFile)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create a session for the command
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	return session.Run(command)
}
