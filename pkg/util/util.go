package util

import (
	"log"
)

// Cleanup 执行一个清理函数，并在发生错误时记录日志。
// taskName 用于在日志中标识是哪个清理任务失败了。
func Cleanup(taskName string, f func() error) {
	if err := f(); err != nil {
		log.Printf("ERROR: cleanup task '%s' failed: %v", taskName, err)
	}
}
