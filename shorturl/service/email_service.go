package service

import (
	"fmt"
	"log"
	"shorturl/config"
	"net/smtp"
	"strings"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	smtpFrom     string
	enabled      bool
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     config.AppConfig.Email.SMTPHost,
		smtpPort:     config.AppConfig.Email.SMTPPort,
		smtpUser:     config.AppConfig.Email.SMTPUser,
		smtpPassword: config.AppConfig.Email.SMTPPassword,
		smtpFrom:     config.AppConfig.Email.SMTPFrom,
		enabled:      config.AppConfig.Email.Enabled,
	}
}

func (s *EmailService) SendCaptcha(email, captcha string) error {
	if !s.enabled {
		log.Printf("email not enabled, skipping send to %s, captcha: %s", email, captcha)
		return nil
	}

	subject := "短链服务验证码"
	body := fmt.Sprintf("您的验证码是：%s\n\n验证码有效期为5分钟，请尽快使用。", captcha)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s", email, subject, body))

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	return smtp.SendMail(addr, auth, s.smtpFrom, []string{email}, msg)
}

func (s *EmailService) SendResetPasswordEmail(email, resetLink string) error {
	if !s.enabled {
		log.Printf("email not enabled, skipping send to %s, reset link: %s", email, resetLink)
		return nil
	}

	subject := "短链服务密码重置"
	body := fmt.Sprintf("请点击以下链接重置密码：\n%s\n\n如果不是您本人操作，请忽略此邮件。", resetLink)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s", email, subject, body))

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	return smtp.SendMail(addr, auth, s.smtpFrom, []string{email}, msg)
}

func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	if len(email) > 100 {
		return false
	}
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	if len(parts[0]) > 64 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	for _, c := range parts[1] {
		if c == '@' || c == ' ' {
			return false
		}
	}
	return true
}