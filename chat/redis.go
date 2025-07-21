// redis.go
package chat

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClientWrapper struct {
	Client *redis.Client
}

func NewRedisClient() *RedisClientWrapper {
	return &RedisClientWrapper{
		Client: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			DB:   3,
		}),
	}
}

func SubscribeAndDispatch(rdb *RedisClientWrapper, hub *Hub) {
	sub := rdb.Client.Subscribe(ctx, hub.ID)
	ch := sub.Channel()

	go func() {
		for msg := range ch {
			var payload RPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				return
			}

			hub.Mutex.RLock()
			client, ok := hub.Clients[payload.To]
			hub.Mutex.RUnlock()

			if ok {
				client.SendCh <- []byte(payload.Message)
			}
		}
	}()
}
