package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const ScanQueue = "quantumfield:scan_jobs"

type ScanJob struct {
	ScanID  string `json:"scan_id"`
	AssetID string `json:"asset_id"`
	UserID  string `json:"user_id"`
	Attempt int    `json:"attempt"`
}

type Client struct {
	redis *redis.Client
}

func New(address, password string, database int) *Client {
	return &Client{redis: redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       database,
	})}
}

func (client *Client) Ping(context context.Context) error {
	return client.redis.Ping(context).Err()
}

func (client *Client) Enqueue(context context.Context, job ScanJob) error {
	payload, err := encodeScanJob(job)
	if err != nil {
		return fmt.Errorf("marshal scan job: %w", err)
	}
	return client.redis.LPush(context, ScanQueue, payload).Err()
}

func (client *Client) Dequeue(context context.Context, timeout time.Duration) (*ScanJob, error) {
	result, err := client.redis.BRPop(context, timeout, ScanQueue).Result()
	if err != nil {
		return nil, err
	}
	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected redis response")
	}
	job, err := decodeScanJob([]byte(result[1]))
	if err != nil {
		return nil, fmt.Errorf("decode scan job: %w", err)
	}
	return job, nil
}

func (client *Client) Close() error {
	return client.redis.Close()
}

func (client *Client) Allow(context context.Context, key string, limit int64, window time.Duration) (bool, int64, time.Duration, error) {
	script := redis.NewScript(`
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then
			redis.call("PEXPIRE", KEYS[1], ARGV[1])
		end
		local ttl = redis.call("PTTL", KEYS[1])
		return {current, ttl}
	`)
	result, err := script.Run(context, client.redis, []string{"quantumfield:rate:" + key}, window.Milliseconds()).Int64Slice()
	if err != nil {
		return false, 0, 0, err
	}
	if len(result) != 2 {
		return false, 0, 0, fmt.Errorf("unexpected rate-limit response")
	}
	remaining := limit - result[0]
	if remaining < 0 {
		remaining = 0
	}
	retryAfter := time.Duration(result[1]) * time.Millisecond
	return result[0] <= limit, remaining, retryAfter, nil
}

func encodeScanJob(job ScanJob) ([]byte, error) {
	return json.Marshal(job)
}

func decodeScanJob(payload []byte) (*ScanJob, error) {
	var job ScanJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return nil, err
	}
	if job.ScanID == "" || job.AssetID == "" || job.UserID == "" || job.Attempt < 0 {
		return nil, fmt.Errorf("scan_id, asset_id, and user_id are required")
	}
	return &job, nil
}
