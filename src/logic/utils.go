package logic

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type StUtils struct {
	CoreMgr *CoreManager
}

// 打开对话框
func (stutils *StUtils) ShowDialog(title string, message string) {
	customDialog := dialog.NewCustom("title", "OK", widget.NewLabel(message), *stutils.CoreMgr.Stapp.Window)
	customDialog.Show()
}

// 获取工作根目录
func (stutils *StUtils) GetWorkRootPath() string {
	exe, _ := os.Executable() // 获取可执行文件路径
	return filepath.Dir(exe)  // 获取路径中的目录部分
}
