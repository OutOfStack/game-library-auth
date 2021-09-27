package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/OutOfStack/game-library-auth/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultRoleName string = "user"

	internalErrorMsg string = "Internal error"
	authErrorMsg     string = "Incorrect username or password"
)

type AuthAPI struct {
	UserRepo user.Repo
	Log      *log.Logger
}

func (aa *AuthAPI) signInHandler(c *fiber.Ctx) error {
	var signIn user.SignIn
	if err := c.BodyParser(&signIn); err != nil {
		aa.Log.Printf("error parsing data: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(ErrResp{"Error parsing data"})
	}

	// fetch user
	usr, err := aa.UserRepo.GetByUsername(c.Context(), signIn.Username)
	if err != nil {
		if err == user.ErrNotFound {
			aa.Log.Printf("error username %s does not exist: %v\n", signIn.Username, err)
			return c.Status(http.StatusUnauthorized).JSON(ErrResp{authErrorMsg})
		}
		aa.Log.Printf("error fetching user %s: %v\n", signIn.Username, err)
		return c.Status(http.StatusInternalServerError).JSON(ErrResp{internalErrorMsg})
	}

	// check password
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(signIn.Password)); err != nil {
		aa.Log.Printf("error password does not match for user %s: %v\n", signIn.Username, err)
		return c.Status(http.StatusUnauthorized).JSON(ErrResp{authErrorMsg})
	}

	return c.JSON(TokenResp{"This is going to be be token"})
}

func (aa *AuthAPI) signUpHandler(c *fiber.Ctx) error {
	var signUp user.SignUp
	if err := c.BodyParser(&signUp); err != nil {
		aa.Log.Printf("error parsing data: %v\n", err)
		return c.Status(http.StatusBadRequest).JSON(ErrResp{"Error parsing data"})
	}

	// check if such username is already taken
	_, err := aa.UserRepo.GetByUsername(c.Context(), signUp.Username)
	if err != nil {
		if err != user.ErrNotFound {
			aa.Log.Printf("error checking existence of user %s: %v\n", signUp.Username, err)
			return c.Status(http.StatusInternalServerError).JSON(ErrResp{internalErrorMsg})
		}
	} else {
		aa.Log.Printf("error username %s already exists\n", signUp.Username)
		return c.Status(http.StatusConflict).JSON(ErrResp{"This username is already taken"})
	}

	// get default role id
	defaultRole, err := aa.UserRepo.GetRoleByName(c.Context(), defaultRoleName)
	if err != nil {
		aa.Log.Printf("Error fetching default role: %v\n", err)
		return c.Status(http.StatusInternalServerError).JSON(ErrResp{internalErrorMsg})
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		aa.Log.Printf("generating password hash: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(ErrResp{internalErrorMsg})
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
		return c.Status(http.StatusInternalServerError).JSON(ErrResp{internalErrorMsg})
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
