package middlewares

func RemoteValidator(token string) (bool, error) {
	return false, nil
}

// Validator which always returns true for testing purposes
func DumbValidator(token string) (bool, error) {
	return true, nil
}
