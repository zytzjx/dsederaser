package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestCreateFile(t *testing.T) {
	f, err := os.OpenFile("logs/test.txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Error("file error")
		return
	}
	f.WriteString("fhdjsafhdsalfd\n")
	f.WriteString("fdasfdsafdsafdsafdsafdsafds\n")
	f.WriteString("fdasfdsafdsafdsafdsafdsafds\n")
	f.WriteString("fdasfdsafdsafdsafdsafdsafds\n")
	f.WriteString("fdasfdsafdsafdsafdsafdsafds\n")
	f.WriteString("fdasfdsafdsafdsafdsafdsafds\n")
	f.Close()

	ff, err := os.OpenFile("logs/test.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		t.Error("file error")
		return
	}
	ff.WriteString("11111111fhdjsafhdsalfd\n")
	ff.WriteString("2222222fdasfdsafdsafdsafdsafdsafds\n")
	ff.Close()
}

func TestHandlelogprogress(t *testing.T) {
	CreateRedisPool(5)
	line := `   1      1 0xff   0.159%   0.159% 00:00:05 00:00:05 10:38:01 00003141    93.18    93.18`
	label := 1
	handlelogprogress(label, line)
}

func TestStringFind(t *testing.T) {
	data := []byte("/dev/sde:\n	Issuing command\n	Operation started in background\n	You may use `--sanitize-status` to check progress")
	aa := string(data)
	if !strings.Contains(aa, "is not supported") {
		t.Log("Success")
	} else {
		t.Error("failed")
	}
	if !strings.ContainsAny(aa, "is not supported") {
		t.Log("Success")
	} else {
		t.Error("failed")
	}
}

func TestRegex(t *testing.T) {
	file, err := os.Open("log.txt")
	if err != nil {
		t.Fatalf("failed to open")

	}
	// The method os.File.Close() is called
	// on the os.File object to close the file
	defer file.Close()

	// The bufio.NewScanner() function is called in which the
	// object os.File passed as its parameter and this returns a
	// object bufio.Scanner which is further used on the
	// bufio.Scanner.Split() method.
	scanner := bufio.NewScanner(file)

	// The bufio.ScanLines is used as an
	// input to the method bufio.Scanner.Split()
	// and then the scanning forwards to each
	// new line using the bufio.Scanner.Scan()
	// method.
	scanner.Split(bufio.ScanLines)
	var validlog = regexp.MustCompile(`^\W+\d+\W+\d+\W+.*?(\d*\.\d*)%\W+(\d*\.\d*)%.*?(\d*\.\d*)$`)

	//r := regexp.MustCompile(`^\W+\d+\W+(\d+)\W+.*?(\d*\.\d*)%\W+(\d*\.\d*)%\W+(\d{2}:\d{2}:\d{2})\W+(\d{2}:\d{2}:\d{2})\W+(\d{2}:\d{2}:\d{2})\W+(\d+)\W+(\d*\.\d*)\W+(\d*\.\d*)$`)
	//fmt.Printf("%#v\n", r.FindStringSubmatch(`   1      1 0xff   0.180%   0.180% 00:00:05 00:00:05 16:28:56 00002782   107.33   107.33`))

	for scanner.Scan() {
		line := scanner.Text()
		if validlog.MatchString(line) {
			fmt.Println("Match:" + line)

			t.Log(line)
		}
	}

}
