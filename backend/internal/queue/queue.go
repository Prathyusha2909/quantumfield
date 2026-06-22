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
	payload, err := json.Marshal(job)
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
	var job ScanJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("decode scan job: %w", err)
	}
	return &job, nil
}

func (client *Client) Close() error {
	return client.redis.Close()
}
