package steam

import "errors"

var (
	UsernameEmptyError                    = errors.New("username is empty")
	PasswordEmptyError                    = errors.New("password is empty")
	InvalidCredentialsError               = errors.New("invalid username or password")
	RequireTwoFactorError                 = errors.New("require two-factor auth")
	InvalidSessionError                   = errors.New("invalid session")
	ApiKeyNotFoundError                   = errors.New("api key not found")
	ApiAccessDeniedError                  = errors.New("access denied to steam web api")
	ConfirmationsNotFoundError            = errors.New("can't find confirmation")
	ConfirmationsDescriptionNotFoundError = errors.New("can't find confirmation description")
)
