package main

import (
	"os"
	"os/exec"
	"testing"
)

func test(t *testing.T, name string) {
	cmd := exec.Command("go", "run", "./"+name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
}

func TestGettingStarted(t *testing.T) {
	test(t, "getting_started")
}

func TestOfflineProcessing(t *testing.T) {
	test(t, "offline_processing")
}

func TestReloadFromFile(t *testing.T) {
	test(t, "reload_from_file")
}

func TestStronglyTyped(t *testing.T) {
	test(t, "update_polling_interval")
}
