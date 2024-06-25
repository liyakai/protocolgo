package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"unicode"
)

type StUtils struct {
}

// 获取工作根目录
func GetWorkRootPath() string {
	exe, _ := os.Executable() // 获取可执行文件路径
	return filepath.Dir(exe)  // 获取路径中的目录部分
}

// 检查字符串是否是正整数
func CheckPositiveInteger(str string) bool {
	num, err := strconv.Atoi(str)

	if err != nil {
		return false
	}

	return num > 0
}

// 检查字符串是否是自然数
func CheckNaturalInteger(str string) bool {
	num, err := strconv.Atoi(str)

	if err != nil {
		return false
	}

	return num >= 0
}

// 检查字符串是否以数字开头
func CheckStartWithNum(str string) bool {
	if len(str) > 0 {
		r := rune(str[0])
		if unicode.IsDigit(r) {
			return true
		} else {
			return false
		}
	}
	return false
}

// 验证IP地址是否合法的函数
func isValidIP(ip string) bool {
	// IPv4地址的正则表达式
	ipv4Pattern := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`

	// 匹配IP地址
	match, err := regexp.MatchString(ipv4Pattern, ip)
	if err != nil {
		return false // 正则表达式错误，视为非法
	}

	return match
}

// 验证端口是否合法的函数
func isValidPort(port string) bool {
	// 端口号的正则表达式，包括1-5位数字，范围在1-65535之间
	portPattern := `^([1-9]|[1-9]\d{1,4}|[1-5]\d{4}|6([0-4]\d{3}|5([0-4]\d{2}|5([0-2]\d|3[0-5]))))$`

	// 匹配端口号
	match, err := regexp.MatchString(portPattern, port)
	if err != nil {
		return false // 正则表达式错误，视为非法
	}

	return match
}
