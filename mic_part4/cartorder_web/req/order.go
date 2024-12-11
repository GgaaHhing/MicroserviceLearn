package req

type OrderReq struct {
	Id        int32  `json:"id,omitempty"`
	AccountId int32  `json:"accountId"`
	Addr      string `json:"addr"`
	PostCode  string `json:"postCode"` //邮编
	Receiver  string `json:"receiver"`
	Mobile    string `json:"mobile"`
	PayType   string `json:"payType"`
}
