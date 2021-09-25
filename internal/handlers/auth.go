package handlers

import (
	"fmt"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/data/user"
	"github.com/gofiber/fiber/v2"
)

func signInHandler(c *fiber.Ctx) error {
	var signIn user.SignIn
	if err := c.BodyParser(&signIn); err != nil {
		fmt.Printf("error parsing data: %v\n", err)
		if err := c.Status(http.StatusBadRequest).JSON(ErrResp{"Error parsing data"}); err != nil {
			return fmt.Errorf("error serializing data: %w", err)
		}
		return nil
	}

	resp := "OK"

	return c.JSON(resp)
}
