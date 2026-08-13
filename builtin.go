package llmguard

import (
	"context"

	"github.com/muonsoft/llm-guard/internal/detect"
	"github.com/muonsoft/llm-guard/internal/nlp"
)

type spanDetector struct {
	name       string
	entity     EntityType
	confidence float64
	scan       func(context.Context, string) ([]detect.Span, error)
}

func (d spanDetector) Name() string {
	return d.name
}

func (d spanDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	spans, err := d.scan(ctx, text)
	if err != nil {
		return nil, err
	}
	if len(spans) == 0 {
		return nil, nil
	}

	findings := make([]Finding, 0, len(spans))
	for _, span := range spans {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		findings = append(findings, Finding{
			Entity:     d.entity,
			Start:      span.Start,
			End:        span.End,
			Confidence: d.confidence,
			Detector:   d.name,
		})
	}
	return findings, nil
}

func nlpSpans(detectFn func(context.Context, string) ([]nlp.Span, error)) func(context.Context, string) ([]detect.Span, error) {
	return func(ctx context.Context, text string) ([]detect.Span, error) {
		src, err := detectFn(ctx, text)
		if err != nil {
			return nil, err
		}
		if len(src) == 0 {
			return nil, nil
		}
		out := make([]detect.Span, len(src))
		for i, span := range src {
			out[i] = detect.Span{Start: span.Start, End: span.End}
		}
		return out, nil
	}
}

// NewEmailDetector returns an immutable built-in EMAIL detector. Register it with
// WithDetector when constructing a Guard.
func NewEmailDetector() Detector {
	return spanDetector{name: "email", entity: EntityEmail, confidence: 0.9, scan: detect.Email}
}

// NewPhoneDetector returns an immutable built-in PHONE detector.
func NewPhoneDetector() Detector {
	return spanDetector{name: "phone", entity: EntityPhone, confidence: 0.85, scan: detect.Phone}
}

// NewURLDetector returns an immutable built-in URL detector.
func NewURLDetector() Detector {
	return spanDetector{name: "url", entity: EntityURL, confidence: 0.88, scan: detect.URL}
}

// NewIPDetector returns an immutable built-in IP_ADDRESS detector.
func NewIPDetector() Detector {
	return spanDetector{name: "ip", entity: EntityIPAddress, confidence: 0.9, scan: detect.IP}
}

// NewINNDetector returns an immutable built-in INN detector.
func NewINNDetector() Detector {
	return spanDetector{name: "inn", entity: EntityINN, confidence: 0.92, scan: detect.INN}
}

// NewSNILSDetector returns an immutable built-in SNILS detector.
func NewSNILSDetector() Detector {
	return spanDetector{name: "snils", entity: EntitySNILS, confidence: 0.91, scan: detect.SNILS}
}

// NewPassportDetector returns an immutable built-in PASSPORT detector.
func NewPassportDetector() Detector {
	return spanDetector{name: "passport", entity: EntityPassport, confidence: 0.86, scan: detect.Passport}
}

// NewBankCardDetector returns an immutable built-in BANK_CARD detector.
func NewBankCardDetector() Detector {
	return spanDetector{name: "bank_card", entity: EntityBankCard, confidence: 0.87, scan: detect.BankCard}
}

// NewBankAccountDetector returns an immutable built-in BANK_ACCOUNT detector.
func NewBankAccountDetector() Detector {
	return spanDetector{name: "bank_account", entity: EntityBankAccount, confidence: 0.85, scan: detect.BankAccount}
}

// NewDateOfBirthDetector returns an immutable built-in DATE_OF_BIRTH detector.
func NewDateOfBirthDetector() Detector {
	return spanDetector{name: "date_of_birth", entity: EntityDateOfBirth, confidence: 0.84, scan: detect.DateOfBirth}
}

// NewPersonDetector returns an immutable built-in PERSON detector for conservative
// Russian full-name and initials sequences. Register it with WithDetector when
// constructing a Guard.
func NewPersonDetector() Detector {
	return spanDetector{name: "person", entity: EntityPerson, confidence: 0.82, scan: nlpSpans(nlp.DetectPersonSpans)}
}

// NewAddressDetector returns an immutable built-in ADDRESS detector for conservative
// compositional Russian addresses. Register it with WithDetector when constructing
// a Guard.
func NewAddressDetector() Detector {
	return spanDetector{name: "address", entity: EntityAddress, confidence: 0.84, scan: nlpSpans(nlp.DetectAddressSpans)}
}

// NewJWTDetector returns an immutable built-in SECRET_JWT detector.
func NewJWTDetector() Detector {
	return spanDetector{name: "secret_jwt", entity: EntitySecretJWT, confidence: 0.92, scan: detect.JWT}
}

// NewPEMPrivateKeyDetector returns an immutable built-in SECRET_PRIVATE_KEY detector.
func NewPEMPrivateKeyDetector() Detector {
	return spanDetector{name: "secret_private_key", entity: EntitySecretPrivateKey, confidence: 0.95, scan: detect.PEMPrivateKey}
}

// NewAPIKeyDetector returns an immutable built-in SECRET_API_KEY detector.
func NewAPIKeyDetector() Detector {
	return spanDetector{name: "secret_api_key", entity: EntitySecretAPIKey, confidence: 0.9, scan: detect.APIKey}
}

// NewDSNDetector returns an immutable built-in CONNECTION_STRING detector.
func NewDSNDetector() Detector {
	return spanDetector{name: "secret_dsn", entity: EntityConnectionString, confidence: 0.93, scan: detect.DSN}
}
