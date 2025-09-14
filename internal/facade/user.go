package facade

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// SignUp creates a new user with provided params and sends verification email if applicable
func (p *Provider) SignUp(ctx context.Context, username, displayName, email, password string, isPublisher bool) (model.User, error) {
	// check if user exists
	if _, err := p.userRepo.GetUserByUsername(ctx, username); err == nil {
		return model.User{}, SignUpUsernameExistsErr
	} else if !errors.Is(err, database.ErrNotFound) {
		p.log.Error("check username exists", zap.String("username", username), zap.Error(err))
		return model.User{}, err
	}

	// if publisher check name uniqueness
	userRole := database.UserRoleName
	if isPublisher {
		exists, err := p.userRepo.CheckUserExists(ctx, displayName, database.PublisherRoleName)
		if err != nil {
			p.log.Error("check publisher name exists", zap.String("name", displayName), zap.Error(err))
			return model.User{}, err
		}
		if exists {
			return model.User{}, SignUpPublisherNameExistsErr
		}
		userRole = database.PublisherRoleName
	}

	// hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		p.log.Error("generate password hash", zap.String("username", username), zap.Error(err))
		return model.User{}, err
	}

	// create user
	usr := database.NewUser(username, displayName, passwordHash, userRole)
	if email != "" {
		usr.SetEmail(email, false)
	}

	if err = p.userRepo.CreateUser(ctx, usr); err != nil {
		p.log.Error("create user", zap.String("username", username), zap.Error(err))
		if errors.Is(err, database.ErrUsernameExists) {
			return model.User{}, SignUpUsernameExistsErr
		}
		return model.User{}, err
	}

	// send verification email
	if usr.Email.Valid {
		if err = p.sendVerificationEmail(ctx, usr.ID, usr.Email.String, usr.Username); err != nil {
			p.log.Error("send verification email on signup", zap.Error(err))
		}
	}

	return mapDBUserToUser(usr), nil
}

// SignIn authenticates user by username and password
func (p *Provider) SignIn(ctx context.Context, username, password string) (model.User, error) {
	// check if user exists
	usr, err := p.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, SignInInvalidCredentialsErr
		}
		p.log.Error("get user by username", zap.String("username", username), zap.Error(err))
		return model.User{}, err
	}

	// check password
	if err = bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return model.User{}, SignInInvalidCredentialsErr
	}

	// send verification code to email if user has unverified email
	if usr.Email.Valid && !usr.EmailVerified {
		if err = p.sendVerificationEmail(ctx, usr.ID, usr.Email.String, usr.Username); err != nil {
			p.log.Error("sending verification email on signin", zap.Error(err))
		}
	}

	return mapDBUserToUser(usr), nil
}

// GoogleOAuth handles Google OAuth sign in
func (p *Provider) GoogleOAuth(ctx context.Context, oauthID, email string) (model.User, error) {
	// check if user exists
	user, err := p.userRepo.GetUserByOAuth(ctx, auth.GoogleAuthTokenProvider, oauthID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return model.User{}, err
	}
	if err == nil {
		return mapDBUserToUser(user), nil
	}

	username, uErr := extractUsernameFromEmail(email)
	if uErr != nil {
		p.log.Error("extract username from email", zap.String("email", email), zap.Error(uErr))
		return model.User{}, InvalidEmailErr
	}

	// create user
	user = database.NewUser(username, "", nil, database.UserRoleName)
	user.SetOAuthID(auth.GoogleAuthTokenProvider, oauthID)
	user.SetEmail(email, true)

	if err = p.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, database.ErrUsernameExists) {
			p.log.Warn("username already exists during oauth", zap.String("username", user.Username))
			return model.User{}, OAutSignInConflictErr
		}
		p.log.Error("create user (google oauth)", zap.String("username", user.Username), zap.Error(err))
		return model.User{}, err
	}

	return mapDBUserToUser(user), nil
}

// UpdateUserProfile updates user profile
func (p *Provider) UpdateUserProfile(ctx context.Context, userID string, params model.UpdateProfileParams) (model.User, error) {
	// check if user exists
	usr, err := p.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, UpdateProfileUserNotFoundErr
		}
		p.log.Error("get user by id", zap.String("userID", userID), zap.Error(err))
		return model.User{}, err
	}

	// update password if provided
	if params.Password != nil {
		if usr.OAuthProvider.Valid {
			return model.User{}, UpdateProfileNotAllowedErr
		}
		if len(usr.PasswordHash) > 0 {
			if err = bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(*params.Password)); err != nil {
				return model.User{}, UpdateProfileInvalidPasswordErr
			}
		}
		passwordHash, gErr := bcrypt.GenerateFromPassword([]byte(*params.NewPassword), bcrypt.DefaultCost)
		if gErr != nil {
			p.log.Error("generate password hash", zap.String("userID", userID), zap.Error(gErr))
			return model.User{}, gErr
		}
		usr.PasswordHash = passwordHash
	}

	// update name if provided
	if params.Name != nil {
		usr.DisplayName = *params.Name
	}

	// update user info
	if err = p.userRepo.UpdateUser(ctx, usr); err != nil {
		p.log.Error("update user", zap.String("userID", userID), zap.Error(err))
		return model.User{}, err
	}

	return mapDBUserToUser(usr), nil
}

// DeleteUser deletes user by id
func (p *Provider) DeleteUser(ctx context.Context, userID string) error {
	return p.userRepo.DeleteUser(ctx, userID)
}

// extracts and sanitizes username from email for OAuth users
func extractUsernameFromEmail(email string) (string, error) {
	if email == "" {
		return "", errors.New("email cannot be empty")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	// extract username part (before @)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", errors.New("invalid email format")
	}

	username := parts[0]

	username = strings.ToLower(strings.TrimSpace(username))

	if len(username) > maxUsernameLen {
		username = username[:maxUsernameLen]
	}

	return username, nil
}
