package Jd

import (
	"encoding/json"
	"fmt"
	"github.com/onmpw/JYGO/config"
	"github.com/onmpw/JYGO/model"
	"productServer/http"
	"productServer/include"
	"reflect"
	"strconv"
	"strings"
)

var platform = "J"

func (o *OrderInfo) BuildData(orderStatus string) error{
	var start string
	var end string
	var flag bool
	o.orderStatus = orderStatus
	o.order = o.order[0:0]
	for _,shop := range include.ShopList {
		var trades []*OrderTrade
		if shop.Type != platform {
			continue
		}
		var t *include.ProductThirdSyncTime
		count := model.Read(new(include.ProductThirdSyncTime)).Filter("platform",platform).Filter("company_id",shop.Cid).Filter("sid",shop.Sid).Count()

		if count >= 1 {
			err := model.Read(new(include.ProductThirdSyncTime)).Filter("platform", platform).Filter("company_id", shop.Cid).Filter("sid", shop.Sid).GetOne(&t)
			if err != nil {
				return err
			}
			start = t.Created
			flag = true    // true 表示记录存在 需要更新
		}else {
			start = include.GetNewShopTime()
			flag = false  // false 表示记录不存在 需要新增
		}
		end = include.Now()

		getParam := map[string]string{
			"cid": strconv.Itoa(shop.Cid),
			"sid": strconv.Itoa(shop.Sid),
			"page": "1",
			"pageSize":"200",
			"startDate": start,
			"endDate":end,
		}
		num,_ := getOrder(getParam,&trades)
		// 获取订单
		//num , _ := model.Read(new(OrderTrade)).Filter("type", OrderStatus[orderStatus]).Filter("cid", shop.Cid).Filter("sid", shop.Sid).Filter("modified",">=",start).Filter("modified","<",end).GetAll(&trades)
		o.SyncTime[shop.Sid] = start
		o.AddOrUp[shop.Sid] = flag
		o.SidToCid[shop.Sid] = shop.Cid
		if num <= 0 {
			continue
		}

		o.order = append(o.order,trades...)
		o.getMaxTime(trades,shop.Sid)
	}

	return nil
}

func getOrder(param map[string]string,models *[]*OrderTrade)(num int,err error ){
	jsons, err := json.Marshal(param)
	if err != nil {
		fmt.Printf("京东商品同步错误：%v",err)
		return 0,err
	}

	trades,_ := http.Get(string(jsons),"Provider\\OrderService@getProduct",config.Conf.C("jd_api_host"))

	*models = parseOrderList(trades)

	return len(*models),err
}

func parseOrderList(value interface{}) (orderList []*OrderTrade) {
	rv := reflect.ValueOf(value)

	orderList = make([]*OrderTrade,rv.Len())
	if rv.Kind() == reflect.Slice {
		l := rv.Len()

		for i:=0;i<l;i++ {
			iv := reflect.Indirect(rv.Index(i))

			r := iv.Interface().(map[string]interface{})

			var trade = new(OrderTrade)
			trade.Id = int(r["id"].(float64))
			trade.Cid = int(r["cid"].(float64))
			trade.Sid = int(r["sid"].(float64))
			trade.Pid = r["pid"].(string)
			trade.Response = r["response"].(string)
			trade.Created = r["created"].(string)
			trade.Modified = r["modified"].(string)
			trade.Type = r["type"].(string)
			orderList[i] = trade
		}
	}
	return orderList
}

func (o *OrderInfo) Send() bool {
	var order string
	if len(o.order) > 0{
		jsons, err := json.Marshal(o.order)

		if err != nil {
			fmt.Printf("京东商品同步错误：%v",err)
			return false
		}

		order  = string(jsons)
	}

	data := map[string]string {
		"platform":"jd",
		"order_status":o.orderStatus,
		"product_list":order,
	}

	jsons, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("京东商品同步错误：%v",err)
		return false
	}
	go o.updateSyncTime()
	return http.Exec(string(jsons))
}

func (o *OrderInfo) getMaxTime(trades []*OrderTrade,sid int) {
	if len(trades) <= 0 {
		return
	}

	for _,trade := range trades {
		if strings.Compare(trade.Modified,o.SyncTime[sid]) == 1 {
			o.SyncTime[sid] = trade.Modified
		}
	}
}

func (o *OrderInfo) updateSyncTime() {
	var syncTime include.ProductThirdSyncTime

	syncTime.Platform = platform
	syncTime.Updatetime = include.Now()
	for sid,created := range o.SyncTime {
		syncTime.Sid = sid
		syncTime.Created = created
		syncTime.Company_id = o.SidToCid[sid]
		if o.AddOrUp[sid] { // 需要更新
			where := []interface{}{[]interface{}{"company_id",o.SidToCid[sid]},[]interface{}{"platform",platform},[]interface{}{"sid",sid}}
			_ , err := model.Update(syncTime,where)
			if err != nil {
				fmt.Println(err)
			}
		}else {
			_,err := model.Add(syncTime)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
