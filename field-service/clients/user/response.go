package clients

import "github.com/google/uuid"

type UserResponse struct {
	Code    int      `json:"code"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    UserData `json:"data"`
}

type UserData struct {
	UUID        uuid.UUID `json:"uuid"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	PhoneNumber string    `json:"phoneNumber"`
}
