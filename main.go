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
	"reflect"
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
	//UnFreeze    = "unified.auth.unfreeze"
	//Reverse    = "unified.auth.reverse"
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

// 构造出用来计算签名的查询字符串
func queryString(r interface{}) []byte {
	params := convertMap(r)
	var fields sort.StringSlice
	for s := range params {
		fields = append(fields, s)
	}
	fields.Sort()
	str := ""
	for i := 0; i < fields.Len(); i++ {
		str += fmt.Sprintf("%s=%s&", fields[i], params[fields[i]])
	}

	return []byte(strings.TrimSuffix(str, "&") + "&key=" + Key)
}

func convertMap(i interface{}) map[string]interface{} {
	v := reflect.ValueOf(i)
	m := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.String {
			name := v.Type().Field(i).Name
			if name == "XMLName" || name == "Sign" {
				continue
			}
			key := v.Type().Field(i).Tag.Get("xml")
			if len(key) == 0 || strings.Contains(key, "omitempty") && v.Field(i).IsZero() {
				continue
			}
			key = strings.TrimSuffix(key, ",omitempty")
			m[key] = v.Field(i).Interface()
		}
		if v.Field(i).Kind() == reflect.Struct {
			for s, i2 := range convertMap(v.Field(i).Interface()) {
				m[s] = i2
			}
		}
	}
	return m
}

// 计算签名，需要查询字符串
func sign(queryString []byte) string {
	sum := md5.Sum(queryString)
	return strings.ToUpper(fmt.Sprintf("%x", sum))
}

// 格式化成xml
func formatXml(r interface{}) ([]byte, error) {
	return xml.MarshalIndent(r, "", "  ")
}

// 向网关发送post请求
func post(content []byte) ([]byte, error) {
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

	switch os.Args[1] {
	case "query":
		flags, err := handleSubCommand("query", map[string]string{
			"out_auth_no": "第三方商户号",
			"auth_no":     "商户号",
		}, os.Args[2:])
		if err != nil {
			panic(err)
		}
		req := QueryReq{
			Req: Req{
				Service:  Query,
				MchId:    MchId,
				NonceStr: string(randStr(NonceStrLen)),
			},
			OutAuthNo: *(flags["out_auth_no"]),
			AuthNo:    *(flags["auth_no"]),
		}
		req.Sign = sign(queryString(req))

		xmlStr, err := formatXml(req)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(xmlStr))
		res, err := post(xmlStr)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(res))
		break
	case "unfreeze":
		break
	default:
		hint()
		break
	}
}
