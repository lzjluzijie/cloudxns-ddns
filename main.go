package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	json "github.com/bitly/go-simplejson"
	"github.com/prestonTao/upnp"
)

var ak string
var sk string
var domain string
var host string

func init() {
	flag.StringVar(&ak, "ak", "", "CloudXns AccessKey")
	flag.StringVar(&sk, "sk", "", "CloudXns SecretKey")
	flag.StringVar(&domain, "domain", "", "Domain")
	flag.StringVar(&host, "host", "", "Host")
	flag.Parse()
}

func main() {
	fmt.Println("Start")

	for {
		ip := getIP()
		domainID := getDomainID()
		recordID := getRecordID(domainID)
		setRecord(domainID, recordID, host, ip)
		fmt.Printf("Updated at %v\nIP %s\n", time.Now(), ip)
		time.Sleep(time.Second * 60)
	}

}

//获取域名ID
func getDomainID() (domainID int) {
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
	length, err := strconv.Atoi(json.Get("total").MustString())
	checkErr(err)
	for i := 0; i < length; i++ {
		if json.Get("data").GetIndex(i).Get("domain").MustString() == domain {
			domainID, err = strconv.Atoi(json.Get("data").GetIndex(i).Get("id").MustString())
			checkErr(err)
		}
	}
	return domainID
}

//获取记录ID
func getRecordID(domainID int) (recordID int) {
	now := time.Now()
	url := "https://www.cloudxns.net/api2/record/" + strconv.Itoa(domainID) + "?host_id=0&row_num=500"
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
	length, err := strconv.Atoi(json.Get("total").MustString())
	checkErr(err)
	for i := 0; i < length; i++ {
		if json.Get("data").GetIndex(i).Get("host").MustString() == host {
			recordID, err = strconv.Atoi(json.Get("data").GetIndex(i).Get("record_id").MustString())
			checkErr(err)
		}
	}
	return recordID
}

//设置解析记录
func setRecord(domain, record int, host, ip string) {
	now := time.Now()
	url := "https://www.cloudxns.net/api2/record/" + fmt.Sprintf("%d", record)
	buf := &bytes.Buffer{}
	data := json.New()
	data.Set("domain_id", domain)
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
	//body, err := ioutil.ReadAll(resp.Body)
	//checkErr(err)
	defer resp.Body.Close()
}

func getIP() string {
	upnp := new(upnp.Upnp)
	err := upnp.ExternalIPAddr()
	checkErr(err)
	return upnp.GatewayOutsideIP
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
