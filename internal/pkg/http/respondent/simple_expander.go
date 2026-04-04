package respondent

import (
	"errors"
	"net/http"
)

type (
	SimpleExpander struct {
		rules    []extractorRule
		fallback ManifestExtractor
	}

	extractorRule struct {
		Pattern   error
		Extractor ManifestExtractor
	}
)

func (se *SimpleExpander) Expand(err error) *Manifest {
	if errors.Is(err, nil) {
		return nil
	}

	for _, rule := range se.rules {
		if errors.Is(err, rule.Pattern) {
			manifest := rule.Extractor(err)
			if manifest != nil {
				return manifest
			}
		}
	}

	if se.fallback != nil {
		return se.fallback(err)
	}

	return nil
}

func (se *SimpleExpander) ExtractFor(sentinelErr error, extractor ManifestExtractor) *SimpleExpander {
	if !errors.Is(sentinelErr, nil) && extractor != nil {
		newRule := extractorRule{
			Extractor: extractor,
			Pattern:   sentinelErr,
		}

		se.rules = append(se.rules, newRule)
	}
	return se
}

func (se *SimpleExpander) FallbackExtractor(extractor ManifestExtractor) *SimpleExpander {
	if extractor != nil {
		se.fallback = extractor
	}

	return se
}

func NewSimpleExpander() *SimpleExpander {
	return &SimpleExpander{
		rules: make([]extractorRule, 0),
		fallback: func(err error) *Manifest {
			return &Manifest{
				Status:      http.StatusInternalServerError,
				Error:       "Internal server error",
				ErrorCode:   50001,
				ErrorDetail: err.Error(),
			}
		},
	}
}

func (se *SimpleExpander) WithoutDetail(err error, status, errorCode int, message string) *SimpleExpander {
	extractor := func(_ error) *Manifest {
		return &Manifest{
			Status:    status,
			Error:     message,
			ErrorCode: errorCode,
		}
	}

	return se.ExtractFor(err, extractor)
}

func (se *SimpleExpander) WithDetail(err error, status, errorCode int, message, detail string) *SimpleExpander {
	extractor := func(_ error) *Manifest {
		return &Manifest{
			Status:      status,
			Error:       message,
			ErrorCode:   errorCode,
			ErrorDetail: detail,
		}
	}

	return se.ExtractFor(err, extractor)
}
