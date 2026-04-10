package types

type AuthRequest struct {
	Email string `json:"email" doc:"Must be a valid email" pattern:"^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"`
	Password string `json:"password" minLength:"8" maxLength:"80"`
}

type AuthResponse struct {
	Token string `json:"token"`
}