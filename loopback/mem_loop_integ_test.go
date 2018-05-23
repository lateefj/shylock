// +build integration

package loopback

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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
		os.MkdirAll(TestFuseMemoryLoopbackKVPath, 0755)
	}
	err := shylock.MountFuse(TestFuseMemoryLoopbackKVPath, FSMemoryLoopbacKV, []byte("Empty Config"))
	if err != nil {
		t.Fatalf("Failed to mount with error %s", err)
	}
	// Cheep way to unmount eveything
	defer shylock.Exit()
	testFileName := "first_test.txt"
	testPath := fmt.Sprintf("%s/%s", TestFuseMemoryLoopbackKVPath, testFileName)
	f, err := os.Open(testPath)
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

	// TODO:
	// 3. Write to a file in the path
	// 4. Verify that the data is in memory
	// 5. Read from the path and make sure the values match.

}
