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

	DefaultBackrestURL     string
	DefaultBackrestAuthRef *v1alpha1.SecretRef

	DefaultRepo v1alpha1.BackrestRepoSpec
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

	snap.DefaultBackrestURL = strings.TrimSpace(cfg.Spec.DefaultBackrest.URL)
	if cfg.Spec.DefaultBackrest.AuthRef != nil && cfg.Spec.DefaultBackrest.AuthRef.Name != "" {
		snap.DefaultBackrestAuthRef = &v1alpha1.SecretRef{Name: cfg.Spec.DefaultBackrest.AuthRef.Name}
	}

	// Copy defaults (preserving optional pointers/slices).
	snap.DefaultRepo = cfg.Spec.BindingGeneration.DefaultRepo
	return snap, nil
}
