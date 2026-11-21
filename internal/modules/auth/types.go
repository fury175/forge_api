package auth

import repo "github.com/fury175/forge_api/internal/adapters/postgresql/sqlc"

type CreatUserPayload struct {
	Body struct {
		Username string `json:"username" minLength:"2"`
		Email    string `json:"email" format:"email"`
		Password string `json:"password" minLength:"6"`
	}
}

type LoginPayload struct {
	Body struct {
		Email    string `json:"email" format:"email"`
		Password string `json:"password" minLength:"2"`
	}
}

type CreateUserRespone struct {
	Body repo.CreateUserRow
}

type WelcomeEmailData struct {
	Username      string
	ActivationURL string
}

type TokenBody struct {
	Token string `json:"token"`
}

type TokenOutput struct {
	Body TokenBody
}
