package bmdatabase

type BmUser struct {
	Idusername int64  `json:"idusername"`
	Mail       string `json:"mail"`
	Token      string `json:"token"`
}
