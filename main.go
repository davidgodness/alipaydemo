package main

import (
	"crypto/md5"
	"encoding/xml"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

const NonceStrLen = 16
const Service = "unified.auth.query"
const MchId = "101520021587"
const Key = "8a4199115aa15cd81e064c796a4da1a6"

//const GatewayUrl = "https://pay.swiftpass.cn/pay/gateway"

type Req struct {
	XMLName   xml.Name `xml:"xml"`
	Service   string   `xml:"service" json:"service,omitempty"`
	Version   string   `xml:"version,omitempty"`
	Charset   string   `xml:"charset,omitempty"`
	SignType  string   `xml:"sign_type,omitempty"`
	MchId     string   `xml:"mch_id"`
	OutAuthNo string   `xml:"out_auth_no,omitempty"`
	AuthNo    string   `xml:"auth_no,omitempty"`
	NonceStr  string   `xml:"nonce_str"`
	Sign      string   `xml:"sign"`
}

func (r Req) toXML() ([]byte, error) {
	return xml.MarshalIndent(r, "", "  ")
}

func (r Req) queryString() []byte {
	v := reflect.ValueOf(r)
	n := v.NumField()
	params := make(map[string]interface{}, n-1)
	var fields sort.StringSlice
	for i := 0; i < n; i++ {
		name := v.Type().Field(i).Name
		tag := v.Type().Field(i).Tag
		if name == "XMLName" {
			continue
		}
		key := tag.Get("xml")
		if strings.Contains(key, "omitempty") && v.Field(i).IsZero() {
			continue
		}
		key = strings.TrimSuffix(key, ",omitempty")
		value := v.Field(i).String()
		if name == "Sign" {
			key = "key"
			value = Key
		}
		params[key] = value
		fields = append(fields, key)
	}

	/*
		构造查询字符串
	*/
	fields.Sort()
	str := ""
	for i := 0; i < fields.Len(); i++ {
		str += fmt.Sprintf("%s=%s&", fields[i], params[fields[i]])
	}

	return []byte(strings.TrimSuffix(str, "&"))
}

func (r Req) sign() string {
	sum := md5.Sum(r.queryString())
	return strings.ToUpper(fmt.Sprintf("%x", sum))
}

func randStr(n int) []byte {
	s := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var b []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		b = append(b, s[r.Intn(len(s))])
	}

	return b
}

func hint() {
	fmt.Println("please input subcommand")
	fmt.Println("query 查询订单")
	os.Exit(-1)
}

func main() {
	if len(os.Args) < 2 {
		hint()
	}
	var err error
	req := Req{
		Service:  Service,
		MchId:    MchId,
		NonceStr: string(randStr(NonceStrLen)),
	}
	req.Sign = req.sign()

	query := flag.NewFlagSet("query", flag.ExitOnError)
	outAuthNo := query.String("out_auth_no", "", "第三方商户号")
	authNo := query.String("auth_no", "", "商户号")

	switch os.Args[1] {
	case "query":
		err = query.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		if query.NFlag() == 0 {
			fmt.Println("Usage of query:")
			query.PrintDefaults()
			os.Exit(-1)
		}
		req.OutAuthNo = *outAuthNo
		req.AuthNo = *authNo
		break
	default:
		hint()
		break
	}

	xmlStr, _ := req.toXML()
	fmt.Println(string(xmlStr))
}
