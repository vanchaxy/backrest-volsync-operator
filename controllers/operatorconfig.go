package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jogotcha/backrest-volsync-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BindingGenerationPolicy string

const (
	BindingPolicyDisabled  BindingGenerationPolicy = "Disabled"
	BindingPolicyAnnotated BindingGenerationPolicy = "Annotated"
	BindingPolicyAll       BindingGenerationPolicy = "All"
)

type OperatorConfigSnapshot struct {
	Found bool

	Paused bool

	BindingPolicy BindingGenerationPolicy

	AllowedVolSyncKinds map[string]bool

	DefaultBackrestURL     string
	DefaultBackrestAuthRef *v1alpha1.SecretRef

	DefaultRepo v1alpha1.BackrestRepoSpec
}

func (s OperatorConfigSnapshot) IsVolSyncKindAllowed(kind string) bool {
	if len(s.AllowedVolSyncKinds) == 0 {
		return true
	}
	return s.AllowedVolSyncKinds[kind]
}

func LoadOperatorConfig(ctx context.Context, c client.Client, nn types.NamespacedName) (OperatorConfigSnapshot, error) {
	if nn.Name == "" || nn.Namespace == "" {
		return OperatorConfigSnapshot{Found: false}, nil
	}

	var cfg v1alpha1.BackrestVolSyncOperatorConfig
	if err := c.Get(ctx, nn, &cfg); err != nil {
		if apierrors.IsNotFound(err) {
			return OperatorConfigSnapshot{Found: false}, nil
		}
		return OperatorConfigSnapshot{}, err
	}

	snap := OperatorConfigSnapshot{Found: true}
	snap.Paused = cfg.Spec.Paused

	policy := strings.TrimSpace(cfg.Spec.BindingGeneration.Policy)
	if policy == "" {
		policy = string(BindingPolicyDisabled)
	}
	switch BindingGenerationPolicy(policy) {
	case BindingPolicyDisabled, BindingPolicyAnnotated, BindingPolicyAll:
		snap.BindingPolicy = BindingGenerationPolicy(policy)
	default:
		return OperatorConfigSnapshot{}, fmt.Errorf("invalid bindingGeneration.policy %q", policy)
	}

	// bindingGeneration.kinds defaults to allowing both VolSync kinds.
	allowedKinds := map[string]bool{
		"ReplicationSource":      false,
		"ReplicationDestination": false,
	}
	if len(cfg.Spec.BindingGeneration.Kinds) == 0 {
		allowedKinds["ReplicationSource"] = true
		allowedKinds["ReplicationDestination"] = true
	} else {
		for _, k := range cfg.Spec.BindingGeneration.Kinds {
			k = strings.TrimSpace(k)
			switch k {
			case "ReplicationSource", "ReplicationDestination":
				allowedKinds[k] = true
			default:
				return OperatorConfigSnapshot{}, fmt.Errorf("invalid bindingGeneration.kinds entry %q", k)
			}
		}
	}
	// Compact map so len==0 means "allow all" is never used; we always populate.
	snap.AllowedVolSyncKinds = allowedKinds

	snap.DefaultBackrestURL = strings.TrimSpace(cfg.Spec.DefaultBackrest.URL)
	if cfg.Spec.DefaultBackrest.AuthRef != nil && cfg.Spec.DefaultBackrest.AuthRef.Name != "" {
		snap.DefaultBackrestAuthRef = &v1alpha1.SecretRef{Name: cfg.Spec.DefaultBackrest.AuthRef.Name}
	}

	// Copy defaults (preserving optional pointers/slices).
	snap.DefaultRepo = cfg.Spec.BindingGeneration.DefaultRepo
	return snap, nil
}
