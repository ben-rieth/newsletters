package types

type AuthInput struct {
	Body struct {
		Email string `json:"email" doc:"Must be a valid email" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
		Password string `json:"password" minLength:"8" maxLength:"80"`
	}
}

type AuthOutputBody struct {
	Token string `json:"token"`
}

type AuthOutput struct {
	Body AuthOutputBody
}