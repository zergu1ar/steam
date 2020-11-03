package steam

func validateCredentials(credentials *Credentials) error {
	if credentials.Username == "" {
		return UsernameEmptyError
	}
	if credentials.Password == "" {
		return PasswordEmptyError
	}

	return nil
}
