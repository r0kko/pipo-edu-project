package service

import "testing"

func TestValidatePlate(t *testing.T) {
	valid := []string{"A123BC77", "M777MM199", "Х123ОР77"}
	invalid := []string{"123ABC77", "A12BC77", "A123B777", ""}

	for _, plate := range valid {
		if err := ValidatePlate(plate); err != nil {
			t.Fatalf("expected valid plate %s, got %v", plate, err)
		}
	}

	for _, plate := range invalid {
		if err := ValidatePlate(plate); err == nil {
			t.Fatalf("expected invalid plate %s", plate)
		}
	}
}

func TestValidateRole(t *testing.T) {
	valid := []string{"admin", "guard", "resident"}
	invalid := []string{"", "owner", "user"}

	for _, role := range valid {
		if err := ValidateRole(role); err != nil {
			t.Fatalf("expected valid role %s, got %v", role, err)
		}
	}

	for _, role := range invalid {
		if err := ValidateRole(role); err == nil {
			t.Fatalf("expected invalid role %s", role)
		}
	}
}
