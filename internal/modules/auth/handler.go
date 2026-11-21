package auth

import (
	"context"
	"errors"

	"github.com/fury175/forge_api/internal/apperrors"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(ctx context.Context, input *CreatUserPayload) (*CreateUserRespone, error) {
	user, err := h.service.CreateUser(ctx, *input)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountAlreadyExists):
			return nil, apperrors.BadRequestError(ctx, err)
		default:
			return nil, apperrors.InternalServerError(ctx, err)
		}
	}

	return &CreateUserRespone{Body: user}, nil
}

func (h *Handler) LoginUser(ctx context.Context, input *LoginPayload) (*TokenOutput, error) {
	token, err := h.service.LoginUser(ctx, *input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, apperrors.UnauthorizedAccessError(ctx, err)
		}
		return nil, apperrors.InternalServerError(ctx, err)
	}
	return &TokenOutput{Body: TokenBody{Token: token}}, nil
}
