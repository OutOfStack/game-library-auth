package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	internalErrorMsg   string = "Internal error"
	validationErrorMsg string = "Validation error"
	authErrorMsg       string = "Incorrect username or password"
)

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	Auth     *auth.Auth
	AuthConf *appconf.Auth
	UserRepo user.Repo
	Log      *zap.Logger
}

// SignInHandler - handler for sign in endpoint
func (a *AuthAPI) SignInHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.UserContext(), "handlers.signin")
	defer span.End()

	var signIn SignInReq
	if err := c.BodyParser(&signIn); err != nil {
		a.Log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.Log.With(zap.String("username", signIn.Username))

	// validate
	if fields, err := web.Validate(signIn); err != nil {
		log.Info("validating credentials", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// fetch user
	usr, err := a.UserRepo.GetByUsername(ctx, signIn.Username)
	if err != nil {
		if err == user.ErrNotFound {
			log.Info("username does not exist", zap.Error(err))
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: authErrorMsg,
			})
		}
		log.Error("fetching user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// check password
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(signIn.Password)); err != nil {
		log.Info("invalid password", zap.Error(err))
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: authErrorMsg,
		})
	}

	// get user's role
	role, err := a.UserRepo.GetRoleByID(ctx, usr.RoleID)
	if err != nil {
		log.Error("fetching role", zap.String("role", usr.RoleID.String()), zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// create claims
	claims := auth.CreateClaims(a.AuthConf.Issuer, usr, role.Name)

	// generate jwt
	tokenStr, err := a.Auth.GenerateToken(claims)
	if err != nil {
		log.Error("generating jwt", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}

// SignUpHandler - handler for sign up endpoint
func (a *AuthAPI) SignUpHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.UserContext(), "handlers.signup")
	defer span.End()

	var signUp SignUpReq
	if err := c.BodyParser(&signUp); err != nil {
		a.Log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.Log.With(zap.String("username", signUp.Name))
	// validate
	if fields, err := web.Validate(signUp); err != nil {
		log.Info("validating sign up data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// check if such username already exists
	_, err := a.UserRepo.GetByUsername(ctx, signUp.Username)
	// if err is ErrNotFound then continue
	if err != nil && err != user.ErrNotFound {
		log.Info("checking existence of user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}
	if err == nil {
		log.Info("username already exists")
		return c.Status(http.StatusConflict).JSON(web.ErrResp{
			Error: "This username is already taken",
		})
	}

	// get role id
	var roleName string
	if signUp.IsPublisher {
		// check uniqueness of publisher name
		exists, err := a.UserRepo.CheckExistPublisherWithName(ctx, signUp.Name)
		if err != nil {
			log.Error("checking existence of publisher with name", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
		if exists {
			log.Info("publisher already exists")
			return c.Status(http.StatusConflict).JSON(web.ErrResp{
				Error: "Publisher with this name already exists",
			})
		}
		roleName = user.PublisherRoleName
	} else {
		roleName = user.DefaultRoleName
	}
	role, err := a.UserRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		log.Error("fetching role", zap.String("role", roleName), zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("generating password hash", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	usr := &user.Info{
		ID:           uuid.New(),
		Username:     signUp.Username,
		Name:         signUp.Name,
		PasswordHash: passwordHash,
		RoleID:       role.ID,
		AvatarURL:    sql.NullString{String: signUp.AvatarURL, Valid: signUp.AvatarURL != ""},
	}

	// create new user
	if _, err := a.UserRepo.Create(ctx, usr); err != nil {
		log.Error("creating new user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	resp := SignUpResp{
		ID: usr.ID,
	}

	return c.JSON(resp)
}

// UpdateProfileHandler - handler for update profile endpoint
func (a *AuthAPI) UpdateProfileHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.UserContext(), "handlers.updateprofile")
	defer span.End()

	var params UpdateProfileReq
	if err := c.BodyParser(&params); err != nil {
		a.Log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.Log.With(zap.String("userid", params.UserID))

	// validate
	if fields, err := web.Validate(params); err != nil {
		log.Info("validating update profile data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	if params.Password == nil && params.NewPassword != nil {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Empty password",
		})
	}
	if params.Password != nil && params.NewPassword == nil {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Empty new password",
		})
	}
	if params.NewPassword != nil && params.ConfirmNewPassword != nil && *params.NewPassword != *params.ConfirmNewPassword {
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Confirm password does not match",
		})
	}

	// check if user exists
	usr, err := a.UserRepo.GetByID(ctx, params.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(http.StatusNotFound).JSON(web.ErrResp{
				Error: "User does not exist",
			})
		}
		log.Info("checking existence of user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	if params.Password != nil {
		// check password
		if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(*params.Password)); err != nil {
			log.Info("invalid password", zap.Error(err))
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: "Invalid current password",
			})
		}

		// hash password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*params.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Error("generating password hash", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}

		usr.PasswordHash = passwordHash
	}

	if params.Name != nil {
		usr.Name = *params.Name
	}
	if params.AvatarURL != nil {
		usr.AvatarURL = sql.NullString{String: *params.AvatarURL, Valid: *params.AvatarURL != ""}
	}

	// update user
	if _, err := a.UserRepo.Update(ctx, usr); err != nil {
		log.Error("update user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// get user's role
	role, err := a.UserRepo.GetRoleByID(ctx, usr.RoleID)
	if err != nil {
		log.Error("fetching role", zap.String("role", usr.RoleID.String()), zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// create claims
	claims := auth.CreateClaims(a.AuthConf.Issuer, usr, role.Name)

	// generate jwt
	tokenStr, err := a.Auth.GenerateToken(claims)
	if err != nil {
		log.Error("generating jwt", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// return token
	return c.JSON(TokenResp{
		AccessToken: tokenStr,
	})
}
