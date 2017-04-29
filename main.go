//CloudXns-DDNS by 哈陆lu
//用Golang写的CloudXns的DDNS服务
//Github https://github.com/lzjluzijie/cloudxns-ddns
//Blog https://halu.lu

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	json "github.com/bitly/go-simplejson"
	"github.com/prestonTao/upnp"
)

var ak string
var sk string
var domain string
var host string
var get string
var t string

//初始化程序 获取参数
func init() {
	flag.StringVar(&ak, "a", "", "Your CloudXns AccessKey")
	flag.StringVar(&sk, "s", "", "Your CloudXns SecretKey")
	flag.StringVar(&domain, "d", "", "Your domain `(eg halu.lu)`")
	flag.StringVar(&host, "h", "", "Your host `(eg home)`")
	flag.StringVar(&get, "g", "upnp", "Use url or upnp to get IP `(eg http://ip.halu.lu:8233)`")
	flag.StringVar(&t, "t", "600s", "Flush time `(eg 60s)`")
	flag.Parse()

	if ak == "" || sk == "" {
		fmt.Println("Please enter your accesskey and secretkey ")
		os.Exit(1)
	}

	if domain == "" {
		fmt.Println("Please enter your domain ")
		os.Exit(1)
	}

	if host == "" {
		fmt.Println("Please enter your host ")
		os.Exit(1)
	}

	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
}

func main() {
	fmt.Println("Start")
	var ip string
	for {
		//开始

		if get == "upnp" {
			ip = getIPu()
			fmt.Printf("Using upnp to get ip\n")
		} else {
			ip = getIPr(get)
			fmt.Printf("Using %s to get ip\n", get)
		}

		domainID, dCode := getDomainID()
		if dCode == 1 {
			recordID, rCode := getRecordID(domainID)
			if rCode == 1 {
				sCode := setRecord(domainID, recordID, host, ip)
				if sCode == 1 {
					fmt.Printf("Successfully updated at %v\nIP %s\n", time.Now(), ip)
				} else {
					fmt.Printf("Set record id err, code %d\n", sCode)
				}
			} else {
				fmt.Printf("Get record id err, code %d\n", rCode)
			}
		} else {
			fmt.Printf("Get domain id err, code %d\n", dCode)
		}

		f, err := time.ParseDuration(t)
		checkErr(err)
		time.Sleep(f)
	}
}

//获取域名ID
func getDomainID() (domainID string, code int) {
	now := time.Now()
	url := "https://www.cloudxns.net/api2/domain"
	hmac := getHMAC(url, "", now.Format(time.RFC1123Z))
	req, err := http.NewRequest("GET", url, nil)
	checkErr(err)
	req.Header.Set("API-KEY", ak)
	req.Header.Set("API-REQUEST-DATE", now.Format(time.RFC1123Z))
	req.Header.Set("API-HMAC", hmac)
	resp, err := http.DefaultClient.Do(req)
	checkErr(err)
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	defer resp.Body.Close()

	json := json.New()
	err = json.UnmarshalJSON(body)
	checkErr(err)
	code = json.Get("code").MustInt()
	if code == 1 {
		length, err := strconv.Atoi(json.Get("total").MustString())
		checkErr(err)
		for i := 0; i < length; i++ {
			if json.Get("data").GetIndex(i).Get("domain").MustString() == domain {
				return json.Get("data").GetIndex(i).Get("id").MustString(), code
			}
		}
	}
	return "err", code
}

//获取记录ID
func getRecordID(domainID string) (recordID string, code int) {
	now := time.Now()
	url := "https://www.cloudxns.net/api2/record/" + domainID + "?host_id=0&row_num=500"
	hmac := getHMAC(url, "", now.Format(time.RFC1123Z))
	req, err := http.NewRequest("GET", url, nil)
	checkErr(err)
	req.Header.Set("API-KEY", ak)
	req.Header.Set("API-REQUEST-DATE", now.Format(time.RFC1123Z))
	req.Header.Set("API-HMAC", hmac)
	resp, err := http.DefaultClient.Do(req)
	checkErr(err)
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	defer resp.Body.Close()

	json := json.New()
	err = json.UnmarshalJSON(body)
	checkErr(err)
	code = json.Get("code").MustInt()
	if code == 1 {
		length, err := strconv.Atoi(json.Get("total").MustString())
		checkErr(err)
		for i := 0; i < length; i++ {
			if json.Get("data").GetIndex(i).Get("host").MustString() == host {
				return json.Get("data").GetIndex(i).Get("record_id").MustString(), code
			}
		}
	}
	return "err", code
}

//设置解析记录
func setRecord(domainID, recordID, host, ip string) (code int) {
	now := time.Now()
	url := "https://www.cloudxns.net/api2/record/" + recordID
	buf := &bytes.Buffer{}
	data := json.New()
	data.Set("domain_id", domainID)
	data.Set("host", host)
	data.Set("value", ip)
	str, err := data.Encode()
	checkErr(err)
	buf.WriteString(string(str))
	hmac := getHMAC(url, buf.String(), now.Format(time.RFC1123Z))
	req, err := http.NewRequest("PUT", url, buf)
	checkErr(err)
	req.Header.Set("API-KEY", ak)
	req.Header.Set("API-REQUEST-DATE", now.Format(time.RFC1123Z))
	req.Header.Set("API-HMAC", hmac)
	resp, err := http.DefaultClient.Do(req)
	checkErr(err)
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	defer resp.Body.Close()

	json := json.New()
	err = json.UnmarshalJSON(body)
	checkErr(err)
	return json.Get("code").MustInt()
}

//通过UPNP获取IP
func getIPu() string {
	upnp := new(upnp.Upnp)
	err := upnp.ExternalIPAddr()
	checkErr(err)
	return upnp.GatewayOutsideIP
}

//通过远程网站获取IP
func getIPr(url string) (ip string) {
	req, err := http.NewRequest("GET", url, nil)
	checkErr(err)
	resp, err := http.DefaultClient.Do(req)
	checkErr(err)
	body, err := ioutil.ReadAll(resp.Body)
	ip = string(body)
	checkErr(err)
	defer resp.Body.Close()
	return ip
}

func getHMAC(url, data, time string) string {
	hash := md5.Sum([]byte(ak + url + data + time + sk))
	return hex.EncodeToString(hash[:])
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
