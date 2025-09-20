package web

import (
	"errors"
	"log"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate   = validator.New()
	translator *ut.UniversalTranslator
	lang       ut.Translator

	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func init() {
	enLocale := en.New()
	translator = ut.New(enLocale, enLocale)
	var found bool
	lang, found = translator.GetTranslator("en")
	if !found {
		log.Fatal("can't init validator translator")
	}
	err := entranslations.RegisterDefaultTranslations(validate, lang)
	if err != nil {
		log.Fatal("can't register validator translations")
	}

	err = validate.RegisterValidation("usernameregex", validateUsernameRegex)
	if err != nil {
		log.Fatal("can't register username validator")
	}
}

// Validate shows validation errors for each invalid field
func Validate(val interface{}) ([]FieldError, error) {
	err := validate.Struct(val)
	if err == nil {
		return nil, nil
	}
	var vErrs validator.ValidationErrors
	ok := errors.As(err, &vErrs)
	if !ok {
		return nil, err
	}

	fields := make([]FieldError, 0, len(vErrs))
	for _, vErr := range vErrs {
		field := FieldError{
			Field: vErr.Field(),
			Error: vErr.Translate(lang),
		}
		fields = append(fields, field)
	}

	return fields, err
}

func validateUsernameRegex(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}
