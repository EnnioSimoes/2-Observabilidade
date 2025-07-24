package address

import "testing"

func TestCheckCep(t *testing.T) {
	// Test case 1: Valid CEP
	validCep := "12345678"
	isValid, err := checkCep(validCep)
	if !isValid {
		t.Errorf("Expected valid CEP %s to return true, but got false", validCep)
	}
	if err != nil {
		t.Errorf("Expected valid CEP %s to return no error, but got %v", validCep, err)
	}

	// Test case 2: Invalid CEP (length < 8)
	shortCep := "12345"
	isValid, err = checkCep(shortCep)
	if isValid {
		t.Errorf("Expected invalid CEP %s to return false, but got true", shortCep)
	}
	if err == nil {
		t.Errorf("Expected invalid CEP %s to return an error, but got none", shortCep)
	}
	expectedError := "invalid zipcode"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}

	// Test case 3: Invalid CEP (length > 8)
	longCep := "123456789"
	isValid, err = checkCep(longCep)
	if isValid {
		t.Errorf("Expected invalid CEP %s to return false, but got true", longCep)
	}
	if err == nil {
		t.Errorf("Expected invalid CEP %s to return an error, but got none", longCep)
	}
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}
