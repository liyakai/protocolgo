package logic

import (
	"bufio"
	"os"
	"os/exec"
	"protocolgo/src/utils"
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
		strProtoFilePath := protopath + "/" + cataElem.Tag + ".proto"
		if !GenStructProto(cataElem, strProtoFilePath) {
			logrus.Error("[GenProtoFile] failed for GenStructProto. strProtoFilePath:", strProtoFilePath)
			return
		}
	}

}

func GenStructProto(catatree *etree.Element, protopath string) bool {
	if nil == catatree {
		logrus.Error("[GenEnumProto] failed for invalid param: catatree.")
		return false
	}

	// 尝试以只写模式打开文件，如果文件不存在，则创建文件,如果文件存在则清空文件.
	fileHandler, err := os.OpenFile(protopath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logrus.Error("[GenStructProto] Failed to open or create file:", err, ", filename:", protopath)
		return false
	}
	defer fileHandler.Close()

	if !GenProtoHead(fileHandler, catatree.Tag) {
		logrus.Error("[GenStructProto] GenProtoHead failed. filename:", protopath)
		return false
	}
	if !GenProtoBody(fileHandler, catatree) {
		logrus.Error("[GenStructProto] GenProtoBody failed. filename:", protopath)
		return false
	}

	fileHandler.Sync()

	logrus.Info("[GenStructProto] GenStructProto done. filename:", protopath)
	return true
}

// 生成 proto 文件的 head
func GenProtoHead(fileHandler *os.File, packageName string) bool {
	if nil == fileHandler {
		logrus.Error("[GenProtoHead] Failed to GenProtoHead for invalid param: fileHandler.")
		return false
	}

	strImport := ""
	if packageName == "data" {
		strImport = `
import "enum.proto";


`
	} else if packageName == "protocol" || packageName == "rpc" {
		strImport = `
import "enum.proto";
import "data.proto";


`
	}

	// 写入 Proto3 文件头
	_, err := fileHandler.WriteString(`syntax = "proto3";

// package ` + packageName + `;
option go_package = "example/` + packageName + `";

` + strImport)
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

	if structType == "enum" {
		if !GenEnumStruct(fileHandler, structType, structTree) {
			logrus.Error("[GenStruct] Failed to GenEnumStruct:")
			return false
		}
	} else if structType == "rpc" {
		GenRpcStruct(fileHandler, structTree)
	} else {
		if !GenMessageStruct(fileHandler, structType, structTree) {
			logrus.Error("[GenStruct] Failed to GenMessageStruct:")
			return false
		}
	}

	return true
}

func GenEnumStruct(fileHandler *os.File, structType string, structTree *etree.Element) bool {
	if nil == fileHandler {
		logrus.Error("[GenEnumStruct] Failed to GenStruct for invalid param: fileHandler.")
		return false
	}
	if nil == structTree {
		logrus.Error("[GenEnumStruct] Failed to GenStruct for invalid param: structTree.")
		return false
	}
	strStructHeadType := "message"
	if structType == "enum" {
		strStructHeadType = "enum"
	}

	_, err := fileHandler.WriteString(strStructHeadType + " " + structTree.Tag + " { \n")
	if err != nil {
		logrus.Error("[GenEnumStruct] Failed toWriteString:", err)
		return false
	}

	for _, elem := range structTree.ChildElements() {
		etreeElemName := elem.SelectAttr("EntryName")
		etreeElemIndex := elem.SelectAttr("EntryIndex")
		etreeElemComment := elem.SelectAttr("EntryComment")
		if etreeElemName == nil || etreeElemIndex == nil || etreeElemComment == nil {
			logrus.Error("[GenEnumStruct] Failed to get entry attr, invalid format. elem:", elem)
			return false
		}
		// 元素数据
		_, err := fileHandler.WriteString("	" + etreeElemName.Value + "		=	" + etreeElemIndex.Value + ";")
		if err != nil {
			logrus.Error("[GenEnumStruct] Failed toWriteString:", err)
			return false
		}
		// 注释
		strComment := "\n"
		if etreeElemComment.Value != "" {
			strComment = "	//" + etreeElemComment.Value + "	\n"
		}
		_, err = fileHandler.WriteString(strComment)
		if err != nil {
			logrus.Error("[GenEnumStruct] Failed toWriteString:", err)
			return false
		}
	}

	_, err = fileHandler.WriteString("} \n\n")
	if err != nil {
		logrus.Error("[GenEnumStruct] Failed toWriteString:", err)
		return false
	}

	return true
}

