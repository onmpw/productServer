package include

type ShopInfo struct {
	Sid		int
	Cid		int
	Is_auto_product_add 	string
	Type 	string
}

func (s *ShopInfo) TableName() string {
	return "shop_taobao"
}


type OrderThirdSyncTime struct {
	Type,Platform,Updatetime 	string
	Sid, Company_id	int
	Created	string
}

type RefundThirdSyncTime struct {
	Platform,Updatetime 	string
	Sid, Company_id	int
	Created	string
}

type ProductThirdSyncTime struct {
	Platform,Updatetime 	string
	Sid, Company_id	int
	Created	string
}

func (o *OrderThirdSyncTime) TableName() string {
	return "order_thirdsync_time"
}

func (r *RefundThirdSyncTime) TableName() string {
	return "refund_thirdsync_time"
}

func (r *ProductThirdSyncTime) TableName() string {
	return "product_thirdsync_time"
}