package main

import (
	"strings"
	"testing"
)

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
