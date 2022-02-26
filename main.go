package main

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"time"
)

const NonceStrLen = 16
const Service = "unified.auth.query"
const MchId = "101520021587"
const Key = "8a4199115aa15cd81e064c796a4da1a6"

type Req struct {
	XMLName   xml.Name `xml:"xml"`
	Service   string   `xml:"service"`
	Version   string   `xml:"version,omitempty"`
	Charset   string   `xml:"charset,omitempty"`
	SignType  string   `xml:"sign_type,omitempty"`
	MchId     string   `xml:"mch_id"`
	OutAuthNo string   `xml:"out_auth_no,omitempty"`
	AuthNo    string   `xml:"auth_no,omitempty"`
	NonceStr  string   `xml:"nonce_str"`
	Sign      string   `xml:"sign"`
}

func (r *Req) toXML() ([]byte, error) {
	return xml.MarshalIndent(r, "", "  ")
}

//func (r *Req) sign() ([]byte, error) {
//
//}

func randStr(n int) []byte {
	s := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var b []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		b = append(b, s[r.Intn(len(s))])
	}

	return b
}

func main() {
	req := Req{
		Service:   Service,
		MchId:     MchId,
		OutAuthNo: "",
		AuthNo:    "",
		NonceStr:  string(randStr(NonceStrLen)),
		Sign:      "",
	}

	ret, err := (&req).toXML()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(ret))
}
