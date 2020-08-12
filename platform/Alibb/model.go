package Alibb

type OrderTrade struct {
	Id				int
	Oid				string
	Aid 			string
	Response		string
	Cid				int
	Created			string
	Modified		string
	Type 			string
	Sid 			int
}

type OrderInfo struct {
	order []*OrderTrade
	orderStatus string
	SyncTime map[int]string
	AddOrUp	map[int]bool
	SidToCid map[int]int
}

func (o *OrderTrade) TableName() string {
	return "jdp_alibb_order_refund_trade"
}
