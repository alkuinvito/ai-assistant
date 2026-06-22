package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"time"

	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/apperror"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ACCESS_TOKEN_EXPIRES_IN = 5 * 60        // 5 minutes
var REFRESH_TOKEN_EXPIRES_IN = 24 * 60 * 60 // 24 hours

type AuthService interface {
	comparePassword(password, hash string) error
	generateTokenPair(userId string) (*DTOAuthResponse, error)
	hashPassword(password string) (string, error)
	Login(ctx context.Context, req *DTOLoginRequest) (*DTOAuthResponse, error)
	Register(ctx context.Context, req *DTORegisterRequest) error
}

type authService struct {
	db          *gorm.DB
	userService users.UserService
}

func NewAuthService(db *gorm.DB, userService users.UserService) AuthService {
	return &authService{
		db:          db,
		userService: userService,
	}
}

func (s *authService) comparePassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return err
	}

	return nil
}

func (s *authService) generateTokenPair(userId string) (*DTOAuthResponse, error) {
	key := os.Getenv("JWT_PRIVATE_KEY")
	tokenId := rand.Text()
	issuer := os.Getenv("JWT_ISSUER")
	subject := userId
	accessTokenExp := time.Now().Add(time.Second * time.Duration(ACCESS_TOKEN_EXPIRES_IN)).Unix()
	refreshTokenExp := time.Now().Add(time.Second * time.Duration(REFRESH_TOKEN_EXPIRES_IN)).Unix()
	issuedAt := time.Now().Unix()

	if key == "" {
		return nil, errors.New("JWT_PRIVATE_KEY environment variable not set")
	}

	if issuer == "" {
		return nil, errors.New("JWT_ISSUER environment variable not set")
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(key))
	if err != nil {
		return nil, err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			"jti": tokenId,
			"iss": issuer,
			"sub": subject,
			"exp": accessTokenExp,
			"iat": issuedAt,
		})
	t.Header["typ"] = "at+jwt"

	accessToken, err := t.SignedString(privateKey)
	if err != nil {
		return nil, err
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jti": tokenId,
		"iss": issuer,
		"sub": subject,
		"exp": refreshTokenExp,
		"iat": issuedAt,
	})
	rt.Header["typ"] = "refresh+jwt"

	refreshToken, err := rt.SignedString(privateKey)
	if err != nil {
		return nil, err
	}

	return &DTOAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    ACCESS_TOKEN_EXPIRES_IN,
	}, nil
}

func (s *authService) hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *authService) Login(ctx context.Context, req *DTOLoginRequest) (*DTOAuthResponse, error) {
	tx := s.db.WithContext(ctx)

	user, err := s.userService.GetUserByEmail(tx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewUnauthorized(ctx, "Email or password is incorrect", err)
		}
		return nil, apperror.NewInternalServerError(ctx, "Failed to retrieve user data", err)
	}

	err = s.comparePassword(req.Password, user.Password)
	if err != nil {
		return nil, apperror.NewUnauthorized(ctx, "Email or password is incorrect", err)
	}

	tokenPair, err := s.generateTokenPair(user.Id.String())
	if err != nil {
		return nil, apperror.NewInternalServerError(ctx, "Failed to generate token pair", err)
	}

	return tokenPair, nil
}

func (s *authService) Register(ctx context.Context, req *DTORegisterRequest) error {
	tx := s.db.WithContext(ctx)

	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return apperror.NewInternalServerError(ctx, "Failed to create new user", err)
	}

	userId, err := uuid.NewV7()
	if err != nil {
		return apperror.NewInternalServerError(ctx, "Failed to create new user", err)
	}

	newUser := &users.User{
		Id:       userId,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = s.userService.CreateUser(tx, newUser)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.NewConflict(ctx, "User already exists", err)
		}
		return apperror.NewInternalServerError(ctx, "Failed to create new user", err)
	}

	return nil
}
