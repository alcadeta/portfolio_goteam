package login

// ReqValidator describes a type that can be used to validate requests sent
// to the login route.
type ReqValidator interface{ Validate(ReqBody) bool }

// Validator is the ReqValidator for the login route.
type Validator struct{}

// NewValidator creates and returns a new Validator.
func NewValidator() Validator { return Validator{} }

// Validate validates the request body sent to the login route.
func (v Validator) Validate(reqBody ReqBody) bool {
	if reqBody.Username == "" || reqBody.Password == "" {
		return false
	}
	return true
}
