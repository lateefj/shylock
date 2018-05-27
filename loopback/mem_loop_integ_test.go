// +build integration

package loopback

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/lateefj/shylock"
)

const (
	TestFuseMemoryLoopbackKVPath = "tmp/shylock/loopbackkv_test"
)

func TestFuseMemoryLoopbackKV(t *testing.T) {
	// Create directory if it doesn't exist
	if _, err := os.Stat(TestFuseMemoryLoopbackKVPath); os.IsNotExist(err) {
		fmt.Printf("Trying to create path %s\n", TestFuseMemoryLoopbackKVPath)
		err = os.MkdirAll(TestFuseMemoryLoopbackKVPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s error: %s", TestFuseMemoryLoopbackKVPath, err)
		}
	}
	err := shylock.MountFuse(TestFuseMemoryLoopbackKVPath, FSMemoryLoopbacKV, []byte("Empty Config"))
	if err != nil {
		t.Fatalf("Failed to mount with error %s", err)
	}
	// Cheep way to unmount everything
	defer shylock.Exit()
	testFileName := "first_test.txt"
	testPath := fmt.Sprintf("%s/%s", TestFuseMemoryLoopbackKVPath, testFileName)
	f, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("Failed to create file %s", err)
	}
	testData := []byte("Test Data")
	f.Write(testData)
	f.Close()

	readFile, err := os.Open(testPath)
	if err != nil {
		t.Fatalf("Failed to open file %s with err %s", testPath, err)
	}

	readData, err := ioutil.ReadAll(readFile)
	if err != nil {
		t.Fatalf("Failed to read file %s with error %s", testPath, err)
	}

	if bytes.Compare(testData, readData) != 0 {
		t.Fatalf("test data %s does not match read data %s", string(testData), string(readData))
	}

	var files []string
	fileExists := false
	err = filepath.Walk(TestFuseMemoryLoopbackKVPath, func(path string, info os.FileInfo, err error) error {

		if path == testPath {
			fileExists = true
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Error listing directory files %s", err)
	}

	if !fileExists {
		t.Fatalf("Expected file %s in directory list %s\n", testPath, files)
	}

	err = os.Remove(testPath)
	if err != nil {
		t.Fatalf("Failed to remove %s from %s error %s\n", testPath, files, err)
	}
	fileExists = false
	err = filepath.Walk(TestFuseMemoryLoopbackKVPath, func(path string, info os.FileInfo, err error) error {

		if path == testPath {
			fileExists = true
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Error listing directory files %s", err)
	}
	if fileExists {
		t.Fatalf("Didn't expect file %s in directory list %s\n", testPath, files)
	}

	// TODO:
	// 4. Verify that the data is in memory
	// 5. Read from the path and make sure the values match.

}
