package evaluation

import (
	"fmt"
	"math/rand"

	"github.com/muonsoft/llm-guard"
)

// GeneratorSeed is the fixed seed for deterministic generated evaluation suites.
const GeneratorSeed int64 = 202608241

const generatedSuiteID = "generated-smoke"
const generatedSourceID = "generated"
const generatedMappingVersion = "gen-v1"

// GenerateSmokeSuite returns a bounded deterministic smoke suite for offline gates.
func GenerateSmokeSuite(seed int64) []SuiteRecord {
	rng := rand.New(rand.NewSource(seed))
	var records []SuiteRecord
	appendRec := func(id, input string, annotations []SuiteAnnotation, lifecycle *SuiteLifecycle) {
		rec := buildSuiteRecord(generatedSourceID, id, generatedMappingVersion, input, annotations)
		rec.SuiteID = generatedSuiteID
		if lifecycle != nil {
			rec.Lifecycle = lifecycle
		}
		records = append(records, rec)
	}

	appendRec("inn-valid-10", "ИНН 7707083893", ann("INN", llmguard.EntityINN, 7, 17), maskLifecycle())
	appendRec("inn-valid-12", "ИНН 500100732259", ann("INN", llmguard.EntityINN, 7, 19), maskLifecycle())
	appendRec("inn-invalid-checksum", "ИНН "+mutateLastDigit("7707083893"), nil, nil)
	appendRec("inn-homogeneous", "ИНН 1111111111", nil, nil)

	appendRec("snils-compact", "СНИЛС 11223344595", ann("SNILS", llmguard.EntitySNILS, 11, 22), maskLifecycle())
	appendRec("snils-formatted", "СНИЛС 123-456-789 64", ann("SNILS", llmguard.EntitySNILS, 11, 25), maskLifecycle())
	appendRec("snils-legacy", "СНИЛС 000-000-001 00", nil, nil)

	appendRec("bank-card-valid", "карта 4111111111111111", ann("BANK_CARD", llmguard.EntityBankCard, 11, 27), maskLifecycle())
	appendRec("bank-card-invalid", "карта "+mutateLastDigit("4111111111111111"), nil, nil)

	acctInput := "расчётный счёт 40702810900000000001"
	appendRec("bank-account-valid", acctInput, ann("BANK_ACCOUNT", llmguard.EntityBankAccount, 28, 48), maskLifecycle())
	appendRec("bank-account-near-miss", "счёт 40702810900000000001", nil, nil)

	passInput := "паспорт 45 08 123456"
	appendRec("passport-valid", passInput, ann("PASSPORT", llmguard.EntityPassport, 15, 27), maskLifecycle())

	appendRec("dob-valid", "дата рождения 12.10.1990", ann("DATE_OF_BIRTH", llmguard.EntityDateOfBirth, 26, 36), maskLifecycle())
	appendRec("phone-valid", "Звоните +7 (999) 123-45-67", ann("PHONE", llmguard.EntityPhone, 15, 33), maskLifecycle())
	appendRec("email-valid", "Contact a@b.co", ann("EMAIL", llmguard.EntityEmail, 8, 14), maskLifecycle())
	appendRec("url-valid", "see https://example.com/a", ann("URL", llmguard.EntityURL, 4, 25), maskLifecycle())
	appendRec("ip-valid", "host 192.0.2.42", ann("IP_ADDRESS", llmguard.EntityIPAddress, 5, 15), maskLifecycle())
	appendRec("address-valid", "ул. Ленина, 10", ann("ADDRESS", llmguard.EntityAddress, 0, 22), maskLifecycle())
	appendRec("person-valid", "Документ подписал Иван Петров.", ann("PERSON", llmguard.EntityPerson, 34, 55), maskLifecycle())

	appendRec("jwt-valid", "token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJzeW50aGV0aWMifQ.c2ln",
		ann("SECRET_JWT", llmguard.EntitySecretJWT, 6, 58), blockLifecycle())
	pemInput := "-----BEGIN PRIVATE KEY-----\nAQIDBA==\n-----END PRIVATE KEY-----"
	appendRec("pem-valid", pemInput, ann("SECRET_PRIVATE_KEY", llmguard.EntitySecretPrivateKey, 0, len(pemInput)), blockLifecycle())
	appendRec("api-key-valid", "ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		ann("SECRET_API_KEY", llmguard.EntitySecretAPIKey, 0, 40), blockLifecycle())
	appendRec("dsn-valid", "postgres://user:pass@localhost/app",
		ann("CONNECTION_STRING", llmguard.EntityConnectionString, 0, 34), blockLifecycle())

	appendRec("unicode-boundary", "кириллица a@b.co конец", ann("EMAIL", llmguard.EntityEmail, 19, 25), maskLifecycle())

	// Seed-driven invalid/near-miss variants keep generator coverage deterministic.
	appendRec("seed-inn-fail", "ИНН "+generateINN10(rng)+"0", nil, nil)
	appendRec("seed-card", "карта "+generateLuhnCard(rng, "4111"), nil, nil)

	appendRec("mutate-recipe", "Contact a@b.co", ann("EMAIL", llmguard.EntityEmail, 8, 14),
		&SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "mutate_placeholder"})
	appendRec("delete-recipe", "Contact a@b.co", ann("EMAIL", llmguard.EntityEmail, 8, 14),
		&SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "delete_placeholder"})
	appendRec("collision-recipe", "Contact a@b.co", ann("EMAIL", llmguard.EntityEmail, 8, 14),
		&SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "collision"})

	return records
}

func ann(label string, entity llmguard.EntityType, start, end int) []SuiteAnnotation {
	return []SuiteAnnotation{{
		SourceLabel:  label,
		MappedEntity: string(entity),
		Start:        start,
		End:          end,
		Disposition:  DispositionSupported,
	}}
}

func maskLifecycle() *SuiteLifecycle {
	return &SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "identity"}
}

func blockLifecycle() *SuiteLifecycle {
	return &SuiteLifecycle{ExpectedAction: "block"}
}

func generateINN10(rng *rand.Rand) string {
	digits := make([]byte, 9)
	for i := range digits {
		digits[i] = byte('0' + rng.Intn(10))
	}
	coef := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * coef[i]
	}
	check := sum % 11 % 10
	return string(digits) + fmt.Sprintf("%d", check)
}

func generateLuhnCard(rng *rand.Rand, prefix string) string {
	digits := prefix
	for len(digits) < 15 {
		digits += fmt.Sprintf("%d", rng.Intn(10))
	}
	sum := 0
	parity := len(digits) % 2
	for i := 0; i < len(digits); i++ {
		n := int(digits[i] - '0')
		if i%2 == parity {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	check := (10 - (sum % 10)) % 10
	return digits + fmt.Sprintf("%d", check)
}

func mutateLastDigit(value string) string {
	if len(value) == 0 {
		return value
	}
	last := value[len(value)-1]
	next := byte('0')
	if last != '9' {
		next = last + 1
	}
	return value[:len(value)-1] + string(next)
}
