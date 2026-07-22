package utils

import "errors"

var SystemError = errors.New("System failed")
var UserError = errors.New("User provided bad input")
