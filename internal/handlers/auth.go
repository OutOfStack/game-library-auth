package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/OutOfStack/game-library-auth/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultRoleName string = "user"

	internalErrorMsg   string = "Internal error"
	validationErrorMsg string = "Validation error"
	authErrorMsg       string = "Incorrect username or password"
)

// AuthAPI describes dependencies for auth endpoints
type AuthAPI struct {
	Auth     *auth.Auth
	AuthConf *appconf.Auth
	UserRepo user.Repo
	Log      *log.Logger
}

// SignUp represents data for user sign up
type SignUp struct {
	Username        string `json:"username" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
}

// SignIn represents data for user sign in
type SignIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Handler for sign in endpoint
func (aa *AuthAPI) signInHandler(c *fiber.Ctx) error {
	var signIn SignIn
	if err := c.BodyParser(&signIn); err != nil {
		aa.Log.Printf("error parsing data: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	// validate
	if fields, err := web.Validate(signIn); err != nil {
		aa.Log.Printf("validation error: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// fetch user
	usr, err := aa.UserRepo.GetByUsername(c.Context(), signIn.Username)
	if err != nil {
		if err == user.ErrNotFound {
			aa.Log.Printf("error username %s does not exist: %v\n", signIn.Username, err)
			return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
				Error: authErrorMsg,
			})
		}
		aa.Log.Printf("error fetching user %s: %v\n", signIn.Username, err)
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// check password
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(signIn.Password)); err != nil {
		aa.Log.Printf("error password does not match for user %s: %v\n", signIn.Username, err)
		return c.Status(http.StatusUnauthorized).JSON(web.ErrResp{
			Error: authErrorMsg,
		})
	}

	// get user's role
	role, err := aa.UserRepo.GetRoleByID(c.Context(), usr.RoleID)
	if err != nil {
		aa.Log.Printf("Error fetching role %v: %v\n", usr.RoleID, err)
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// create claims
	claims := auth.CreateClaims(aa.AuthConf.Issuer, usr.ID, role.Name)

	// generate jwt
	tokenStr, err := aa.Auth.GenerateToken(claims)
	if err != nil {
		aa.Log.Printf("generating jwt: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	return c.JSON(web.TokenResp{
		AccessToken: tokenStr,
	})
}

// Handler for sign up handler
func (aa *AuthAPI) signUpHandler(c *fiber.Ctx) error {
	var signUp SignUp
	if err := c.BodyParser(&signUp); err != nil {
		aa.Log.Printf("error parsing data: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	// validate
	if fields, err := web.Validate(signUp); err != nil {
		aa.Log.Printf("validation error: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// check if such username is already taken
	_, err := aa.UserRepo.GetByUsername(c.Context(), signUp.Username)
	if err != nil {
		if err != user.ErrNotFound {
			aa.Log.Printf("error checking existence of user %s: %v\n", signUp.Username, err)
			return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
				Error: internalErrorMsg,
			})
		}
	} else {
		aa.Log.Printf("error username %s already exists\n", signUp.Username)
		return c.Status(http.StatusConflict).JSON(web.ErrResp{
			Error: "This username is already taken",
		})
	}

	// get default role id
	defaultRole, err := aa.UserRepo.GetRoleByName(c.Context(), defaultRoleName)
	if err != nil {
		aa.Log.Printf("Error fetching role %s: %v\n", defaultRoleName, err)
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		aa.Log.Printf("generating password hash: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	usr := user.Info{
		ID:           uuid.New(),
		Username:     signUp.Username,
		PasswordHash: hash,
		RoleID:       defaultRole.ID,
		DateCreated:  time.Now().UTC(),
		DateUpdated:  sql.NullTime{},
	}

	// create new user
	if _, err := aa.UserRepo.Create(c.Context(), usr); err != nil {
		aa.Log.Printf("error creating new user: %v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	getUsr := user.GetUser{
		ID:          usr.ID,
		Username:    usr.Username,
		RoleID:      usr.RoleID,
		DateCreated: usr.DateCreated.String(),
		DateUpdated: types.NullTimeString(usr.DateUpdated),
	}

	return c.JSON(getUsr)
}
