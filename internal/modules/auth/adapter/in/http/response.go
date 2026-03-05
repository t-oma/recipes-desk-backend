package httphandler

type UserResponse struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	CreatedAt         string `json:"createdAt"`
	PasswordUpdatedAt string `json:"passwordUpdatedAt"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}
