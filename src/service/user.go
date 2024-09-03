package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/NJUPT-SAST/sast-link-backend/http/request"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/NJUPT-SAST/sast-link-backend/validator"
	"github.com/pkg/errors"
	_ "github.com/redis/go-redis/v9"
)

type UserService struct {
	*BaseService
}

func NewUserService(store *BaseService) *UserService {
	return &UserService{store}
}

// CreateUserAndProfile will create a new user and profile.
//
// The password will be hash in this function.
func (s *UserService) CreateUserAndProfile(email, password string) error {
	// Split email with @
	split := regexp.MustCompile(`@`)
	uid := split.Split(email, 2)[0]
	// Lowercase the uid
	uid = strings.ToLower(uid)

	if !validator.ValidatePassword(password) {
		return errors.New("Password is invalid")
	}

	// Encrypt password
	pwdEncrypt := util.ShaHashing(password)

	if err := s.Store.CreateUserAndProfile(&store.User{
		Email:    &email,
		Password: &pwdEncrypt,
		Uid:      &uid,
	}, &store.Profile{
		Nickname: &uid,
		Email:    &email,
		OrgId:    -1,
	}); err != nil {
		return errors.Wrap(err, "Create user and profile failed")
	}

	return nil
}

// VerifyAccount handles account verification based on the operation flag
func (s *UserService) VerifyAccount(ctx context.Context, username string, flag int) (string, error) {
	switch flag {
	case 0: // Register
		log.Debugf("[%s] enter register verify\n", username)
		return s.processRegistration(ctx, username)
	case 1: // Login
		log.Debugf("[%s] enter login verify\n", username)
		return s.processLogin(ctx, username)
	case 2: // Reset Password
		log.Debugf("[%s] enter resetPWD verify\n", username)
		return s.processPasswordReset(ctx, username)
	default:
		return "", errors.New("Invalid request parameter")
	}
}

// processRegistration handles the account registration verification
func (s *UserService) processRegistration(ctx context.Context, username string) (string, error) {
	if user, err := s.VerifyAccountRegister(ctx, username); err != nil || user != nil {
		return "", errors.New("User already exists")
	}

	ticket, err := util.GenerateTokenWithExp(ctx, request.RegisterJWTSubKey(username), s.Config.Secret, request.REGISTER_TICKET_EXP)
	if err != nil {
		return "", errors.Wrap(err, "generate token failed")
	}

	if err := s.Store.Set(ctx, ticket, request.VERIFY_STATUS["VERIFY_ACCOUNT"], request.REGISTER_TICKET_EXP); err != nil {
		return "", err
	}

	return ticket, nil
}

// processLogin handles the account login verification
func (s *UserService) processLogin(ctx context.Context, username string) (string, error) {
	user, err := s.VerifyAccountLogin(ctx, username)
	if err != nil || user == nil {
		return "", err
	}

	ticket, err := util.GenerateTokenWithExp(ctx, request.LoginTicketJWTSubKey(*user.Uid), s.Config.Secret, request.LOGIN_TICKET_EXP)
	if err != nil {
		return "", err
	}

	if err := s.Store.Set(ctx, request.LoginTicketKey(*user.Uid), ticket, request.LOGIN_TICKET_EXP); err != nil {
		return "", err
	}

	return ticket, nil
}

// processPasswordReset handles the account password reset verification
func (s *UserService) processPasswordReset(ctx context.Context, username string) (string, error) {
	if user, err := s.VerifyAccountResetPWD(ctx, username); err != nil || user == nil {
		return "", err
	}

	ticket, err := util.GenerateTokenWithExp(ctx, request.ResetPwdJWTSubKey(username), s.Config.Secret, request.RESETPWD_TICKET_EXP)
	if err != nil {
		return "", err
	}

	if err := s.Store.Set(ctx, ticket, request.VERIFY_STATUS["VERIFY_ACCOUNT"], request.RESETPWD_TICKET_EXP); err != nil {
		return "", err
	}

	return ticket, nil
}

// VerifyAccountResetPWD verifies if the user's email is valid for password reset
func (s *UserService) VerifyAccountResetPWD(ctx context.Context, username string) (*store.User, error) {
	return s.verifyUserByEmail(ctx, username)
}

// VerifyAccountRegister verifies if the user's email is valid for registration
func (s *UserService) VerifyAccountRegister(ctx context.Context, username string) (*store.User, error) {
	return s.verifyUserByEmail(ctx, username)
}

