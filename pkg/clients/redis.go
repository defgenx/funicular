package clients

import (
	"github.com/defgenx/funicular/internal/utils"
	"github.com/go-redis/redis"
	"log"
	"net"
	"strconv"
	"time"
)

type RedisConfig struct {
	Host string
	Port uint16
	DB uint8
}

func (rc *RedisConfig) ToOption() *redis.Options {
	return &redis.Options{
		Addr: net.JoinHostPort(rc.Host, strconv.Itoa(int(rc.Port))),
		DB: int(rc.DB),
	}
}

//------------------------------------------------------------------------------

type RedisManager struct {
	Clients map[string][]*RedisWrapper
}

func NewRedisManager() *RedisManager {
	return &RedisManager{
		Clients: make(map[string][]*RedisWrapper),
	}
}

func (rw *RedisManager) AddClient(config RedisConfig, category string, channel string) (*RedisWrapper, error) {
	if category == "" {
		return nil, utils.ErrorPrint("category must be filled")
	}
	if channel == "" {
		channel = category
	}
	client := NewRedisWrapper(config, channel)
	rw.add(client, category)
	return client, nil
}

func (rw *RedisManager) Close() error {
	var manageClientsCopy map[string][]*RedisWrapper
	manageClientsCopy = copyRedisClients(rw.Clients)
	if len(manageClientsCopy) > 0 {
		for category, clients := range manageClientsCopy {
			for _, client := range clients {
				err := client.Client.Close()
				if err != nil {
					return utils.ErrorPrintf("an error occurred while closing client connexion pool: %v", err)
				}
				rw.Clients[category] = rw.Clients[category][1:]
			}
			delete(rw.Clients, category)
		}
	} else {
		log.Print("Manager have no clients to close...")
	}
	return nil
}

func (rw *RedisManager) add(redisWrapper *RedisWrapper, category string) {
	mm, ok := rw.Clients[category]
	if !ok {
		mm = make([]*RedisWrapper, 0)
	}
	mm = append(mm, redisWrapper)
	rw.Clients[category] = mm
}

//------------------------------------------------------------------------------

type RedisWrapper struct {
	Client  *redis.Client
	config  *RedisConfig
	channel string
}

func NewRedisWrapper(config RedisConfig, channel string) *RedisWrapper {
	client := redis.NewClient(config.ToOption())
	return &RedisWrapper{
		Client: client,
		config: &config,
		channel: channel,
	}
}

func (w *RedisWrapper) SendMessage(data map[string]interface{}) (string, error) {
	xAddArgs := &redis.XAddArgs{
		Stream: w.channel,
		Values: data,
	}
	result := w.Client.XAdd(xAddArgs)
	return result.Result()
}

func (w *RedisWrapper) ReadMessage(last_id string, count int64, block time.Duration) ([]redis.XStream, error) {
	var channels = make([]string, 0)
	channels = append(channels, w.channel)
	channels = append(channels, last_id)
	xReadArgs := &redis.XReadArgs{
		Streams: channels,
		Count: count,
		Block: block,
	}
	result := w.Client.XRead(xReadArgs)
	return result.Result()
}

func (w *RedisWrapper) GetChannel() string {
	return w.channel
}

func (w *RedisWrapper) ReadRangeMessage(start string, stop string) ([]redis.XMessage, error) {
	result := w.Client.XRange(w.channel, start, stop)
	return result.Result()
}

func (w *RedisWrapper) DeleteMessage(ids ...string) (int64, error) {
	result := w.Client.XDel(w.channel, ids...)
	return result.Result()
}

func (w *RedisWrapper) CreateGroup(group string, start string) (string, error) {
	result := w.Client.XGroupCreate(w.channel, group, start)
	return result.Result()
}

func (w *RedisWrapper) DeleteGroup(group string) (int64, error) {
	result := w.Client.XGroupDestroy(w.channel, group)
	return result.Result()
}

func (w *RedisWrapper) PendingMessage(group string) (*redis.XPending, error) {
	result := w.Client.XPending(w.channel, group)
	return result.Result()
}

func (w *RedisWrapper) AckMessage(group string, ids ...string) (int64, error) {
	result := w.Client.XAck(w.channel, group, ids...)
	return result.Result()
}

//------------------------------------------------------------------------------
// MISC
//------------------------------------------------------------------------------

func copyRedisClients(originalMap map[string][]*RedisWrapper) map[string][]*RedisWrapper {
	var newMap = make(map[string][]*RedisWrapper)
	for key, values := range originalMap {
		newMap[key] = values
	}
	return newMap
}