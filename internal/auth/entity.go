package auth

type DTOLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

type DTOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type DTORegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=32"`
	Name     string `json:"name" validate:"required,min=3,max=32"`
}

type DTOSendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type DTOVerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type DTOMessageResponse struct {
	Message string `json:"message"`
}
