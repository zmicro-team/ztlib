package verified_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redisV9 "github.com/zmicro-team/ztlib/limiter/verified/redis/v9"
	"github.com/zmicro-team/ztlib/limiter/verified/tests"
)

func TestCaptcha_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	tests.GenericTestCaptcha_Improve_Cover(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestCaptcha_Unsupported_Driver(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestCaptcha_Unsupported_Driver(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestCaptcha_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestCaptcha_RedisUnavailable(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestCaptcha_OneTime(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestCaptcha_OneTime(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestCaptcha_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestCaptcha_In_Quota(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestCaptcha_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestCaptcha_Over_Quota(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// // TODO: success in redis, but failed in miniredis
// func TestCaptcha_RedisV9_Onetime_Timeout(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	assert.NoError(t, err)

// 	defer mr.Close()

// 	tests.GenericTestCaptcha_Onetime_Timeout(
// 		t,
// 		// redisV8.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 0})),
// 	)
// }
