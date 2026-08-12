package llmguard

// EntityType identifies the kind of sensitive data a finding refers to.
// Custom entity values are allowed alongside the built-in constants.
type EntityType string

// Built-in entity type constants for common PII and secret categories.
const (
	EntityPerson      EntityType = "PERSON"
	EntityAddress     EntityType = "ADDRESS"
	EntityEmail       EntityType = "EMAIL"
	EntityPhone       EntityType = "PHONE"
	EntityIPAddress   EntityType = "IP_ADDRESS"
	EntityURL         EntityType = "URL"
	EntityINN         EntityType = "INN"
	EntitySNILS       EntityType = "SNILS"
	EntityPassport    EntityType = "PASSPORT"
	EntityBankCard    EntityType = "BANK_CARD"
	EntityBankAccount EntityType = "BANK_ACCOUNT"
	EntityDateOfBirth EntityType = "DATE_OF_BIRTH"

	EntitySecretJWT        EntityType = "SECRET_JWT"
	EntitySecretPrivateKey EntityType = "SECRET_PRIVATE_KEY"
	EntitySecretAPIKey     EntityType = "SECRET_API_KEY"
	EntityConnectionString EntityType = "CONNECTION_STRING"

	// EntityCustom is the stable observability bucket label for non-built-in entity
	// types in production-safe observer events.
	EntityCustom EntityType = "CUSTOM"
)

func isBuiltinEntity(entity EntityType) bool {
	switch entity {
	case EntityPerson, EntityAddress, EntityEmail, EntityPhone, EntityIPAddress, EntityURL,
		EntityINN, EntitySNILS, EntityPassport, EntityBankCard, EntityBankAccount, EntityDateOfBirth,
		EntitySecretJWT, EntitySecretPrivateKey, EntitySecretAPIKey, EntityConnectionString:
		return true
	default:
		return false
	}
}

func observabilityEntity(entity EntityType) EntityType {
	if isBuiltinEntity(entity) {
		return entity
	}
	return EntityCustom
}
