package logic

import (
	"bufio"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

// type GenProtoFile struct {
// 	protopath string
// }

func GenProto(filetree *etree.Document, protopath string) {
	if nil == filetree {
		logrus.Error("[GenProtoFile] failed for invalid param: filetree.")
		return
	}
	if protopath == "" || !PathExists(protopath) {
		logrus.Error("[GenProtoFile] failed for invalid param: protopath:", protopath)
		return
	}
	// 遍历各个类型去生成文件
	for _, cataElem := range filetree.ChildElements() {
		switch cataElem.Tag {
		case "enum":
			GenEnumProto(cataElem, protopath)
		case "data":
			GenDataProto(cataElem, protopath)
		case "protocol":
			GenProtocolProto(cataElem, protopath)
		case "rpc":
			GenRpcProto(cataElem, protopath)
		}
	}

}

func GenEnumProto(catatree *etree.Element, protopath string) {
	if nil == catatree {
		logrus.Error("[GenEnumProto] failed for invalid param: catatree.")
		return
	}
	strFilePath := protopath + "/" + "enum.proto"

	// 尝试以只写模式打开文件，如果文件不存在，则创建文件,如果文件存在则清空文件.
	fileHandler, err := os.OpenFile(strFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logrus.Error("Failed to open or create file:", err, ", filename:", strFilePath)
		return
	}
	defer fileHandler.Close()

	if !GenProtoHead(fileHandler, catatree.Tag) {
		logrus.Error("GenProtoHead failed. filename:", strFilePath)
		return
	}
	if !GenProtoBody(fileHandler, catatree) {
		logrus.Error("GenProtoBody failed. filename:", strFilePath)
		return
	}

	fileHandler.Sync()

	logrus.Info("[GenEnumProto] GenEnumProto done. strFilePath:", strFilePath)
}

func GenDataProto(catatree *etree.Element, protopath string) {
	if nil == catatree {
		logrus.Error("[GenDataProto] failed for invalid param: catatree.")
		return
	}
	logrus.Info("[GenDataProto] GenDataProto done.")
}

func GenProtocolProto(catatree *etree.Element, protopath string) {
	if nil == catatree {
		logrus.Error("[GenProtocolProto] failed for invalid param: catatree.")
		return
	}
	logrus.Info("[GenProtocolProto] GenProtocolProto done.")
}

func GenRpcProto(catatree *etree.Element, protopath string) {
	if nil == catatree {
		logrus.Error("[GenRpcProto] failed for invalid param: catatree.")
		return
	}
	logrus.Info("[GenRpcProto] GenRpcProto done.")
}

// 生成 proto 文件的 head
func GenProtoHead(fileHandler *os.File, packageName string) bool {
	if nil == fileHandler {
		logrus.Error("[GenProtoHead] Failed to GenProtoHead for invalid param: fileHandler.")
		return false
	}

	// 写入 Proto3 文件头
	_, err := fileHandler.WriteString(`syntax = "proto3";

package ` + packageName + `;

`)
	if err != nil {
		logrus.Error("[GenProtoHead] Failed toWriteString:", err)
		return false
	}

	return true
}

func GenProtoBody(fileHandler *os.File, catatree *etree.Element) bool {
	if nil == fileHandler {
		logrus.Error("[GenProtoHead] Failed to GenProtoHead for invalid param: fileHandler.")
		return false
	}
	if nil == catatree {
		logrus.Error("[GenProtoHead] Failed to GenProtoHead for invalid param: catatree.")
		return false
	}

	for _, elem := range catatree.ChildElements() {
		if !GenStruct(fileHandler, catatree.Tag, elem) {
			logrus.Error("[GenProtoHead] GenStruct failed. catatree.Tag:", catatree.Tag)
			return false
		}
	}
	return true
}

func GenStruct(fileHandler *os.File, structType string, structTree *etree.Element) bool {
	if nil == fileHandler {
		logrus.Error("[GenStruct] Failed to GenStruct for invalid param: fileHandler.")
		return false
	}
	if nil == structTree {
		logrus.Error("[GenStruct] Failed to GenStruct for invalid param: structTree.")
		return false
	}
	// 处理结构注释
	for _, child := range structTree.Child {
		// 检查该子元素是否为注释
		if comment, ok := child.(*etree.Comment); ok {
			if comment.Data != "" {
				scanner := bufio.NewScanner(strings.NewReader(comment.Data))

				for scanner.Scan() {
					line := scanner.Text()
					_, err := fileHandler.WriteString("// " + line + "\n")
					if err != nil {
						logrus.Error("[GenStruct] Failed toWriteString:", err)
						return false
					}
				}
			}
			break
		}
	}

	strStructHeadType := "message"
	if structType == "enum" {
		strStructHeadType = "enum"
	}

	_, err := fileHandler.WriteString(strStructHeadType + " " + structTree.Tag + " { \n")
	if err != nil {
		logrus.Error("[GenStruct] Failed toWriteString:", err)
		return false
	}

	for _, elem := range structTree.ChildElements() {
		etreeElemName := elem.SelectAttr("EntryName")
		etreeElemIndex := elem.SelectAttr("EntryIndex")
		etreeElemComment := elem.SelectAttr("EntryComment")
		if etreeElemName == nil || etreeElemIndex == nil || etreeElemComment == nil {
			logrus.Error("[GenStruct] Failed to get entry attr, invalid format. elem:", elem)
			return false
		}
		// 元素数据
		_, err := fileHandler.WriteString("    " + etreeElemName.Value + " = " + etreeElemIndex.Value + ";")
		if err != nil {
			logrus.Error("[GenStruct] Failed toWriteString:", err)
			return false
		}
		// 注释
		strComment := "\n"
		if etreeElemComment.Value != "" {
			strComment = "    //" + etreeElemComment.Value + " \n"
		}
		_, err = fileHandler.WriteString(strComment)
		if err != nil {
			logrus.Error("[GenStruct] Failed toWriteString:", err)
			return false
		}
	}

	_, err = fileHandler.WriteString("} \n\n")
	if err != nil {
		logrus.Error("[GenStruct] Failed toWriteString:", err)
		return false
	}

	return true
}
