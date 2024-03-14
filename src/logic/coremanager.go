package logic

import (
	"protocolgo/src/gui"

	"github.com/beevik/etree"
	"github.com/rs/zerolog"
)

type CoreManager struct {
	Stapp *gui.StApp

	Utils  *StUtils       // 工具
	Logger zerolog.Logger // 自定义Logger

	DocEtree *etree.Document
}

func (Stapp *CoreManager) CreateNewXml() {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.DocEtree {
		Stapp.SaveToFile()
	}
	// 创建新的xml
	DocEtree := etree.NewDocument()
	DocEtree.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	DocEtree.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)
}

func (Stapp *CoreManager) ReadXmlFromFile() {
}

func (Stapp *CoreManager) SaveToFile() {
	if nil != Stapp.DocEtree {
		Stapp.SaveToFile()
	}
}
