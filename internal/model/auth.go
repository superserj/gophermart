package model

// AuthRequest — тело запросов register и login.
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
