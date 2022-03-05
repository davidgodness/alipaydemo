package main

import (
	"encoding/xml"
)

// Req 通用请求
type Req struct {
	XMLName  xml.Name `xml:"xml"`
	Service  string   `xml:"service" json:"service,omitempty"`
	Version  string   `xml:"version,omitempty"`
	Charset  string   `xml:"charset,omitempty"`
	SignType string   `xml:"sign_type,omitempty"`
	MchId    string   `xml:"mch_id"`
	NonceStr string   `xml:"nonce_str"`
	Sign     string   `xml:"sign"`
}

// QueryReq 查询订单请求
type QueryReq struct {
	Req
	OutAuthNo string `xml:"out_auth_no,omitempty"`
	AuthNo    string `xml:"auth_no,omitempty"`
}

// UnFreezeReq 解冻请求
type UnFreezeReq struct {
	Req
	OutRequestNo string
	AuthNo       string
	TotalFee     int
	Remark       string
	StoreId      string
	TerminalId   string
	NotifyUrl    string
}
