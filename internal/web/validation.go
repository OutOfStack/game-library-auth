package web

import (
	"errors"
	"log"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate   = validator.New()
	translator *ut.UniversalTranslator
	lang       ut.Translator
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

	var fields []FieldError
	for _, vErr := range vErrs {
		field := FieldError{
			Field: vErr.Field(),
			Error: vErr.Translate(lang),
		}
		fields = append(fields, field)
	}

	return fields, err
}
