package main

import (
	"flag"
	"fmt"
	"github.com/onmpw/JYGO/config"
	"github.com/onmpw/JYGO/model"
	"io/ioutil"
	"os"
	"os/signal"
	"productServer/include"
	"productServer/platform/Alibb"
	"productServer/platform/Jd"
	"productServer/platform/Pdd"
	"strconv"
	"syscall"
	"time"
)

var (
	pid         = os.Getpid()
	stop        = false
	childFinish = false
	gloPidFile string
	ErrInfo = "{'errCode':%d,'errMsg':'%s'}"
	ErrCode = 0
	c = make(chan int,1)
)

func main() {
	var tree = map[string]*include.Data{
		"pdd": {
				Platform:    "pdd",
				OrderStatus: "",
				OrderInfo: &Pdd.OrderInfo{
					SyncTime: make(map[int]string),
					AddOrUp:  make(map[int]bool),
					SidToCid: make(map[int]int),
				}},
		//"1688":{
		//		Platform:    "1688",
		//		OrderStatus: "",
		//		OrderInfo: &Alibb.OrderInfo{
		//			SyncTime: make(map[int]string),
		//			AddOrUp:  make(map[int]bool),
		//			SidToCid: make(map[int]int),
		//		}},
		"jd":{
				Platform:    "jd",
				OrderStatus: "",
				OrderInfo: &Jd.OrderInfo{
					SyncTime: make(map[int]string),
					AddOrUp:  make(map[int]bool),
					SidToCid: make(map[int]int),
				}},
	}

	handleSignal()

	if ! Init() {
		return
	}

	for {
		childFinish = false
		var shopList []*include.ShopInfo
		now := time.Unix(time.Now().Unix(), 0).Format(include.DateTimeFormat)

		num, _ := model.Read(new(include.ShopInfo)).Filter("is_delete", 0).Filter("end_date", ">", now).GetAll(&shopList)
		if num > 0 {
			include.ShopList = shopList
			for _, val := range tree {

				go start(val)
			}
		}
		wait()
	}

}

func start(v *include.Data) {
	defer func(){
		if e := recover(); e != nil {
			fmt.Printf(v.Platform+"平台"+v.OrderStatus+"商品同步 ERROR:%v\n",e)
			include.C <- 1
		}
	}()

	err := v.OrderInfo.BuildData(v.OrderStatus)

	if err != nil {
		fmt.Println(err)
	} else {
		_ = v.OrderInfo.Send()
	}

	include.C <- 1
}

func Init() bool{
	var iniFile = flag.String("ini", "hello", "string类型参数")
	var pidFile = flag.String("pf","./productServer.pid","进程id存放路径")
	flag.Parse()
	var r bool
	err := config.Init(*iniFile)
	if err != nil {
		errMsg := fmt.Sprintf("初始化配置文件%s失败，错误：%v\n",*iniFile,err)
		ErrCode = -3
		errStr := fmt.Sprintf(ErrInfo,ErrCode,errMsg)
		fmt.Println(errStr)
		r = false
	}
	ModelInit()
	r = ProcessInit(*pidFile)
	return r
}

func ModelInit() {
	model.Init()
	model.RegisterModel(new(Pdd.OrderTrade), new(include.ShopInfo), new(include.ProductThirdSyncTime),new(Alibb.OrderTrade))
}

func ProcessInit(pidFile string) bool{
	gloPidFile = pidFile
	// 如果文件存在 则读取进程id看是否还在运行
	if fileIsExist(pidFile){
		fi, err := os.Open(pidFile)
		if err != nil {
			errMsg := fmt.Sprintf("打开文件%s失败，错误：%v",pidFile,err)
			ErrCode = -1
			errStr := fmt.Sprintf(ErrInfo,ErrCode,errMsg)
			fmt.Println(errStr)  // -1 表示打开当前存放进程id的文件失败
			return false
		}
		defer fi.Close()
		fd, err := ioutil.ReadAll(fi)
		currPid ,err:= strconv.Atoi(string(fd))

		if err != nil {
			errMsg := fmt.Sprintf("读取进程id失败，错误：%v",err)
			ErrCode = -2
			errStr := fmt.Sprintf(ErrInfo,ErrCode,errMsg)
			fmt.Println(errStr)  // -2 表示从进程id存放文件中读取进程id，转换为整型失败
			return false
		}

		// 判断进程是否有效
		if checkProcess(currPid) {
			// 进程依然有效，则终止服务再次开启
			errMsg := fmt.Sprintf("服务正在运行，不能重复开启")
			ErrCode = 1
			errStr := fmt.Sprintf(ErrInfo,ErrCode,errMsg)
			fmt.Println(errStr) // 1 表示重复开启服务
			return false
		}
	}

	err := ioutil.WriteFile(pidFile,[]byte(strconv.Itoa(pid)),0644)
	if err != nil {
		errMsg := fmt.Sprintf("进程id：%d 写文件%s失败，错误：%v",pid,pidFile,err)
		ErrCode = 2
		errStr := fmt.Sprintf(ErrInfo,ErrCode,errMsg)
		fmt.Println(errStr) 	// 2 表示进程写文件失败
		return false
	}
	return true
}

func checkProcess(pid int) bool{
	processor,err := os.FindProcess(pid)
	if err != nil {  // 无法找到pid对应的进程，认为进程无效
		return false
	}

	err = processor.Signal(syscall.Signal(0))

	if err != nil { // 进程无效
		return false
	}
	return true
}

// fileIsExist 判断文件是否存在
// @return true 存在
// @return false 不存在
func fileIsExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}

	return true
}

func wait() {
	for i := 0; i < include.TypeNum; i++ {
		<-include.C
	}

	childFinish = true

	waitStop()

	<-time.After(time.Minute * 2)

	waitStop()
}

func handleSignal() {
	signal.Ignore(os.Interrupt)
	signal.Ignore(syscall.SIGHUP)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)
		<-c
		stop = true
		if childFinish {
			killProcess()
		}
	}()
}

func waitStop() {
	if stop {
		killProcess()
	}
}

func killProcess() {
	processor, err := os.FindProcess(pid)

	if err != nil {
		fmt.Println(err)
		return
	}

	_ = os.Remove(gloPidFile)

	_ = processor.Kill()
}
