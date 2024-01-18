package ocrcommon

import (
	"errors"

	"github.com/pelletier/go-toml"
	pkgerrors "github.com/pkg/errors"
)

// CloneSet returns a copy of the input map.
func CloneSet(in map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// ValidateExplicitlySetKeys checks if the values in expected are present and the values in notExpected are not present
// in the toml tree. Works on top level keys only.
func ValidateExplicitlySetKeys(tree *toml.Tree, expected map[string]struct{}, notExpected map[string]struct{}, peerType string) error {
	var err error
	// top level keys only
	for _, k := range tree.Keys() {
		if _, ok := notExpected[k]; ok {
			err = errors.Join(err, pkgerrors.Errorf("unrecognised key for %s peer: %s", peerType, k))
		}
		delete(expected, k)
	}
	for missing := range expected {
		err = errors.Join(err, pkgerrors.Errorf("missing required key %s", missing))
	}
	return err
}
