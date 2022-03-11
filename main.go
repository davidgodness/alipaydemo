package main

import (
	"crypto/md5"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	NonceStrLen = 16
	MchId       = "101520021587"
	Key         = "8a4199115aa15cd81e064c796a4da1a6"
	GatewayUrl  = "https://pay.swiftpass.cn/pay/gateway"
	Query       = "unified.auth.query"
	UnFreeze    = "unified.auth.unfreeze"
)

// 生成指定长度的随机字符串
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
	fmt.Println("unfreeze 解冻订单")
	os.Exit(-1)
}

func handleSubCommand(c string, flags map[string]string, args []string) (map[string]*string, error) {
	s := flag.NewFlagSet(c, flag.ExitOnError)
	v := make(map[string]*string, len(flags))
	for name, comment := range flags {
		v[name] = s.String(name, "", comment)
	}
	err := s.Parse(args)
	if err != nil {
		return nil, err
	}
	if s.NFlag() == 0 {
		fmt.Printf("Usage of %s:\n", c)
		s.PrintDefaults()
		return nil, err
	}
	return v, nil
}

type Req map[string]string

// 构造出用来计算签名的查询字符串
func (r Req) queryString() []byte {
	var fields sort.StringSlice
	for s := range r {
		fields = append(fields, s)
	}
	fields.Sort()
	str := ""
	for i := 0; i < fields.Len(); i++ {
		if len(r[fields[i]]) > 0 {
			str += fmt.Sprintf("%s=%s&", fields[i], r[fields[i]])
		}
	}

	return []byte(strings.TrimSuffix(str, "&") + "&key=" + Key)
}

func (r Req) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Local: "xml",
	}
	tokens := []xml.Token{start}
	for k, v := range r {
		t := xml.StartElement{Name: xml.Name{Local: k}}
		tokens = append(tokens, t, xml.Directive(fmt.Sprintf("[CDATA[%s]]", v)), xml.EndElement{Name: t.Name})
	}
	tokens = append(tokens, xml.EndElement{Name: start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	err := e.Flush()
	if err != nil {
		return err
	}

	return nil
}

// 计算签名，需要查询字符串
func (r Req) sign() string {
	fmt.Println(string(r.queryString()))
	sum := md5.Sum(r.queryString())
	return strings.ToUpper(fmt.Sprintf("%x", sum))
}

// 向网关发送post请求
func post(content []byte) ([]byte, error) {
	fmt.Println(string(content))
	resp, err := http.Post(GatewayUrl, "application/xml", strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	return io.ReadAll(resp.Body)
}

func main() {
	if len(os.Args) < 2 {
		hint()
	}

	var req Req
	switch os.Args[1] {
	case "query":
		flags, err := handleSubCommand("query", map[string]string{
			"out_auth_no": "商户授权号",
			"auth_no":     "平台授权号",
		}, os.Args[2:])
		if err != nil {
			panic(err)
		}
		if flags == nil {
			os.Exit(-1)
		}
		req = Req{
			"service":     Query,
			"mch_id":      MchId,
			"nonce_str":   string(randStr(NonceStrLen)),
			"out_auth_no": *(flags["out_auth_no"]),
			"auth_no":     *(flags["auth_no"]),
		}
		req["sign"] = req.sign()
	case "unfreeze":
		flags, err := handleSubCommand("unfreeze", map[string]string{
			"auth_no":   "平台授权号",
			"total_fee": "解冻金额",
			"remark":    "解冻描述",
		}, os.Args[2:])
		if err != nil {
			panic(err)
		}
		if flags == nil {
			os.Exit(-1)
		}
		req = Req{
			"service":        UnFreeze,
			"mch_id":         MchId,
			"nonce_str":      string(randStr(NonceStrLen)),
			"out_request_no": string(randStr(NonceStrLen)),
			"total_fee":      *(flags["total_fee"]),
			"auth_no":        *(flags["auth_no"]),
			"remark":         *(flags["remark"]),
		}
		req["sign"] = req.sign()
	default:
		hint()
	}
	xmlStr, err := xml.MarshalIndent(req, "", "  ")
	if err != nil {
		panic(err)
	}
	res, err := post(xmlStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(res))
}
