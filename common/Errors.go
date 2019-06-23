package common

import "errors"

var (
	ERR_DISTRIBUTED_ALREADY_LOCK = errors.New("分布式锁已被占用")

	ERR_NO_FOUND_LOCAL_IP = errors.New("没有找到网卡IP")
)
