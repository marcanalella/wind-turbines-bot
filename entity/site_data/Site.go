package site_data

type Site struct {
	DataStatus string `json:"data_status"`
	Ts         string `json:"ts"`
	Device     Device `json:"device"`
}
