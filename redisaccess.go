package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})
var clients map[int]*redis.Client

//GetLabelsCnt get labels
func GetLabelsCnt() int {
	labels, err := rdb.Get(ctx, "labelcnt").Int()
	if err != nil {
		labels = 48
		return labels
	}
	return labels
}

// CreateRedisPool create redis client
func CreateRedisPool(count int) (map[int]*redis.Client, error) {
	clients = make(map[int]*redis.Client)
	clients[0] = rdb
	for i := 1; i <= count; i++ {
		var c = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       i,  // use default DB
		})

		clients[i] = c
	}
	return clients, nil
}

// PublishTaskDone publis for other Subscribe
func PublishTaskDone(label int, value interface{}) (int64, error) {
	client, ok := clients[label]
	if !ok {
		return 0, errors.New("not found label")
	}
	aa := map[string]interface{}{"taskdone": value}
	rsProcess, _ := json.Marshal(map[string]interface{}{"label": label, "msg": aa})
	return client.Publish(ctx, "progress", rsProcess).Result()
}

// Publish publis for other Subscribe
func Publish(label int, channel string, msg interface{}) (int64, error) {
	client, ok := clients[label]
	if !ok {
		return 0, errors.New("not found label")
	}
	return client.Publish(ctx, channel, msg).Result()
}

func ping(label int) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	fmt.Println(pong, err)
	return nil
}

func GetTaskStatus(label int) bool {
	client, ok := clients[label]
	if !ok {
		return false
	}
	ss, err := client.Get(ctx, "processingstatus").Result()
	if err != nil {
		return false
	}
	if ss == "start" {
		return true
	}
	return false
}

// SetTransaction set into transaction hash map
func SetTransaction(label int, key string, value interface{}) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	kv := make(map[string]interface{})
	kv[key] = value
	return client.HSet(ctx, "transaction", kv).Err()
}

// GetTransaction Transaction
func GetTransaction(label int) (map[string]string, error) {
	client, ok := clients[label]
	if !ok {
		return nil, errors.New("not found label")
	}
	result, err := client.HGetAll(ctx, "transaction").Result()
	if err != nil {
		return nil, err
	}
	if err == redis.Nil {
		return map[string]string{}, nil
	}
	return result, nil
}

// Set set key value to Redis{key:value}
func Set(label int, key string, value interface{}, expiration time.Duration) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	return client.Set(ctx, key, value, expiration).Err()
}

// GetString get value
func GetString(label int, key string) (string, error) {
	client, ok := clients[label]
	if !ok {
		return "", errors.New("not found label")
	}
	return client.Get(ctx, key).Result()
}

// GetInt get Int
func GetInt(label int, key string) (int, error) {
	client, ok := clients[label]
	if !ok {
		return 0, errors.New("not found label")
	}
	return client.Get(ctx, key).Int()
}

func setErrorDone(label int) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	kv := make(map[string]interface{})
	kv["progress"] = "100.00%"
	rsProcess, _ := json.Marshal(map[string]interface{}{"label": label, "msg": kv})
	Publish(label, "progress", rsProcess)

	return client.HSet(ctx, "processing", kv).Err()
}

func setProgressbar(label int, values []string) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	if len(values) < 9 {
		return errors.New("parser progress line error")
	}
	//[1 1 0x00 98.644% 98.64% 00:28 00:28:40 21:27:57 00:29 172 MB/s 172.05]
	kv := make(map[string]interface{})
	kv["speed"] = values[9]
	if values[5] == "" {
		values[5] = "00:00"
	}
	kv["time"] = values[5]
	kv["est"] = values[8]
	kv["progress"] = values[4]
	kv["optime"] = fmt.Sprintf("%s/%s", kv["time"], kv["est"])

	rsProcess, _ := json.Marshal(map[string]interface{}{"label": label, "msg": kv})
	Publish(label, "progress", rsProcess)

	return client.HSet(ctx, "processing", kv).Err()
}

// GetFloat get float
func GetFloat(label int, key string) (float32, error) {
	client, ok := clients[label]
	if !ok {
		return 0.0, errors.New("not found label")
	}
	return client.Get(ctx, key).Float32()
}

// GetTime get time
func GetTime(label int, key string) (time.Time, error) {
	client, ok := clients[label]
	if !ok {
		return time.Now(), errors.New("not found label")
	}
	return client.Get(ctx, key).Time()
}

// AddSet for sets to redis
func AddSet(label int, key string, values ...interface{}) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	return client.SAdd(ctx, key, values...).Err()
}

// Del remove keys
func Del(label int, key ...string) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	return client.Del(ctx, key...).Err()
}

//FlushDB clean current database
func FlushDB(label int) error {
	client, ok := clients[label]
	if !ok {
		return errors.New("not found label")
	}
	return client.FlushDB(ctx).Err()

}
