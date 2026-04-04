package respondent

import (
	"github.com/badAkne/order-service/internal/pkg/http/httph"
)

type SimpleApplicator struct{}

func (sa *SimpleApplicator) Apply(ctx any, manifest *Manifest) {
	type ManifestJSON struct {
		Status       int      `json:"-"`
		Error        string   `json:"error"`
		ErrorID      string   `json:"error_id,omitempty"`
		ErrorCode    int      `json:"error_code"`
		ErrorDetail  string   `json:"error_detail,omitempty"`
		ErrorDetails []string `json:"error_details,omitempty"`
	}

	httpCtx, ok := ctx.(HttpContext)
	if !ok {
		return
	}
	w := httpCtx.W

	manifestJSON := (*ManifestJSON)(manifest)

	httph.SendJSON(w, manifest.Status, manifestJSON)
}

func NewSimpleApplicator() *SimpleApplicator {
	return &SimpleApplicator{}
}
