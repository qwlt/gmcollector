package middlewares

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var (
	errMissingOrMalformedAPIKey = errors.New("missing or malformed API Key")
)

type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// SuccessHandler defines a function which is executed for a valid key.
	// Optional. Default: nil
	SuccessHandler fiber.Handler

	// ErrorHandler defines a function which is executed for an invalid key.
	// It may be used to define a custom error.
	// Optional. Default: 401 Invalid or expired key
	ErrorHandler fiber.ErrorHandler

	// Validator is a function to validate key.
	// Optional. Default: nil
	Validator func(*fiber.Ctx, string) (bool, error)

	// AuthScheme determine which http authentication scheme to use
	AuthScheme string

	// Context key to store the bearertoken from the token into context.
	// Optional. Default: "token".
	ContextKey string

	KeyExtractor func(*fiber.Ctx) (string, error)
}

// New ...
func New(config ...Config) fiber.Handler {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.SuccessHandler == nil {
		cfg.SuccessHandler = func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *fiber.Ctx, err error) error {
			if err == errMissingOrMalformedAPIKey {
				return c.Status(fiber.StatusBadRequest).SendString(err.Error())
			}
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired API Key")
		}
	}

	if cfg.AuthScheme == "" {
		cfg.AuthScheme = "Bearer"

	}

	if cfg.Validator == nil {
		cfg.Validator = func(c *fiber.Ctx, t string) (bool, error) {
			return false, nil
		}
	}
	if cfg.ContextKey == "" {
		cfg.ContextKey = "token"
	}
	if cfg.KeyExtractor == nil {
		cfg.KeyExtractor = ExtractFromHeader("Basic")
	}

	// Return middleware handler
	return func(c *fiber.Ctx) error {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}

		apiKey, err := cfg.KeyExtractor(c)
		if err != nil {
			cfg.ErrorHandler(c, err)
		}

		valid, err := cfg.Validator(c, apiKey)
		if err != nil {
			cfg.ErrorHandler(c, err)
		}
		if err == nil && valid {

			c.Locals(cfg.ContextKey, apiKey)
			return cfg.SuccessHandler(c)
		}
		return cfg.ErrorHandler(c, err)
	}
}

func ExtractFromHeader(authScheme string) func(c *fiber.Ctx) (string, error) {
	return func(c *fiber.Ctx) (string, error) {
		header := c.Get("Authorization", "")
		if header == "" {
			return "", errMissingOrMalformedAPIKey
		}
		s := strings.Split(header, " ")
		if len(s) > 1 {
			return s[1], nil
		}
		return "", errMissingOrMalformedAPIKey
	}
}

func AlwaysPassFilter(ctx *fiber.Ctx) bool {
	return true
}
