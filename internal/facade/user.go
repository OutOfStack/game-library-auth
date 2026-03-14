package facade

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// errors
var (
	ErrInvalidEmail                 = errors.New("invalid email")
	ErrOAuthSignInConflict          = errors.New("oauth sign in name conflict")
	ErrUpdateProfileUserNotFound    = errors.New("update profile: user not found")
	ErrUpdateProfileInvalidPassword = errors.New("update profile: invalid current password")
	ErrUpdateProfileNotAllowed      = errors.New("update profile: password change not allowed for oauth users")
	ErrSignInInvalidCredentials     = errors.New("sign in: invalid credentials")
	ErrSignUpUsernameExists         = errors.New("sign up: username already exists")
	ErrSignUpEmailExists            = errors.New("sign up: email already exists")
	ErrSignUpEmailRequired          = errors.New("sign up: email is required")
	ErrSignUpPublisherNameExists    = errors.New("sign up: publisher name already exists")
	ErrOAuthPublisherConflict       = errors.New("oauth sign in: publisher must use password")
	ErrOAuthEmailUnverified         = errors.New("oauth sign in: email not verified by provider")
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

	// if publisher, check name uniqueness
	userRole := model.UserRoleName
	if isPublisher {
		userRole = model.PublisherRoleName

		// email is required for publishers
		if email == "" {
			return model.User{}, ErrSignUpEmailRequired
		}

		// check if publisher name already exists in user database
		companyExists, cErr := p.userRepo.CheckUserExists(ctx, displayName, model.PublisherRoleName)
		if cErr != nil {
			p.log.Error("check publisher name exists", zap.String("name", displayName), zap.Error(cErr))
			return model.User{}, cErr
		} else if companyExists {
			return model.User{}, ErrSignUpPublisherNameExists
		}

		// check if publisher is one of well-known companies
		companyExists, cErr = p.infoAPIClient.CompanyExists(ctx, displayName)
		if cErr != nil {
			p.log.Error("check company exists in game library", zap.String("name", displayName), zap.Error(cErr))
			return model.User{}, cErr
		}
		if companyExists {
			p.log.Warn("attempt to use well-known company name", zap.String("name", displayName))
			return model.User{}, ErrSignUpPublisherNameExists
		}
	}

	// hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		p.log.Error("generate password hash", zap.String("username", username), zap.Error(err))
		return model.User{}, err
	}

	// create user
	user := database.NewUser(username, displayName, passwordHash, userRole)
	if isPublisher {
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

		// send verification email only for publishers
		if isPublisher {
			if err = p.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
				switch {
				case errors.Is(err, ErrSendVerifyEmailUnsubscribed):
					// ignore 'user unsubscribed' error as emails are unique and this situation should not happen
					p.log.Warn("user is unsubscribed", zap.String("email", user.Email.String))
				case AsTooManyRequestsError(err) != nil:
					// ignore 'too many requests' error as emails are unique and this situation should not happen
					p.log.Warn("vrf code is requested too many times", zap.String("email", user.Email.String))
				default:
					p.log.Error("send verification email on signup", zap.Error(err))
					return err
				}
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
func (p *Provider) SignIn(ctx context.Context, username, password string) (model.User, error) {
	// check if user exists
	user, err := p.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.User{}, ErrSignInInvalidCredentials
		}
		p.log.Error("get user by username", zap.String("username", username), zap.Error(err))
		return model.User{}, err
	}

	// check password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return model.User{}, ErrSignInInvalidCredentials
	}

	// send verification code to email if publisher has unverified email
	if !user.EmailVerified && user.Role == model.PublisherRoleName {
		if err = p.sendVerificationEmail(ctx, user.ID, user.Email.String, user.Username); err != nil {
			switch {
			case errors.Is(err, ErrSendVerifyEmailUnsubscribed):
				// ignore 'user unsubscribed' error as such error will be displayed on resend email attempt
				p.log.Warn("user is unsubscribed", zap.String("email", user.Email.String))
			case AsTooManyRequestsError(err) != nil:
				// ignore 'too many requests' error as user should already have vrf code sent to them
				p.log.Warn("vrf code is requested too many times", zap.String("email", user.Email.String))
			default:
				p.log.Error("sending verification email on sign in", zap.Error(err))
				return model.User{}, err
			}
		}
	}

	return mapDBUserToUser(user), nil
}

