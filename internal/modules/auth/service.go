package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/fury175/forge_api/internal/adapters/postgresql/pgdb"
	repo "github.com/fury175/forge_api/internal/adapters/postgresql/sqlc"
	"github.com/fury175/forge_api/internal/mailer"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
)

type Service interface {
	CreateUser(ctx context.Context, args CreatUserPayload) (repo.CreateUserRow, error)
	LoginUser(ctx context.Context, args LoginPayload) (string, error)
}

type svc struct {
	repo        repo.Querier
	db          pgdb.DB
	mailer      mailer.Client
	env         string
	frontendUrl string
}

func NewService(repo repo.Querier, db *pgxpool.Pool, mailer mailer.Client, env string, frontendUrl string) Service {
	return &svc{repo: repo, db: pgdb.New(db), mailer: mailer, env: env, frontendUrl: frontendUrl}
}

func (s *svc) LoginUser(ctx context.Context, args LoginPayload) (string, error) {
	user, err := s.repo.GetUserByEmailAndPassword(ctx, args.Body.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	// compare passwords
	if err := s.comparePassword(args.Body.Password, user.Password); err != nil {
		return "", ErrInvalidCredentials
	}

	return "user_token", nil

}

func (s *svc) CreateUser(ctx context.Context, args CreatUserPayload) (repo.CreateUserRow, error) {
	if _, err := s.checkIfUserExists(ctx, args.Body.Email); err == nil {
		return repo.CreateUserRow{}, ErrAccountAlreadyExists
	}
	hash, err := s.hashPassword(args.Body.Password)
	if err != nil {
		return repo.CreateUserRow{}, err
	}
	var user repo.CreateUserRow
	if err := s.db.WithTransaction(ctx, func(q repo.Querier) error {
		var txError error

		user, txError = q.CreateUser(ctx, repo.CreateUserParams{
			Username: args.Body.Username,
			Password: hash,
			Email:    args.Body.Email,
		})
		if txError != nil {
			return txError
		}

		token, txError := s.generateInviteToken(ctx, q, user.ID)
		if txError != nil {
			return txError
		}

		isSandbox := s.env == "development"
		activationUrl := s.frontendUrl + "/confirm/" + token
		vars := WelcomeEmailData{
			ActivationURL: activationUrl,
			Username:      user.Username,
		}

		status, txError := s.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, isSandbox)
		if txError != nil {
			return txError
		}

		fmt.Println("Email sent with status", status)
		return nil
	}); err != nil {
		return repo.CreateUserRow{}, err
	}

	return user, nil
}

// helper functions
func (s *svc) generateInviteToken(ctx context.Context, q repo.Querier, userId int64) (string, error) {
	token, err := uuid.NewV6()
	if err != nil {
		return "", err
	}

	plainToken := token.String()
	hash := sha256.Sum256([]byte(plainToken))
	hashedToken := hex.EncodeToString(hash[:])

	exp := pgtype.Timestamptz{Time: time.Now().Add(time.Hour * 3), Valid: true}

	if err := q.GenerateInvite(ctx, repo.GenerateInviteParams{
		UserID:   userId,
		InviteID: hashedToken,
		Expiry:   exp,
	}); err != nil {
		return "", err
	}

	return plainToken, nil
}

func (s *svc) checkIfUserExists(ctx context.Context, email string) (repo.GetUserByEmailRow, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return repo.GetUserByEmailRow{}, err
	}

	return user, nil
}

func (s *svc) hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (s *svc) comparePassword(password string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}
