package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) {
	if err := validateRedisEnv(); err != nil {
		return nil, err
	}

	db, err := parseRedisDB()
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient (&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: os.Getenv("REDIS_PASSWORD"), 
			DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

func validateRedisEnv() error {
	required := map[string]string{
		"REDIS_HOST": os.Getenv("REDIS_HOST"),
		"REDIS_PORT": os.Getenv("REDIS_PORT"),
		"REDIS_PASSWORD": os.Getenv("REDIS_PASSWORD"),
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("environment variable %s is not set", key)
		}
	}
	return nil
}

func parseRedisDB() (int, error) {
	dbStr := os.Getenv("REDIS_DB")
	if dbStr == "" {
		return 0, nil
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return 0, fmt.Errorf("invalid REDIS_DB value %q: must be an integer", dbStr)
	}
	return db, nil
}