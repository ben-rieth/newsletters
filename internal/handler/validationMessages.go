package handler

import "github.com/danielgtaylor/huma/v2/validation"

// Huma bakes these into each schema when its route is registered, so this must
// run before any handlers are added to the API.
func UseFriendlyValidationMessages() {
	validation.MsgExpectedMinLength = "must be at least %d characters"
	validation.MsgExpectedMaxLength = "must be at most %d characters"
	validation.MsgExpectedMinimumNumber = "must be at least %v"
	validation.MsgExpectedMaximumNumber = "must be at most %v"
	validation.MsgExpectedExclusiveMinimumNumber = "must be greater than %v"
	validation.MsgExpectedExclusiveMaximumNumber = "must be less than %v"
	validation.MsgExpectedBePattern = "must be %s"
	validation.MsgExpectedMatchPattern = "is not in a valid format"
	validation.MsgExpectedOneOf = "must be one of: %s"
	validation.MsgExpectedRequiredProperty = "%s is required"
}
