package facade

import (
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
)

func mapDBUserToUser(user database.User) model.User {
	return model.User{
		ID:            user.ID,
		Username:      user.Username,
		DisplayName:   user.DisplayName,
		Email:         user.Email.String,
		EmailVerified: user.EmailVerified,
		Role:          string(user.Role),
		OAuthProvider: user.OAuthProvider.String,
		OAuthID:       user.OAuthID.String,
	}
}
