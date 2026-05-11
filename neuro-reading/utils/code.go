package utils

import (
	"context"
	"fmt"
	"neuro-reading/db"
	"time"
)

const codePrefix = "verify_code:"
const codeExpiry = 5 * time.Minute
const codeCooldown = time.Minute

func StoreVerifyCode(account, code, codeType string) error {
	ctx := context.Background()
	key := codePrefix + account

	exists, err := db.Redis.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		ttl, err := db.Redis.TTL(ctx, key).Result()
		if err != nil {
			return err
		}
		if ttl > codeExpiry-codeCooldown {
			return fmt.Errorf("发送过于频繁")
		}
	}

	data := map[string]interface{}{
		"code": code,
		"type": codeType,
	}

	pipe := db.Redis.Pipeline()
	pipe.HSet(ctx, key, data)
	pipe.Expire(ctx, key, codeExpiry)
	_, err = pipe.Exec(ctx)
	return err
}

func GetVerifyCode(account string) (code, codeType string, err error) {
	ctx := context.Background()
	key := codePrefix + account

	result, err := db.Redis.HGetAll(ctx, key).Result()
	if err != nil {
		return "", "", err
	}
	if len(result) == 0 {
		return "", "", fmt.Errorf("验证码不存在或已过期")
	}

	return result["code"], result["type"], nil
}

func DeleteVerifyCode(account string) error {
	ctx := context.Background()
	return db.Redis.Del(ctx, codePrefix+account).Err()
}
