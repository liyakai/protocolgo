package logic

import "encoding/xml"

type ProtoStore struct {
	XMLName     xml.Name    `xml:"protostore"`
	ProtoUnites []ProtoUnit `xml:"Proto"`
}

type ProtoUnit struct {
	XMLName xml.Name `xml:"protounit"`
	Protoes []Proto  `xml:"proto"`
}

type Proto struct {
	XMLName xml.Name `xml:"proto"`
}
