package site_data

type Device struct {
	Power     string `json:"power"`
	Wndspd    string `json:"wndspd"`
	Ambtmp    string `json:"ambtmp"`
	Energy    string `json:"energy"`
	Warning   string `json:"warning"`
	Env       string `json:"env"`
	DevId     string `json:"dev_id"`
	Yawpos    string `json:"yawpos"`
	Yawerr    string `json:"yawerr"`
	Operating string `json:"operating"`
	Faulted   string `json:"faulted"`
	Ext       string `json:"ext"`
}
