package facade

import (
	"context"
	"errors"
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
	_, err := p.userRepo.GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		p.log.Error("check username exists", zap.String("username", username), zap.Error(err))
		return model.User{}, err
	} else if err == nil {
		return model.User{}, ErrSignUpUsernameExists
	}

	// if publisher check name uniqueness
	userRole := database.UserRoleName
	if isPublisher {
		exists, uErr := p.userRepo.CheckUserExists(ctx, displayName, database.PublisherRoleName)
		if uErr != nil {
			p.log.Error("check publisher name exists", zap.String("name", displayName), zap.Error(uErr))
			return model.User{}, uErr
		} else if exists {
			return model.User{}, ErrSignUpPublisherNameExists
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
	user := database.NewUser(username, displayName, passwordHash, userRole)
	if email != "" {
		user.SetEmail(email, false)
	}

	txErr := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		if err = p.userRepo.CreateUser(ctx, user); err != nil {
			if errors.Is(err, database.ErrUserExists) {
				return ErrSignUpEmailExists
			}
			p.log.Error("create user", zap.String("username", user.Username), zap.String("email", user.Email.String), zap.Error(err))
			return err
		}

		// send verification email
		if user.Email.Valid {
			if err = p.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
				p.log.Error("send verification email on signup", zap.Error(err))
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return model.User{}, txErr
	}

	return mapDBUserToUser(user), nil
}

// SignIn authenticates user by username/email and password
func (p *Provider) SignIn(ctx context.Context, login, password string) (model.User, error) {
	var user database.User
	var err error
	// check if user exists
	// support both username (for older registrations) and email
	if strings.Contains(login, "@") {
		user, err = p.userRepo.GetUserByEmail(ctx, login)
	} else {
		user, err = p.userRepo.GetUserByUsername(ctx, login)
	}
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, ErrSignInInvalidCredentials
		}
		p.log.Error("get user by login", zap.String("login", login), zap.Error(err))
		return model.User{}, err
	}

	// check password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return model.User{}, ErrSignInInvalidCredentials
	}

	// send verification code to email if user has unverified email
	if user.Email.Valid && !user.EmailVerified {
		if err = p.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
			p.log.Error("sending verification email on signin", zap.Error(err))
			return model.User{}, err
		}
	}

	return mapDBUserToUser(user), nil
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

	username, err := extractUsernameFromEmail(email)
	if err != nil {
		p.log.Error("extract username from email", zap.String("email", email), zap.Error(err))
		return model.User{}, ErrInvalidEmail
	}

	// create user
	user = database.NewUser(username, username, nil, database.UserRoleName)
	user.SetOAuthID(auth.GoogleAuthTokenProvider, oauthID)
	user.SetEmail(email, true)

	if err = p.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, database.ErrUserExists) {
			p.log.Warn("user already exists during oauth", zap.String("username", user.Username), zap.String("email", user.Email.String))
			return model.User{}, ErrOAuthSignInConflict
		}
		p.log.Error("create user (google oauth)", zap.String("username", user.Username), zap.String("email", user.Email.String), zap.Error(err))
		return model.User{}, err
	}

	return mapDBUserToUser(user), nil
}

// UpdateUserProfile updates user profile
func (p *Provider) UpdateUserProfile(ctx context.Context, userID string, params model.UpdateProfileParams) (model.User, error) {
	var user database.User

	txErr := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		var err error

		// check if user exists
		user, err = p.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return ErrUpdateProfileUserNotFound
			}
			p.log.Error("get user by id", zap.String("userID", userID), zap.Error(err))
			return err
		}

		// update password if provided
		if params.Password != nil {
			if user.OAuthProvider.Valid {
				return ErrUpdateProfileNotAllowed
			}
			if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(*params.Password)); err != nil {
				return ErrUpdateProfileInvalidPassword
			}
			passwordHash, gErr := bcrypt.GenerateFromPassword([]byte(*params.NewPassword), bcrypt.DefaultCost)
			if gErr != nil {
				p.log.Error("generate password hash", zap.String("userID", userID), zap.Error(gErr))
				return gErr
			}
			user.PasswordHash = passwordHash
		}

		// update name if provided
		if params.Name != nil {
			user.DisplayName = *params.Name
		}

		// update user info
		if err = p.userRepo.UpdateUser(ctx, user); err != nil {
			p.log.Error("update user", zap.String("userID", userID), zap.Error(err))
			return err
		}

		return nil
	})
	if txErr != nil {
		return model.User{}, txErr
	}

	return mapDBUserToUser(user), nil
}

// DeleteUser deletes user by id
func (p *Provider) DeleteUser(ctx context.Context, userID string) error {
	return p.userRepo.DeleteUser(ctx, userID)
}

// extracts and sanitizes username from email for OAuth users
func extractUsernameFromEmail(email string) (string, error) {
	if email == "" {
		return "", ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", ErrInvalidEmail
	}

	// extract username part (before @)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", ErrInvalidEmail
	}

	username := parts[0]

	username = strings.ToLower(strings.TrimSpace(username))

	if len(username) > maxUsernameLen {
		username = username[:maxUsernameLen]
	}

	return username, nil
}