// VerifyAccountLogin verifies if the user's email or UID is valid for login
func (s *UserService) VerifyAccountLogin(ctx context.Context, username string) (*store.User, error) {
	if strings.Contains(username, "@") {
		return s.verifyUserByEmail(ctx, username)
	}
	return s.Store.UserByField(ctx, "uid", username)
}

// verifyUserByEmail verifies if the user's email is valid
func (s *UserService) verifyUserByEmail(ctx context.Context, email string) (*store.User, error) {
	if !validator.ValidateEmail(email) {
		return nil, errors.New("Invalid email address")
	}

	user, err := s.Store.UserByField(ctx, "email", email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(username string, password string) (string, error) {
	// Check password
	uid, err := s.Store.CheckPassword(username, password)
	return uid, err
}

func (s *UserService) ModifyPassword(ctx context.Context, username, oldPassword, newPassword string) error {
	// Check password
	uid, err := s.Store.CheckPassword(username, oldPassword)
	if err != nil {
		return err
	}

	if uid == "" {
		return errors.New("Password is incorrect")
	}
	if err := s.Store.ChangePassword(uid, newPassword); err != nil {
		log.Errorf("Change password failed: %s", err.Error())
		return errors.Wrap(err, "Change password failed")
	}
	return nil
}

func (s *UserService) ResetPassword(username, newPassword string) error {
	// Check password form
	if !validator.ValidatePassword(newPassword) {
		return errors.New("Password is invalid")
	}

	split := regexp.MustCompile(`@`)
	uid := split.Split(username, 2)[0]
	uid = strings.ToLower(uid)

	cErr := s.Store.ChangePassword(uid, newPassword)
	if cErr != nil {
		return cErr
	}
	return nil
}

// UserInfo returns the user information of the current user
func (s *UserService) UserInfo(ctx context.Context, studentID string) (*store.User, error) {
	return s.Store.UserInfo(studentID)
}

func (s *UserService) SendEmail(ctx context.Context, username, status, title string) error {
	// Determine if the ticket is correct
	if status != request.VERIFY_STATUS["VERIFY_ACCOUNT"] {
		return errors.New("Ticket is not correct")
	}

	code := store.GenerateVerifyCode()
	s.Store.Set(ctx, request.VerifyCodeKey(username), code, request.VERIFY_CODE_EXP)
	content := store.InsertCode(code)
	if err := s.Store.SendEmail(ctx, username, content, title); err != nil {
		log.Errorf("Send email failed: %s", err.Error())
		return errors.Wrap(err, "Send email failed")
	}

	log.Debugf("Send Email to [%s] with code [%s]\n", username, code)
	return nil
}

func (s *UserService) CheckVerifyCode(ctx context.Context, status, code, flag, username string) error {
	if status != request.VERIFY_STATUS["SEND_EMAIL"] {
		return errors.New("Ticket is not correct")
	}

	target, err := s.Store.Get(ctx, request.VerifyCodeKey(username))
	if err != nil {
		log.Errorf("CheckVerifyCode error: %s", err.Error())
		return err
	}
	if target == "" {
		return errors.New("Verify code is expired")
	}

	if code != target {
		log.Errorf("Verify code is incorrect, expect [%s], got [%s]\n", target, code)
		return errors.New("Verify code is incorrect")
	}

	return nil
}

// DeleteUserAccessToken deletes the access token of the user
func (s *UserService) DeleteUserAccessToken(ctx context.Context, studentID, accessToken string) error {
	// Delete token from redis
	go s.Store.Delete(ctx, request.LoginTokenKey(studentID))

	userAccessTokens, err := s.Store.GetUserAccessTokens(ctx, studentID)
	if err != nil {
		return errors.Wrap(err, "Failed to get user access tokens")
	}
	updateAccessTokens := make([]*store.UserSetting_AccessToken, 0)
	for _, token := range userAccessTokens {
		if token.AccessToken != accessToken {
			updateAccessTokens = append(updateAccessTokens, token)
		}
	}
	updateUserSetting := store.AccessTokensUserSetting{AccessTokens: updateAccessTokens}
	if err := s.Store.UpsetUserSetting(ctx, &store.UserSetting{
		UserID: studentID,
		Key:    store.USER_SETTING_ACCESS_TOKENS,
		Value:  updateUserSetting.String(),
	}); err != nil {
		return errors.Wrap(err, "Failed to upset user setting")
	}

	return nil
}
