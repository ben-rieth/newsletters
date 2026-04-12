package handler

import "github.com/danielgtaylor/huma/v2"

var internalServerError = huma.Error500InternalServerError("Something went wrong on our end. Please try again.")
var unauthorizedError = huma.Error401Unauthorized("Not authorized")