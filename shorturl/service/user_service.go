package service

import (
	"context"
	"errors"
	"shorturl/config"
	"shorturl/dao"
	"shorturl/model"
	"shorturl/util"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	dao          *dao.UserDAO
	emailService *EmailService
	redisClient  *redis.Client
}

func NewUserService() *UserService {
	return &UserService{
		dao:          dao.NewUserDAO(),
		emailService: NewEmailService(),
		redisClient:  config.RedisClient,
	}
}

func (s *UserService) Register(email, password string) error {
	if !ValidateEmail(email) {
		return errors.New("invalid email format")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	existingEmail, err := s.dao.GetByEmail(email)
	if err != nil {
		return err
	}
	if existingEmail != nil {
		return errors.New("email already exists")
	}

	username := extractUsernameFromEmail(email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	return s.dao.Create(user)
}

func extractUsernameFromEmail(email string) string {
	for i, c := range email {
		if c == '@' {
			return email[:i]
		}
	}
	return email
}

func (s *UserService) Login(email, password string) (string, error) {
	if !ValidateEmail(email) {
		return "", errors.New("invalid email format")
	}

	user, err := s.dao.GetByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	return util.GenerateToken(user.ID, user.Username, config.AppConfig.JWT.Secret)
}

func (s *UserService) SendCaptcha(ctx context.Context, email string) error {
	if !ValidateEmail(email) {
		return errors.New("invalid email format")
	}

	user, err := s.dao.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("email not registered")
	}

	captcha := util.GenerateCaptcha()
	if err := util.StoreCaptcha(ctx, s.redisClient, util.CaptchaKey(email), captcha); err != nil {
		return err
	}

	return s.emailService.SendCaptcha(email, captcha)
}

func (s *UserService) LoginWithCaptcha(ctx context.Context, email, captcha string) (string, error) {
	if !ValidateEmail(email) {
		return "", errors.New("invalid email format")
	}

	valid, err := util.VerifyCaptcha(ctx, s.redisClient, util.CaptchaKey(email), captcha)
	if err != nil {
		return "", errors.New("captcha expired")
	}
	if !valid {
		return "", errors.New("invalid captcha")
	}

	user, err := s.dao.GetByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("email not registered")
	}

	return util.GenerateToken(user.ID, user.Username, config.AppConfig.JWT.Secret)
}

func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	if !ValidateEmail(email) {
		return errors.New("invalid email format")
	}

	user, err := s.dao.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("email not registered")
	}

	captcha := util.GenerateCaptcha()
	if err := util.StoreCaptcha(ctx, s.redisClient, util.CaptchaKey(email), captcha); err != nil {
		return err
	}

	return s.emailService.SendCaptcha(email, captcha)
}

func (s *UserService) ResetPassword(ctx context.Context, email, captcha, newPassword string) error {
	if !ValidateEmail(email) {
		return errors.New("invalid email format")
	}
	if len(newPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	valid, err := util.VerifyCaptcha(ctx, s.redisClient, util.CaptchaKey(email), captcha)
	if err != nil {
		return errors.New("captcha expired")
	}
	if !valid {
		return errors.New("invalid captcha")
	}

	user, err := s.dao.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("email not registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.dao.UpdatePassword(user.ID, string(hashedPassword))
}