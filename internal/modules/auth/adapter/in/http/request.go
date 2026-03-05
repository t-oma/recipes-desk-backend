package httphandler

type RegisterRequest struct {
	Email     string `json:"email"     binding:"required,email"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName"  binding:"required"`
	Password  string `json:"password"  binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}