func GenRpcStruct(fileHandler *os.File, structTree *etree.Element) bool {
	if nil == fileHandler {
		logrus.Error("[GenRpcStruct] Failed to GenStruct for invalid param: fileHandler.")
		return false
	}
	if nil == structTree {
		logrus.Error("[GenRpcStruct] Failed to GenStruct for invalid param: structTree.")
		return false
	}
	for _, elemRPC := range structTree.ChildElements() {
		strStructHeadType := "message"
		etreeRpcType := elemRPC.SelectAttr("RpcType")
		if etreeRpcType == nil {
			logrus.Error("[GenRpcStruct] Failed to get RpcType, invalid format. elem:", elemRPC)
			return false
		}
		_, err := fileHandler.WriteString(strStructHeadType + " " + elemRPC.Tag + etreeRpcType.Value + " { \n")
		if err != nil {
			logrus.Error("[GenRpcStruct] Failed toWriteString:", err)
			return false
		}

		for _, elem := range elemRPC.ChildElements() {
			etreeElemOption := elem.SelectAttr("EntryOption")
			etreeElemType := elem.SelectAttr("EntryType")
			etreeElemName := elem.SelectAttr("EntryName")
			etreeElemIndex := elem.SelectAttr("EntryIndex")
			etreeElemComment := elem.SelectAttr("EntryComment")
			if etreeElemOption == nil || etreeElemType == nil || etreeElemName == nil || etreeElemIndex == nil || etreeElemComment == nil {
				logrus.Error("[GenRpcStruct] Failed to get entry attr, invalid format. elem:", elem)
				return false
			}
			// 元素数据
			_, err := fileHandler.WriteString("	" + etreeElemOption.Value + "	" + etreeElemType.Value + "			" + etreeElemName.Value + "	=	" + etreeElemIndex.Value + ";")
			if err != nil {
				logrus.Error("[GenRpcStruct] Failed toWriteString:", err)
				return false
			}
			// 注释
			strComment := "\n"
			if etreeElemComment.Value != "" {
				strComment = "	//" + etreeElemComment.Value + "	\n"
			}
			_, err = fileHandler.WriteString(strComment)
			if err != nil {
				logrus.Error("[GenRpcStruct] Failed toWriteString:", err)
				return false
			}
		}

		_, err = fileHandler.WriteString("} \n\n")
		if err != nil {
			logrus.Error("[GenRpcStruct] Failed toWriteString:", err)
			return false
		}
	}

	return true
}

func GenMessageStruct(fileHandler *os.File, structType string, structTree *etree.Element) bool {
	if nil == fileHandler {
		logrus.Error("[GenMessageStruct] Failed to GenStruct for invalid param: fileHandler.")
		return false
	}
	if nil == structTree {
		logrus.Error("[GenMessageStruct] Failed to GenStruct for invalid param: structTree.")
		return false
	}

	strStructHeadType := "message"
	if structType == "enum" {
		strStructHeadType = "enum"
	}

	_, err := fileHandler.WriteString(strStructHeadType + " " + structTree.Tag + " { \n")
	if err != nil {
		logrus.Error("[GenMessageStruct] Failed toWriteString:", err)
		return false
	}

	for _, elem := range structTree.ChildElements() {
		etreeElemOption := elem.SelectAttr("EntryOption")
		etreeElemType := elem.SelectAttr("EntryType")
		etreeElemName := elem.SelectAttr("EntryName")
		etreeElemIndex := elem.SelectAttr("EntryIndex")
		etreeElemComment := elem.SelectAttr("EntryComment")
		if etreeElemOption == nil || etreeElemType == nil || etreeElemName == nil || etreeElemIndex == nil || etreeElemComment == nil {
			logrus.Error("[GenMessageStruct] Failed to get entry attr, invalid format. elem:", elem)
			return false
		}
		// 元素数据
		_, err := fileHandler.WriteString("	" + etreeElemOption.Value + "	" + etreeElemType.Value + "			" + etreeElemName.Value + "	=	" + etreeElemIndex.Value + ";")
		if err != nil {
			logrus.Error("[GenMessageStruct] Failed toWriteString:", err)
			return false
		}
		// 注释
		strComment := "\n"
		if etreeElemComment.Value != "" {
			strComment = "	//" + etreeElemComment.Value + "	\n"
		}
		_, err = fileHandler.WriteString(strComment)
		if err != nil {
			logrus.Error("[GenMessageStruct] Failed toWriteString:", err)
			return false
		}
	}

	_, err = fileHandler.WriteString("} \n\n")
	if err != nil {
		logrus.Error("[GenMessageStruct] Failed toWriteString:", err)
		return false
	}

	return true
}

func GenPbFromProto(protopath string, outputPath string) {
	if protopath == "" || !PathExists(protopath) {
		logrus.Error("[GenPbFromProto] failed for invalid param: protopath:", protopath)
		return
	}
	if outputPath == "" || !PathExists(outputPath) {
		logrus.Error("[GenPbFromProto] failed for invalid param: outputPath:", outputPath)
		return
	}
	logrus.Debug("[GenPbFromProto] param:protopath:", protopath, ",outputPath:", outputPath)
	// 定义要执行的 protoc 命令，包括所有需要的参数
	// 这里以生成 Go 相关代码为例，确保您已定义好 .proto 文件
	command := utils.GetWorkRootPath() + "/data/protoc"
	// args := []string{"./output_protofiles/*.proto"}
	args := []string{"--proto_path=" + protopath, "--go_out=" + outputPath, "--go_opt=paths=source_relative", utils.GetWorkRootPath() + "/data/output_protofiles/*.proto"}

	// ./protoc --proto_path=./output_protofiles --go_out=./output_pbfiles --go_opt=paths=source_relative ./output_protofiles/*.proto
	// 使用 exec.Command 创建命令
	cmd := exec.Command(command, args...)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Error("Error executing protoc command:", err, ",output:", string(output), ",command:", command)
		return
	}

	// 打印命令输出
	logrus.Info("protoc command output:", string(output))

}
