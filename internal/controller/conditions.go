// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// readyConditionType is the common "Ready" condition type used by every
// inventory kind. The per-kind type constants (e.g. RegionReady, SiteReady)
// all resolve to this string — we hard-code it here so that the helpers in
// this file do not need to depend on any particular kind's types package.
const readyConditionType = "Ready"

// SetReady upserts a Ready=True condition on the supplied conditions slice.
// It sets ObservedGeneration to the supplied generation and delegates
// LastTransitionTime handling to meta.SetStatusCondition.
func SetReady(generation int64, conds *[]metav1.Condition, reason, msg string) {
	meta.SetStatusCondition(conds, metav1.Condition{
		Type:               readyConditionType,
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            msg,
		ObservedGeneration: generation,
	})
}

// SetNotReady upserts a Ready=False condition on the supplied conditions
// slice, with the same semantics as SetReady.
func SetNotReady(generation int64, conds *[]metav1.Condition, reason, msg string) {
	meta.SetStatusCondition(conds, metav1.Condition{
		Type:               readyConditionType,
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            msg,
		ObservedGeneration: generation,
	})
}
