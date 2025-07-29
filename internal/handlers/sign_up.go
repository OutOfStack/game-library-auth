package handlers

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// SignUpHandler 	godoc
// @Summary 		Register a new user
// @Description 	Create a new user account with the provided information
// @Tags  			auth
// @Accept 			json
// @Produce 		json
// @Param 			signup body SignUpReq true "User signup information"
// @Success			200 {object} SignUpResp "Successfully registered user"
// @Failure 		400 {object} web.ErrResp "Invalid input data"
// @Failure 		409 {object} web.ErrResp "Username or publisher name already exists"
// @Failure 		500 {object} web.ErrResp "Internal server error"
// @Router 			/signup [post]
func (a *AuthAPI) SignUpHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "signUp")
	defer span.End()

	var signUp SignUpReq
	if err := c.BodyParser(&signUp); err != nil {
		a.log.Error("parsing data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error: "Error parsing data",
		})
	}

	log := a.log.With(zap.String("username", signUp.Username))
	// validate
	if fields, err := web.Validate(signUp); err != nil {
		log.Info("validating sign up data", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(web.ErrResp{
			Error:  validationErrorMsg,
			Fields: fields,
		})
	}

	// check if such username already exists
	_, err := a.userRepo.GetUserByUsername(ctx, signUp.Username)
	// if err is ErrNotFound then continue
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		log.Error("checking existence of user", zap.Error(err))
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

	// get role
	var userRole = database.UserRoleName
	if signUp.IsPublisher {
		// check uniqueness of publisher name
		exists, cErr := a.userRepo.CheckUserExists(ctx, signUp.DisplayName, database.PublisherRoleName)
		if cErr != nil {
			log.Error("checking existence of publisher with name", zap.Error(cErr))
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
		userRole = database.PublisherRoleName
	}

	// hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("generating password hash", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(web.ErrResp{
			Error: internalErrorMsg,
		})
	}

	usr := database.NewUser(signUp.Username, signUp.DisplayName, passwordHash, userRole)

	// create new user
	if err = a.userRepo.CreateUser(ctx, usr); err != nil {
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
