package common

import "errors"

var (
	ERR_DISTRIBUTED_ALREADY_LOCK = errors.New("分布式锁已被占用")
)
