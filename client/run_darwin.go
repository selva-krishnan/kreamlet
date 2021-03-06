package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Run(sshPort string, kubePort string, cpus string, memory string, disk string) error {

	// Checking for hostfile entry
	err := Hostfile()
	if err != nil {
		fmt.Println(err)
	}

	// Checking for linuxkit binary
	binary, lookErr := exec.LookPath("linuxkit")
	if lookErr != nil {
		fmt.Println("downloading linuxkit binary")
		fileUrl := "https://github.com/linuxkit/linuxkit/releases/download/v0.5/linuxkit-darwin-amd64"
		dest := "/usr/local/bin/linuxkit"
		err := DownloadFile(fileUrl, dest)
		if err != nil {
			fmt.Println("there has been an error downloading the linuxkit binary")
			os.Exit(1)
		}

		chx := os.Chmod(dest, 0755)
		if chx != nil {
			fmt.Println("could not change permissions on the linuxkit binary")
			os.Exit(1)
		}

		binary = dest
	}

	// Getting users home dir to use later
	homedir := os.Getenv("HOME")

	// Removing old state if the run function is called
	if _, err := os.Stat(homedir + "/.kream/kube-master-state"); err != nil {
		fmt.Println("creating a new cluster state directory")
		mk := os.MkdirAll(homedir+"/.kream/kube-master-state", 0700)
		if mk != nil {
			return err
		}
		_, err := os.OpenFile(homedir+"/.kream/kube-master-state/metadata.json", os.O_RDONLY|os.O_CREATE, 0700)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("removing the old cluster state")
		// We need to recreate the state folder to add the metadata.json file
		rm := os.RemoveAll(homedir + "/.kream/kube-master-state")
		if rm != nil {
			return rm
		}
		mk := os.MkdirAll(homedir+"/.kream/kube-master-state", 0700)
		if mk != nil {
			return mk
		}
		_, err := os.OpenFile(homedir+"/.kream/kube-master-state/metadata.json", os.O_RDONLY|os.O_CREATE, 0700)
		if err != nil {
			return err
		}
	}

	// Check if iso is already downloaded
	if _, err := os.Stat(homedir + "/.kream/kube-master-efi.iso"); os.IsNotExist(err) {
		fileUrl := "https://s3.amazonaws.com/puppet-cloud-and-containers/kream/kube-master-efi.iso"
		dest := (homedir + "/.kream/" + filepath.Base(fileUrl))
		fmt.Println("downloading the kubernetes iso")
		err := DownloadFile(fileUrl, dest)
		if err != nil {
			panic(err)
		}

	}

	// Check for .ssh folder in the kream directory

	if _, err := os.Stat(homedir + "/.kream/ssh"); err != nil {
		mk := os.MkdirAll(homedir+"/.kream/ssh", 0700)
		if mk != nil {
			return err
		}
	}

	// Check if ssh keys are already there
	if _, err := os.Stat(homedir + "/.kream/ssh/id_rsa"); os.IsNotExist(err) {
		fileUrl := "https://s3.amazonaws.com/puppet-cloud-and-containers/kream/id_rsa"
		dest := (homedir + "/.kream/ssh/" + filepath.Base(fileUrl))
		err := DownloadFile(fileUrl, dest)
		if err != nil {
			panic(err)
		}
	}

	//TODO vendor in Linuxkit
	args := []string{
		"run",
		"hyperkit",
		"-publish", sshPort + ":22",
		"-publish", kubePort + ":6443",
		"-publish", "50091:50091",
		"-networking", "default",
		"-cpus", cpus,
		"-mem", memory,
		"-state", homedir + "/.kream/kube-master-state",
		"-disk", "size=" + disk,
		"-data-file", homedir + "/.kream/kube-master-state/metadata.json",
		"--uefi",
		"-iso", homedir + "/.kream/kube-master-efi.iso"}

	// This uses exec.Command func to pass the above args to run linuxkit
	execErr := exec.Command(binary, args...).Start()
	if execErr != nil {
		fmt.Println("failed here")
		fmt.Println(execErr)
	}
	return nil

}
