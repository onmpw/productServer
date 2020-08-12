package include

import "time"

type Data struct {
	Platform, OrderStatus string
	OrderInfo             OrderTradeContract
}

type OrderTradeContract interface {
	BuildData(orderStatus string)	error
	Send()	bool
}

var (
	ShopList []*ShopInfo
	TypeNum           = 2

	C = make(chan int, TypeNum) // channel 用于控制多协程

	DateTimeFormat = "2006-01-02 15:04:05"
)

func GetNewShopTime() string{
	now := time.Now()
	return time.Unix(now.Unix()-60*60*24, 0).Format(DateTimeFormat)
}

func Now() string {
	now := time.Now()
	return time.Unix(now.Unix(), 0).Format(DateTimeFormat)
}

