package binding

import (
	"errors"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entr "github.com/go-playground/validator/v10/translations/en"
	rutr "github.com/go-playground/validator/v10/translations/ru"
	"github.com/rs/zerolog/log"

	"github.com/badAkne/order-service/internal/pkg/http/respondent"
)

var (
	defaultEnTranslator ut.Translator
	defaultRuTranslator ut.Translator
)

var (
	ErrMalformedSource  = errors.New("malformed request source")
	ErrValidationFailed = (*validationFailedError)(nil)
)

type validationFailedError struct {
	originalErr validator.ValidationErrors
}

func (e *validationFailedError) Error() string {
	return "Validation failed"
}

func (e *validationFailedError) Is(other error) bool {
	var err *validationFailedError

	return errors.As(other, &err)
}

func init() {
	v, _ := Validator.Engine().(*validator.Validate)

	enLocale := en.New()
	ruLocale := ru.New()

	uni := ut.New(enLocale, enLocale, ruLocale)

	var found bool
	defaultEnTranslator, found = uni.GetTranslator("en")
	if !found {
		log.Panic().Msg("EN translator not found")
	}

	if err := entr.RegisterDefaultTranslations(v, defaultEnTranslator); err != nil {
		log.Panic().Msg("Failed to register EN translations: " + err.Error())
	}

	defaultRuTranslator, found = uni.GetTranslator("ru")
	if !found {
		log.Panic().Msg("RU translator not found")
	}

	if err := rutr.RegisterDefaultTranslations(v, defaultRuTranslator); err != nil {
		log.Panic().Msg("Failed to register RU translations: " + err.Error())
	}

	registerCustomRussianTranslations(v, defaultRuTranslator)
}

func NewRespondentManifestExtractor(status, errorCode int, message string) respondent.ManifestExtractor {
	return func(err error) *respondent.Manifest {
		manifest := respondent.Manifest{
			Status:    status,
			ErrorCode: errorCode,
			Error:     message,
		}

		var errList validator.ValidationErrors
		var validationErrs validator.ValidationErrors
		var validationFailed *validationFailedError

		switch {
		case errors.As(err, &validationErrs):
			errList = validationErrs
		case errors.As(err, &validationFailed):
			errList = validationFailed.originalErr
		default:
		}

		manifest.ErrorDetails = make([]string, len(errList))

		for i := 0; i < len(errList); i++ {
			switch {
			case defaultRuTranslator != nil:
				manifest.ErrorDetails[i] = errList[i].Translate(defaultRuTranslator)
			case defaultEnTranslator != nil:
				manifest.ErrorDetails[i] = errList[i].Translate(defaultEnTranslator)
			default:
				manifest.ErrorDetails[i] = errList.Error()
			}
		}

		return &manifest
	}
}

func registerCustomRussianTranslations(v *validator.Validate, trans ut.Translator) {
	// lte - less than or equal
	err := v.RegisterTranslation(
		"lte",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("lte", "Поле {0} должно быть {1} или меньше", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("lte", fe.Field(), fe.Param())
			return t
		})
	if err != nil {
		log.Debug().Msgf("unable to register translation to russian: %s", err.Error())
	}
	// gte - greater than or equal
	err = v.RegisterTranslation(
		"gte",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("gte", "Поле {0} должно быть больше или равно {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gte", fe.Field(), fe.Param())
			return t
		})
	if err != nil {
		log.Panic().Msgf("unable to register translation to russian: %s", err.Error())
	}
	// oneof
	err = v.RegisterTranslation(
		"oneof",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("oneof", "Поле{0} должно быть одним из: {1}", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("oneof", fe.Field(), fe.Param())
			return t
		})
	if err != nil {
		log.Panic().Msgf("unable to register translation to russian: %s", err.Error())
	}

	// uuid
	err = v.RegisterTranslation(
		"uuid",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("uuid", "Поле {0} должно быть валидным UUID", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("uuid", fe.Field())
			return t
		},
	)
	if err != nil {
		log.Panic().Msgf("unable to register translation to russian: %s", err.Error())
	}

	// email
	err = v.RegisterTranslation(
		"email",
		trans,
		func(ut ut.Translator) error {
			return ut.Add("email", "Поле {0} должно быть корректным email адресом", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("email", fe.Field())
			return t
		},
	)
	if err != nil {
		log.Panic().Msgf("unable to register translation to russian: %s", err.Error())
	}
}
