package v8

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/zmicro-team/ztlib/limiter/limit_verified/tests"
)

func Test_RedisV8_Name(t *testing.T) {
	tests.GenericTestName(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})),
	)
}

func Test_RedisV8_SendCode_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestSendCode_RedisUnavailable(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func Test_RedisV8_SendCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Success(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_Err_Failure(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Err_Failure(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_MaxSendPerDay(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_Concurrency_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Concurrency_MaxSendPerDay(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_ResendTooFrequently(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_SendCode_Concurrency_ResendTooFrequently(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestSendCode_Concurrency_ResendTooFrequently(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_VerifyCode_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_Success(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_VerifyCode_CodeRequired(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_CodeRequired(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: mini redis 测试失败, 但redis是成功的
// func Test_RedisV8_VerifyCode_CodeExpired(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	require.Nil(t, err)
// 	defer mr.Close()

// 	tests.GenericTestVerifyCode_CodeExpired(
// 		t,
// 		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 	)
// }

func Test_RedisV8_VerifyCode_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_CodeMaxError(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8_VerifyCode_Concurrency_CodeMaxError(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTestVerifyCode_Concurrency_CodeMaxError(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8__INCR_MaxSendPerDay(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_INCR_MaxSendPerDay(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func Test_RedisV8__INCR_DECR(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()

	tests.GenericTest_INCR_DECR(
		t,
		NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}
