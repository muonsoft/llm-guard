package evaluation

import "github.com/muonsoft/llm-guard"

// SchemaVersion is the supported evaluation corpus schema version.
const SchemaVersion = 1

var mvpEntityOrder = [...]llmguard.EntityType{
	llmguard.EntityPerson,
	llmguard.EntityAddress,
	llmguard.EntityEmail,
	llmguard.EntityPhone,
	llmguard.EntityIPAddress,
	llmguard.EntityURL,
	llmguard.EntityINN,
	llmguard.EntitySNILS,
	llmguard.EntityPassport,
	llmguard.EntityBankCard,
	llmguard.EntityBankAccount,
	llmguard.EntityDateOfBirth,
	llmguard.EntitySecretJWT,
	llmguard.EntitySecretPrivateKey,
	llmguard.EntitySecretAPIKey,
	llmguard.EntityConnectionString,
}

func mvpEntities() []llmguard.EntityType {
	out := make([]llmguard.EntityType, len(mvpEntityOrder))
	copy(out, mvpEntityOrder[:])
	return out
}

// MVPEntityCount returns the number of built-in MVP entities in evaluation reports.
func MVPEntityCount() int {
	return len(mvpEntityOrder)
}
