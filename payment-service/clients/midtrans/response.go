package clients

type MidTransResponse struct {
	Code   int          `json:"code"`
	Status string       `json:"status"`
	Data   MidTransData `json:"data"`
}
type MidTransData struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}
