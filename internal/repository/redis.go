package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"mkk-luna-test-task/internal/task"

	goredis "github.com/redis/go-redis/v9"
)

type redis struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewRedis(client *goredis.Client, ttl time.Duration) *redis {
	return &redis{client: client, ttl: ttl}
}

func taskKey(id int) string {
	return "task:" + strconv.Itoa(id)
}

func taskKeyFromStr(idStr string) string {
	return "task:" + idStr
}

func teamTasksKey(teamId int) string {
	return "team_tasks:" + strconv.Itoa(teamId)
}

func (r *redis) CacheTasks(ctx context.Context, teamId int, tasks []task.Model) error {
	if len(tasks) == 0 {
		return nil
	}

	zs := make([]goredis.Z, 0, len(tasks))

	zKey := teamTasksKey(teamId)

	type kv struct {
		key string
		val []byte
	}

	kvs := make([]kv, 0, len(tasks))

	for _, t := range tasks {
		jsonData, err := json.Marshal(t)

		if err != nil {
			return err
		}

		kvs = append(kvs, kv{key: taskKey(t.Id), val: jsonData})
		zs = append(zs, goredis.Z{Score: float64(t.Id), Member: strconv.Itoa(t.Id)})
	}

	pipe := r.client.Pipeline()

	for _, pair := range kvs {
		pipe.Set(ctx, pair.key, pair.val, r.ttl)
	}

	pipe.ZAdd(ctx, zKey, zs...)
	pipe.Expire(ctx, zKey, r.ttl)

	_, err := pipe.Exec(ctx)

	return err
}

func (r *redis) ListTasksFromCache(ctx context.Context, teamId int, limit int, startFromID *int) (tasks []task.Model, hit bool, err error) {
	if limit <= 0 {
		return nil, false, fmt.Errorf("invalid limit")
	}

	zKey := teamTasksKey(teamId)

	var start int64 = 0

	if startFromID != nil {
		rank, err := r.client.ZRank(ctx, zKey, strconv.Itoa(*startFromID)).Result()

		if err != nil && !errors.Is(err, goredis.Nil) {
			return nil, false, err
		}

		if errors.Is(err, goredis.Nil) {
			return nil, false, nil
		}

		start = rank + 1
	}

	end := start + int64(limit) - 1

	idStrs, err := r.client.ZRange(ctx, zKey, start, end).Result()

	if err != nil {
		return nil, false, err
	}

	if len(idStrs) == 0 {
		return nil, false, nil
	}

	keys := make([]string, 0, len(idStrs))

	for _, id := range idStrs {
		keys = append(keys, taskKeyFromStr(id))
	}

	mgetRes, err := r.client.MGet(ctx, keys...).Result()

	if err != nil {
		return nil, false, err
	}

	results := make([]task.Model, 0, len(mgetRes))

	for _, v := range mgetRes {
		if v == nil {
			return nil, false, nil
		}

		var b []byte

		switch vv := v.(type) {
		case string:
			b = []byte(vv)
		case []byte:
			b = vv
		default:
			return nil, false, nil
		}

		var t task.Model

		if err := json.Unmarshal(b, &t); err != nil {
			return nil, false, nil
		}

		results = append(results, t)
	}

	return results, true, nil
}

func (r *redis) UpdateTaskInCache(ctx context.Context, t task.Model) error {
	jsonData, err := json.Marshal(t)
	if err != nil {
		return err
	}

	zKey := teamTasksKey(t.TeamId)

	pipe := r.client.TxPipeline()

	pipe.Set(ctx, taskKey(t.Id), jsonData, r.ttl)

	pipe.ZAdd(ctx, zKey, goredis.Z{
		Score:  float64(t.Id),
		Member: strconv.Itoa(t.Id),
	})

	pipe.Expire(ctx, zKey, r.ttl)

	_, err = pipe.Exec(ctx)

	return err
}
