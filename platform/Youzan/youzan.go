package Youzan

import (
	"encoding/json"
	"fmt"
	"github.com/onmpw/JYGO/model"
	"orderServer/http"
	"orderServer/include"
	"strings"
)

var OrderStatus = map[string]string {
	"WAIT_SELLER_SEND":"WAIT_SELLER_SEND_GOODS",
	"WAIT_BUYER_CONFIRM":"WAIT_BUYER_CONFIRM_GOODS",
	"TRADE_SUCCESS":"TRADE_SUCCESS",
}
var platform = "Y"

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
		var t *include.OrderThirdSyncTime
		count := model.Read(new(include.OrderThirdSyncTime)).Filter("platform",platform).Filter("type",OrderStatus[orderStatus]).Filter("company_id",shop.Cid).Filter("sid",shop.Sid).Count()

		if count >= 1 {
			err := model.Read(new(include.OrderThirdSyncTime)).Filter("platform", platform).Filter("type", OrderStatus[orderStatus]).Filter("company_id", shop.Cid).Filter("sid", shop.Sid).GetOne(&t)
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
		// 获取订单
		num , _ := model.Read(new(OrderTrade)).Filter("type", OrderStatus[orderStatus]).Filter("cid", shop.Cid).Filter("sid", shop.Sid).Filter("modified",">=",start).Filter("modified","<",end).GetAll(&trades)
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

func (o *OrderInfo) Send() bool {
	var order string
	if len(o.order) > 0{
		jsons, err := json.Marshal(o.order)

		if err != nil {
			return false
		}

		order  = string(jsons)
	}

	data := map[string]string {
		"platform":"youzan",
		"order_status":o.orderStatus,
		"order_list":order,
	}

	jsons, err := json.Marshal(data)
	if err != nil {
		return false
	}
	o.updateSyncTime()
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
	var syncTime include.OrderThirdSyncTime

	syncTime.Type = OrderStatus[o.orderStatus]
	syncTime.Platform = platform
	syncTime.Updatetime = include.Now()
	for sid,created := range o.SyncTime {
		syncTime.Sid = sid
		syncTime.Created = created
		syncTime.Company_id = o.SidToCid[sid]
		if o.AddOrUp[sid] { // 需要更新
			where := []interface{}{[]interface{}{"company_id",o.SidToCid[sid]},[]interface{}{"platform",platform},[]interface{}{"sid",sid},[]interface{}{"type",OrderStatus[o.orderStatus]}}
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
