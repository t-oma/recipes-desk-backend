package dto

type RegisterInput struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshTokensInput struct {
	RefreshToken string
}
