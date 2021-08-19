package main

import (
	"testing"
	"time"
)

func TestCreateRedisPool(t *testing.T) {
	CreateRedisPool(5)
	ping(1)
}

func TestWriteStartTime(t *testing.T) {
	CreateRedisPool(5)
	SetTransaction(1, "StartTime", time.Now().Format("2006-01-02 15:04:05Z"))
	SetTransaction(1, "errorCode", 1)
}
