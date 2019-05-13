package redis

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

//Client redis client
type Client struct {
	conn redis.Conn
}

//GetDB get redis client
func GetDB(redisserver string, password string, db int) (c *Client, err error) {
	options := make([]redis.DialOption, 0)
	if password != "" {
		options = append(options, redis.DialPassword(password))
	}
	options = append(options, redis.DialDatabase(db))

	redisDB, err := redis.Dial("tcp", redisserver, options...)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return nil, err
	}
	return &Client{conn: redisDB}, err
}

//SetValue save
func (c *Client) SetValue(key string, data []byte) error {
	if _, err := c.conn.Do("SET", key, data); err != nil {
		return err
	}
	return nil
}

//GetValue get value from dis
func (c *Client) GetValue(key string) ([]byte, error) {
	return redis.Bytes(c.conn.Do("GET", key))
}

//GetAllKeys 循环获取所有的key
func (c *Client) GetAllKeys() ([]string, error) {
	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.MultiBulk(c.conn.Do("SCAN", iter))
		if err != nil {
			return nil, err
		}
		iter, _ = redis.Int(arr[0], nil)
		keyslice, _ := redis.Strings(arr[1], nil)
		keys = append(keys, keyslice...)
		if iter == 0 {
			break
		}
	}
	return keys, nil
}
