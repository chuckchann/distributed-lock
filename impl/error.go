package impl

import "errors"

var (
	ErrLockFailed  = errors.New("get lock failed")
	ErrLockTimeout = errors.New("get lock timeout")
	ErrUnLock      = errors.New("please Lock before you call UnLock")
)
