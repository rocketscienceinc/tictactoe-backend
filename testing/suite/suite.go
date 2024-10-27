package suite

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
)

const (
	expireDuration  = 120
	maxWaitDuration = 120 * time.Second
)

const (
	redisPort  = "6379/tcp"
	redisImage = "redis"
	redisTag   = "alpine"
)

type Suite struct {
	*testing.T
	Logger *slog.Logger

	Storage *redis.Client
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), maxWaitDuration)
	t.Cleanup(func() {
		cancel()
	})

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %v", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: redisImage,
		Tag:        redisTag,
		Env:        []string{},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("could not start resource: %v", err)
	}

	// never returns error
	_ = resource.Expire(expireDuration) // Tell docker to hard kill the container in 120 seconds

	redisHost := resource.GetHostPort(redisPort)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = maxWaitDuration

	var redisClient *redis.Client
	if err = pool.Retry(func() error {
		redisClient = redis.NewClient(&redis.Options{
			Addr: redisHost,
		})
		return redisClient.Ping(ctx).Err()
	}); err != nil {
		if err = pool.Purge(resource); err != nil {
			t.Fatalf("could not purge resource: %v", err)
		}

		t.Fatalf("could not connect to redis: %v", err)
	}

	if err = redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("could not flush database: %v", err)
	}

	t.Cleanup(func() {
		t.Helper()

		if err = pool.Purge(resource); err != nil {
			t.Fatalf("could not purge resource: %v", err)
		}
	})

	return ctx, &Suite{
		T:       t,
		Logger:  logger,
		Storage: redisClient,
	}
}
