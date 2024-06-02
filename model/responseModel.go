package model

type GoogleResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type SuccessResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"Error message"`
}
