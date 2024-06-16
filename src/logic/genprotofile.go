package logic

import (
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
	if protopath == "" || PathExists(protopath) {
		logrus.Error("[GenProtoFile] failed for invalid param: protopath:", protopath)
		return
	}

	// for _, cataElem := range filetree.ChildElements() {

	// }

}
