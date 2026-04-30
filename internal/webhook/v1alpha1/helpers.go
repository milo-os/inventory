// SPDX-License-Identifier: AGPL-3.0-only

package v1alpha1

import "fmt"

// maxChildNames is the maximum number of child names included inline in a
// deletion-rejection message. Any additional children are summarized by
// truncationSuffix.
const maxChildNames = 5

// childNames returns up to maxChildNames names extracted from items via the
// given name accessor. The generic helper keeps the webhook files free of
// per-kind boilerplate when formatting rejection messages.
func childNames[T any](items []T, name func(T) string) []string {
	limit := len(items)
	if limit > maxChildNames {
		limit = maxChildNames
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, name(items[i]))
	}
	return out
}

// truncationSuffix returns a " (and N more)" fragment when total exceeds
// maxChildNames, or the empty string otherwise. The returned fragment is
// designed to be appended directly after the child-name list.
func truncationSuffix(total int) string {
	if total <= maxChildNames {
		return ""
	}
	return fmt.Sprintf(" (and %d more)", total-maxChildNames)
}
