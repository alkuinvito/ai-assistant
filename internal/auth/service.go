package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alkuinvito/ai-assistant/internal/users"
	"github.com/alkuinvito/ai-assistant/pkg/cache"
	"github.com/alkuinvito/ai-assistant/pkg/mailer"
	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/alkuinvito/ai-assistant/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ACCESS_TOKEN_EXPIRES_IN = 5 * 60        // 5 minutes
var REFRESH_TOKEN_EXPIRES_IN = 24 * 60 * 60 // 24 hours

const VERIFICATION_TOKEN_TTL = 5 * time.Minute

type AuthService interface {
	comparePassword(password, hash string) error
	GenerateAndSendVerificationToken(ctx context.Context, email string) error
	generateTokenPair(userId string) (*DTOAuthResponse, error)
	generateVerificationToken() (string, error)
	hashPassword(password string) (string, error)
	Login(ctx context.Context, req *DTOLoginRequest) (*DTOAuthResponse, error)
	Register(ctx context.Context, req *DTORegisterRequest) error
	sendEmailVerification(ctx context.Context, email, token string) error
	VerifyEmail(ctx context.Context, req *DTOVerifyEmailRequest) error
}

type authService struct {
	cache       cache.Cache
	db          *gorm.DB
	mailer      mailer.Mailer
	userService users.UserService
}

func NewAuthService(cache cache.Cache, db *gorm.DB, mailer mailer.Mailer, userService users.UserService) AuthService {
	return &authService{
		cache:       cache,
		db:          db,
		mailer:      mailer,
		userService: userService,
	}
}

func (s *authService) comparePassword(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return err
	}

	return nil
}

func (s *authService) GenerateAndSendVerificationToken(ctx context.Context, email string) error {
	db := s.db.WithContext(ctx)

	limitKey := fmt.Sprintf("email_verification_limit:%s", email)
	limit, err := s.cache.Get(ctx, limitKey)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return service_error.NewInternalServerError(ctx, "failed to get verification limit", err)
		}
	}
	if limit != "" {
		return service_error.NewTooManyRequests(ctx, "verification limit exceeded", nil)
	}

	user, err := s.userService.GetUserByEmail(db, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || !user.IsActive {
			return service_error.NewNotFound(ctx, "user not found", nil)
		}
		return service_error.NewInternalServerError(ctx, "failed to get user", err)
	}

	if user.IsVerified {
		return service_error.NewBadRequest(ctx, "user is already verified", nil)
	}

	token, err := s.generateVerificationToken()
	if err != nil {
		return service_error.NewInternalServerError(ctx, "failed to generate verification token", err)
	}

	cacheKey := fmt.Sprintf("email_verification:%s", token)
	if err := s.cache.Set(ctx, cacheKey, email, VERIFICATION_TOKEN_TTL); err != nil {
		return service_error.NewInternalServerError(ctx, "failed to store verification token", err)
	}

	if err := s.cache.Set(ctx, limitKey, "1", VERIFICATION_TOKEN_TTL); err != nil {
		return service_error.NewInternalServerError(ctx, "failed to store verification limit", err)
	}

	if err := s.sendEmailVerification(ctx, email, token); err != nil {
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

func (s *authService) generateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *authService) hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *authService) Login(ctx context.Context, req *DTOLoginRequest) (*DTOAuthResponse, error) {
	db := s.db.WithContext(ctx)
	loginAttemptsKey := fmt.Sprintf("login_attempts:%s", utils.GetIp(ctx))

	if attempts, _ := s.cache.Get(ctx, loginAttemptsKey); attempts != "" {
		n, _ := strconv.Atoi(attempts)
		if n >= 3 {
			return nil, service_error.NewTooManyRequests(ctx, "login attempts exceeded", nil)
		}
	}

	recordFailed := func() {
		attempts, _ := s.cache.Get(ctx, loginAttemptsKey)
		n, _ := strconv.Atoi(attempts)
		s.cache.Set(ctx, loginAttemptsKey, strconv.Itoa(n+1), 3*time.Minute)
	}

	user, err := s.userService.GetUserByEmail(db, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			recordFailed()
			return nil, service_error.NewUnauthorized(ctx, "email or password is incorrect", err)
		}
		return nil, service_error.NewInternalServerError(ctx, "failed to retrieve user data", err)
	}

	if !user.IsActive {
		recordFailed()
		return nil, service_error.NewUnauthorized(ctx, "email or password is incorrect", err)
	}

	if !user.IsVerified {
		return nil, service_error.NewUnauthorized(ctx, "email is not verified", err)
	}

	if err = s.comparePassword(req.Password, user.Password); err != nil {
		recordFailed()
		return nil, service_error.NewUnauthorized(ctx, "email or password is incorrect", err)
	}

	tokenPair, err := s.generateTokenPair(user.Id.String())
	if err != nil {
		return nil, service_error.NewInternalServerError(ctx, "failed to generate token pair", err)
	}

	return tokenPair, nil
}

