package encrypt

import (
	"fmt"
	"testing"
)

func TestGenPasswordHash(t *testing.T) {
	password := "admin123"
	hashed, err := GenPasswordHash([]byte(password))
	if err != nil {
		t.Fatalf("Failed to generate password hash: %v", err)
	}

	if !ValidatePasswordHash(password, string(hashed)) {
		t.Errorf("Password validation failed")
	}

	fmt.Println("Hashed password:", string(hashed))

	if ValidatePasswordHash("wrongpassword", string(hashed)) {
		t.Errorf("Password validation should have failed for wrong password")
	}
}
