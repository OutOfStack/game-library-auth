package web

import (
	en "github.com/go-playground/locales/en"
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
	lang, _ = translator.GetTranslator("en")
	entranslations.RegisterDefaultTranslations(validate, lang)
}

// Validate shows validation errors for each invalid field
func Validate(val interface{}) ([]FieldError, error) {
	if err := validate.Struct(val); err != nil {
		errors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, err
		}
		lang, _ := translator.GetTranslator("en")

		var fields []FieldError
		for _, verror := range errors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(lang),
			}
			fields = append(fields, field)
		}

		return fields, err
	}
	return nil, nil
}