func (s *authService) Register(ctx context.Context, req *DTORegisterRequest) error {
	db := s.db.WithContext(ctx)

	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return service_error.NewInternalServerError(ctx, "failed to create new user", err)
	}

	userId, err := uuid.NewV7()
	if err != nil {
		return service_error.NewInternalServerError(ctx, "failed to create new user", err)
	}

	newUser := &users.User{
		Id:       userId,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = s.userService.CreateUser(db, newUser)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return service_error.NewConflict(ctx, "user already exists", err)
		}
		return service_error.NewInternalServerError(ctx, "failed to create new user", err)
	}

	err = s.GenerateAndSendVerificationToken(ctx, req.Email)
	if err != nil {
		return service_error.NewInternalServerError(ctx, "failed to send verification email", err)
	}

	return err
}

func (s *authService) sendEmailVerification(ctx context.Context, email, token string) error {
	subject := "Verify your email address"
	verificationLink := fmt.Sprintf("%s/auth/verify?token=%s", os.Getenv("APP_HOST"), token)

	template := fmt.Sprintf(`<body style="margin:0;padding:32px 16px;background:#f4f5f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
		<table role="presentation" width="100%%" cellpadding="0" cellspacing="0"><tr><td align="center">
		<table role="presentation" cellpadding="0" cellspacing="0" style="max-width:440px;width:100%%;background:#fff;border-radius:8px">
		<tr><td style="padding:32px 24px">
		<h1 style="margin:0 0 8px;font-size:20px;color:#1a1a2e">Verify your email</h1>
		<p style="margin:0 0 20px;font-size:14px;color:#5a6170;line-height:1.5">Click the button below to confirm your email. This link expires in 5 minutes.</p>
		<table role="presentation" cellpadding="0" cellspacing="0"><tr><td>
		<a href="%s" style="display:inline-block;padding:11px 24px;background:#2563eb;color:#fff;font-size:14px;font-weight:600;text-decoration:none;border-radius:6px">Verify Email</a>
		</td></tr></table>
		</td></tr>
		<tr><td style="padding:16px 24px;border-top:1px solid #e5e7eb">
		<p style="margin:0;font-size:12px;color:#9ca3af">If you did not sign up, ignore this email.</p>
		</td></tr>
		</table>
		</td></tr></table>
		</body>`, verificationLink)

	err := s.mailer.Send(ctx, email, subject, template)
	if err != nil {
		return service_error.NewInternalServerError(ctx, "failed to send email verification", err)
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *DTOVerifyEmailRequest) error {
	db := s.db.WithContext(ctx)

	cacheKey := fmt.Sprintf("email_verification:%s", req.Token)
	email, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		return service_error.NewUnauthorized(ctx, "invalid or expired verification token", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := s.userService.VerifyUserEmail(tx, email); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return service_error.NewUnauthorized(ctx, "invalid or expired verification token", err)
			}
			return service_error.NewInternalServerError(ctx, "failed to verify user", err)
		}

		if err := s.cache.Del(ctx, cacheKey); err != nil {
			return service_error.NewInternalServerError(ctx, "failed to remove verification token", err)
		}

		return nil
	})

	return err
}
