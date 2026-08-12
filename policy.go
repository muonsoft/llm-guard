package llmguard

// Action selects how a resolved finding is handled during Mask.
type Action string

const (
	// ActionAllow leaves the span unchanged and does not create a token mapping.
	ActionAllow Action = "allow"
	// ActionMask replaces the span with a reversible placeholder.
	ActionMask Action = "mask"
	// ActionBlock aborts Mask with a safe checkable error and zero result.
	ActionBlock Action = "block"
)

func validateAction(action Action) error {
	switch action {
	case ActionAllow, ActionMask, ActionBlock:
		return nil
	default:
		return newInvalidConfigError("unknown action")
	}
}

func isSecretEntity(entity EntityType) bool {
	switch entity {
	case EntitySecretJWT, EntitySecretPrivateKey, EntitySecretAPIKey, EntityConnectionString:
		return true
	default:
		return false
	}
}

type secretActionOption struct {
	action Action
}

func (o secretActionOption) apply(cfg *guardConfig) error {
	if err := validateAction(o.action); err != nil {
		return err
	}
	if cfg.secretActionSet {
		return newInvalidConfigError("duplicate secret action")
	}
	cfg.secretActionSet = true
	cfg.secretAction = o.action
	return nil
}

// WithSecretAction sets the default action for all built-in secret entities.
func WithSecretAction(action Action) Option {
	return secretActionOption{action: action}
}

type entityActionOption struct {
	entity EntityType
	action Action
}

func (o entityActionOption) apply(cfg *guardConfig) error {
	if o.entity == "" {
		return newInvalidConfigError("entity is empty")
	}
	if err := validateAction(o.action); err != nil {
		return err
	}
	if _, exists := cfg.entityActions[o.entity]; exists {
		return newInvalidConfigError("duplicate entity action")
	}
	if cfg.entityActions == nil {
		cfg.entityActions = make(map[EntityType]Action)
	}
	cfg.entityActions[o.entity] = o.action
	return nil
}

// WithEntityAction overrides the action for a single entity type.
func WithEntityAction(entity EntityType, action Action) Option {
	return entityActionOption{entity: entity, action: action}
}

func copyEntityActions(src map[EntityType]Action) map[EntityType]Action {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[EntityType]Action, len(src))
	for entity, action := range src {
		dst[entity] = action
	}
	return dst
}

func (g *Guard) actionForEntity(entity EntityType) Action {
	if action, ok := g.entityActions[entity]; ok {
		return action
	}
	if isSecretEntity(entity) {
		return g.secretAction
	}
	return ActionMask
}
