package http

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/onmpw/JYGO/config"
	"io"
	"io/ioutil"
	"net/http"
	"orderServer/include"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	SendData map[string]string
)

const (
	AppSecret = "3270064d0afe3be41ae838cd9e667b1c"
	AppId     = "1001"
)

func buildPostData() map[string]string{
	return map[string]string{
		"head":"",
		"app_id":AppId,
		"nonce": Md5(strconv.FormatInt(time.Now().Unix(),10)),
		"ip":"",
		"method":"Provider\\SyncOrderService@productSyncDistribute",
	}
}

func setPostData(SendData *map[string]string,key string, val string) {
	(*SendData)[key] = val
}

func Exec(value string) bool {
	sendData := buildPostData()
	setPostData(&sendData,"data",value)
	sign := createSign(sendData)
	setPostData(&sendData,"sign",sign)
	res,err := post(sendData)
	if err != nil {
		return false
	}

	result := parseResult(res)

	if result["errno"] != "0" {
		fmt.Println(include.Now()+":"+result["errmsg"])
		return false
	}

	return true

}

func Get(param string,method string,host string) (interface{}, error) {
	sendData := buildPostData()
	setPostData(&sendData,"data",param)
	setPostData(&sendData,"method",method)
	sign := createSign(sendData)
	setPostData(&sendData,"sign",sign)

	return get(sendData,host,true)
}

func get(SendData map[string]string,host string,decrypt bool) (interface{},error) {
	jsons , _ := json.Marshal(SendData)
	requestBody := string(jsons)
	res, err := http.Post(host, "application/json;charset=utf-8", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil,err
	}

	body, err := ioutil.ReadAll(res.Body)


	if decrypt {
		r := decryptor(string(body))
		return r,nil
	}

	var result = make(map[string]interface{})
	err = json.Unmarshal(body,&result)

	if err != nil {
		return nil,err
	}
	return result["product"],nil
}

func decryptor(data string) interface{} {
	sendData := buildPostData()
	setPostData(&sendData,"data",data)
	setPostData(&sendData,"method","Provider\\DecryptService@decrypt")
	sign := createSign(sendData)
	setPostData(&sendData,"sign",sign)

	res,err := get(sendData,config.Conf.C("api_host"),false)

	if err != nil {
		return nil
	}

	return res
}

func parseResult(value interface{}) (res map[string]string) {
	rv := reflect.ValueOf(value)

	iter := rv.MapRange()

	res = make(map[string]string)

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		str := fmt.Sprintf("%v",val)
		res[key.String()] = str
	}

	return res
}

func post(SendData map[string]string) (interface{},error) {
	jsons , _ := json.Marshal(SendData)
	requestBody := string(jsons)
	res, err := http.Post(config.Conf.C("api_host"), "application/json;charset=utf-8", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil,err
	}

	body, err := ioutil.ReadAll(res.Body)

	var result = make(map[string]interface{})
	err = json.Unmarshal(body,&result)

	if err != nil {
		return nil,err
	}

	return result,nil
}

func createSign(SendData map[string]string)(sign  string){
	var contact []string
	for key,val := range SendData {
		contact = append(contact,key+val)
	}
	sort.Strings(contact)
	sign = AppSecret
	for _,str := range contact {
		sign += str
	}
	sign += AppSecret

	return strings.ToUpper(Md5(sign))
}


func Md5(value string) string {
	w := md5.New()
	_,err := io.WriteString(w,value)

	if err != nil {
		return "error"
	}

	return fmt.Sprintf("%x",w.Sum(nil))
}