// GoogleOAuth handles Google OAuth sign in
func (p *Provider) GoogleOAuth(ctx context.Context, oauthID, email string, emailVerified bool) (model.User, error) {
	// check if google oauth user exists
	user, err := p.userRepo.GetUserByOAuthLink(ctx, model.GoogleAuthTokenProvider, oauthID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return model.User{}, err
	}
	if err == nil {
		return mapDBUserToUser(user), nil
	}

	// validate email and extract username
	username, err := extractUsernameFromEmail(email)
	if err != nil {
		p.log.Error("extract username from email", zap.String("email", email), zap.Error(err))
		return model.User{}, err
	}

	// require a verified email
	if !emailVerified {
		return model.User{}, ErrOAuthEmailUnverified
	}

	// check if user with same email already exists
	if user, err = p.findExistingUserByEmail(ctx, email); err != nil {
		return model.User{}, err
	} else if !user.IsEmpty() {
		// store the OAuth link for the existing user
		link := database.NewUserOAuthLink(user.ID, model.GoogleAuthTokenProvider, oauthID)
		if lErr := p.userRepo.CreateUserOAuthLink(ctx, link); lErr != nil {
			p.log.Error("create oauth link (google, existing user)", zap.String("userID", user.ID), zap.Error(lErr))
			return model.User{}, lErr
		}
		return mapDBUserToUser(user), nil
	}

	// create user and oauth link in a transaction
	newUser := database.NewUser(username, username, nil, model.UserRoleName)
	newUser.SetEmail(email, true)

	txErr := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		if err = p.userRepo.CreateUser(ctx, newUser); err != nil {
			if errors.Is(err, database.ErrUserExists) {
				p.log.Warn("user already exists during oauth", zap.String("username", newUser.Username), zap.String("email", newUser.Email.String))
				return ErrOAuthSignInConflict
			}
			p.log.Error("create user (google oauth)", zap.String("username", newUser.Username), zap.String("email", newUser.Email.String), zap.Error(err))
			return err
		}

		link := database.NewUserOAuthLink(newUser.ID, model.GoogleAuthTokenProvider, oauthID)
		if err = p.userRepo.CreateUserOAuthLink(ctx, link); err != nil {
			p.log.Error("create oauth link (google, new user)", zap.String("userID", newUser.ID), zap.Error(err))
			return err
		}

		return nil
	})
	if txErr != nil {
		return model.User{}, txErr
	}

	return mapDBUserToUser(newUser), nil
}

// GitHubOAuth handles GitHub OAuth sign in
func (p *Provider) GitHubOAuth(ctx context.Context, oauthID, email, username string, emailVerified bool) (model.User, error) {
	// check if github oauth user exists
	user, err := p.userRepo.GetUserByOAuthLink(ctx, model.GitHubAuthTokenProvider, oauthID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return model.User{}, err
	}
	if err == nil {
		return mapDBUserToUser(user), nil
	}

	// validate email address
	_, err = extractUsernameFromEmail(email)
	if err != nil {
		return model.User{}, err
	}

	// require a verified email
	if !emailVerified {
		return model.User{}, ErrOAuthEmailUnverified
	}

	// check if user with same email already exists
	if user, err = p.findExistingUserByEmail(ctx, email); err != nil {
		return model.User{}, err
	} else if !user.IsEmpty() {
		// store the OAuth link for the existing user
		link := database.NewUserOAuthLink(user.ID, model.GitHubAuthTokenProvider, oauthID)
		if lErr := p.userRepo.CreateUserOAuthLink(ctx, link); lErr != nil {
			p.log.Error("create oauth link (github, existing user)", zap.String("userID", user.ID), zap.Error(lErr))
			return model.User{}, lErr
		}
		return mapDBUserToUser(user), nil
	}

	// truncate username to max length
	if len(username) > maxUsernameLen {
		username = username[:maxUsernameLen]
	}
	username = strings.ToLower(username)

	// create user and oauth link in a transaction
	newUser := database.NewUser(username, username, nil, model.UserRoleName)
	newUser.SetEmail(email, true)

	txErr := p.userRepo.RunWithTx(ctx, func(ctx context.Context) error {
		if err = p.userRepo.CreateUser(ctx, newUser); err != nil {
			if errors.Is(err, database.ErrUserExists) {
				p.log.Warn("user already exists during oauth", zap.String("username", newUser.Username), zap.String("email", newUser.Email.String))
				return ErrOAuthSignInConflict
			}
			p.log.Error("create user (github oauth)", zap.String("username", newUser.Username), zap.String("email", newUser.Email.String), zap.Error(err))
			return err
		}

		link := database.NewUserOAuthLink(newUser.ID, model.GitHubAuthTokenProvider, oauthID)
		if err = p.userRepo.CreateUserOAuthLink(ctx, link); err != nil {
			p.log.Error("create oauth link (github, new user)", zap.String("userID", newUser.ID), zap.Error(err))
			return err
		}

		return nil
	})
	if txErr != nil {
		return model.User{}, txErr
	}

	return mapDBUserToUser(newUser), nil
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
			hasOAuth, oErr := p.userRepo.HasOAuthLink(ctx, userID)
			if oErr != nil {
				p.log.Error("check oauth link", zap.String("userID", userID), zap.Error(oErr))
				return oErr
			}
			if hasOAuth {
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
		if params.Password != nil {
			err = p.userRepo.DeleteRefreshTokensByUserID(ctx, userID)
			if err != nil {
				p.log.Error("delete refresh tokens", zap.String("userID", userID), zap.Error(err))
				return err
			}
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

// checks if a user with the given email already exists and returns them if they are a regular user
func (p *Provider) findExistingUserByEmail(ctx context.Context, email string) (database.User, error) {
	user, err := p.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return database.User{}, nil
		}
		p.log.Error("get user by email (oauth)", zap.String("email", email), zap.Error(err))
		return database.User{}, err
	}
	if user.Role == model.PublisherRoleName {
		return database.User{}, ErrOAuthPublisherConflict
	}
	return user, nil
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
