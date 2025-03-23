package policy

import (
	"fmt"
)

type PolicyFactory struct {
	policies map[string]Policy
	// add other policies here
}

func NewPolicyFactory() *PolicyFactory {
	policies := map[string]Policy{}

	return &PolicyFactory{
		policies: policies,
	}
}

// ForResource возвращает политику для указанного ресурса
func (f *PolicyFactory) ForResource(resource string) (Policy, error) {
	policy, exists := f.policies[resource]
	if !exists {
		return nil, fmt.Errorf("policy for resource '%s' not found", resource)
	}
	return policy, nil
}

// RegisterPolicy регистрирует новую политику для ресурса
func (f *PolicyFactory) RegisterPolicy(resource string, policy Policy) {
	f.policies[resource] = policy
}
