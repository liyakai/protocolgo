package utils

import (
	"os"
	"path/filepath"
)

type StUtils struct {
}

// 获取工作根目录
func GetWorkRootPath() string {
	exe, _ := os.Executable() // 获取可执行文件路径
	return filepath.Dir(exe)  // 获取路径中的目录部分
}
