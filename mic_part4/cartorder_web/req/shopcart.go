package req

type ShopCartReq struct {
	Id        int32 `json:"id"`
	AccountId int32 `json:"account_id"`
	ProductId int32 `json:"product_id"`
	Num       int32 `json:"num"`
	Checked   bool  `json:"checked"`
}
