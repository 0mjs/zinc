package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/0mjs/zinc"
	jwtgo "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenMissing   = errors.New("zincjwt: token missing")
	ErrTokenMalformed = errors.New("zincjwt: token malformed")
	ErrTokenInvalid   = errors.New("zincjwt: token invalid")
)

type Extractor func(*zinc.Context) (string, error)

type KeyFunc func(*zinc.Context, *jwtgo.Token) (any, error)

type ParseTokenFunc func(*zinc.Context, string) (*jwtgo.Token, error)

type ValidateFunc func(*zinc.Context, *jwtgo.Token) error

type ErrorHandler func(*zinc.Context, error) error

// Config controls JWT authentication.
type Config struct {
	// Skipper skips the middleware when it returns true.
	Skipper func(*zinc.Context) bool
	// Extractor pulls a token from the request. The default reads
	// Authorization: Bearer <token>.
	Extractor Extractor
	// NewClaims allocates the claims value used during parsing.
	// If nil, jwt.MapClaims is used.
	NewClaims func(*zinc.Context) jwtgo.Claims
	// KeyFunc returns the verification key. It is required unless
	// ParseTokenFunc is provided.
	KeyFunc KeyFunc
	// ParseTokenFunc overrides the default parser. When set, KeyFunc,
	// NewClaims, and ParserOptions are ignored.
	ParseTokenFunc ParseTokenFunc
	// ParserOptions are forwarded to jwt.ParseWithClaims.
	ParserOptions []jwtgo.ParserOption
	// Validate runs after a token has been parsed and marked valid.
	Validate ValidateFunc
	// SuccessHandler runs after the token has been stored in the context.
	// If nil, the middleware continues with c.Next().
	SuccessHandler zinc.RouteHandler
	// ErrorHandler handles authentication failures.
	// If nil, the middleware responds with 401 and a Bearer challenge.
	ErrorHandler ErrorHandler
	// Realm is included in the default WWW-Authenticate challenge when set.
	Realm string
}

type contextKey int

const (
	tokenContextKey contextKey = iota
	claimsContextKey
	tokenStringContextKey
)

// DefaultConfig returns JWT middleware defaults.
func DefaultConfig() Config {
	return Config{
		Extractor: FromAuthHeader("Bearer"),
		SuccessHandler: func(c *zinc.Context) error {
			return c.Next()
		},
	}
}

// New returns JWT middleware configured by config.
func New(config Config) zinc.Middleware {
	cfg := defaultConfig(config)

	return func(c *zinc.Context) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		tokenString, err := cfg.Extractor(c)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		token, err := cfg.ParseTokenFunc(c, tokenString)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}
		if token == nil || !token.Valid {
			return cfg.ErrorHandler(c, ErrTokenInvalid)
		}

		if cfg.Validate != nil {
			if err := cfg.Validate(c, token); err != nil {
				return cfg.ErrorHandler(c, err)
			}
		}

		c.Set(tokenContextKey, token)
		c.Set(claimsContextKey, token.Claims)
		c.Set(tokenStringContextKey, tokenString)
		return cfg.SuccessHandler(c)
	}
}

// FromAuthHeader returns an extractor for Authorization: <scheme> <token>.
func FromAuthHeader(scheme string) Extractor {
	scheme = strings.TrimSpace(scheme)
	if scheme == "" {
		scheme = "Bearer"
	}
	return FromHeaderPrefix(zinc.HeaderAuthorization, scheme+" ")
}

// FromHeader returns an extractor that uses the raw value of a header.
func FromHeader(header string) Extractor {
	header = textproto.CanonicalMIMEHeaderKey(header)

	return func(c *zinc.Context) (string, error) {
		value := strings.TrimSpace(c.GetHeader(header))
		if value == "" {
			return "", fmt.Errorf("%w: %s header", ErrTokenMissing, header)
		}
		return value, nil
	}
}

// FromHeaderPrefix returns an extractor that strips prefix from a header value.
func FromHeaderPrefix(header, prefix string) Extractor {
	header = textproto.CanonicalMIMEHeaderKey(header)
	if prefix == "" {
		return FromHeader(header)
	}

	return func(c *zinc.Context) (string, error) {
		value := strings.TrimSpace(c.GetHeader(header))
		if value == "" {
			return "", fmt.Errorf("%w: %s header", ErrTokenMissing, header)
		}
		if len(value) <= len(prefix) || !strings.EqualFold(value[:len(prefix)], prefix) {
			return "", fmt.Errorf("%w: %s header", ErrTokenMalformed, header)
		}
		token := strings.TrimSpace(value[len(prefix):])
		if token == "" {
			return "", fmt.Errorf("%w: %s header", ErrTokenMalformed, header)
		}
		return token, nil
	}
}

// FromCookie returns an extractor that reads a token from a cookie.
func FromCookie(name string) Extractor {
	return func(c *zinc.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				return "", fmt.Errorf("%w: %s cookie", ErrTokenMissing, name)
			}
			return "", err
		}
		if cookie.Value == "" {
			return "", fmt.Errorf("%w: %s cookie", ErrTokenMalformed, name)
		}
		return cookie.Value, nil
	}
}

// FromQuery returns an extractor that reads a token from the query string.
func FromQuery(name string) Extractor {
	return func(c *zinc.Context) (string, error) {
		value := c.Query(name)
		if value == "" {
			return "", fmt.Errorf("%w: %s query value", ErrTokenMissing, name)
		}
		return value, nil
	}
}

// FromFirst tries extractors in order and returns the first extracted token.
// Missing-token errors fall through to the next extractor.
func FromFirst(extractors ...Extractor) Extractor {
	list := append([]Extractor(nil), extractors...)

	return func(c *zinc.Context) (string, error) {
		var lastMissing error
		for _, extractor := range list {
			if extractor == nil {
				continue
			}
			token, err := extractor(c)
			if err == nil {
				return token, nil
			}
			if errors.Is(err, ErrTokenMissing) {
				lastMissing = err
				continue
			}
			return "", err
		}
		if lastMissing != nil {
			return "", lastMissing
		}
		return "", ErrTokenMissing
	}
}

// Token returns the parsed JWT token stored by the middleware.
func Token(c *zinc.Context) (*jwtgo.Token, bool) {
	if c == nil {
		return nil, false
	}
	value, ok := c.Get(tokenContextKey)
	if !ok {
		return nil, false
	}
	token, ok := value.(*jwtgo.Token)
	return token, ok
}

// MustToken returns the parsed JWT token or panics if it is missing.
func MustToken(c *zinc.Context) *jwtgo.Token {
	token, ok := Token(c)
	if !ok {
		panic("zincjwt: token not found")
	}
	return token
}

// TokenString returns the extracted raw token string.
func TokenString(c *zinc.Context) (string, bool) {
	if c == nil {
		return "", false
	}
	value, ok := c.Get(tokenStringContextKey)
	if !ok {
		return "", false
	}
	token, ok := value.(string)
	return token, ok
}

// MustTokenString returns the raw token string or panics if it is missing.
func MustTokenString(c *zinc.Context) string {
	token, ok := TokenString(c)
	if !ok {
		panic("zincjwt: token string not found")
	}
	return token
}

// Claims returns the stored claims value with type T.
func Claims[T any](c *zinc.Context) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return zero, false
	}
	claims, ok := value.(T)
	if !ok {
		return zero, false
	}
	return claims, true
}

// MustClaims returns the stored claims value with type T or panics.
func MustClaims[T any](c *zinc.Context) T {
	claims, ok := Claims[T](c)
	if !ok {
		panic("zincjwt: claims not found")
	}
	return claims
}

func defaultConfig(config Config) Config {
	cfg := DefaultConfig()

	cfg.Skipper = config.Skipper
	if config.Extractor != nil {
		cfg.Extractor = config.Extractor
	}
	if config.Validate != nil {
		cfg.Validate = config.Validate
	}
	if config.SuccessHandler != nil {
		cfg.SuccessHandler = config.SuccessHandler
	}
	if config.Realm != "" {
		cfg.Realm = config.Realm
	}
	cfg.ParserOptions = append([]jwtgo.ParserOption(nil), config.ParserOptions...)

	if config.ParseTokenFunc != nil {
		cfg.ParseTokenFunc = config.ParseTokenFunc
	} else {
		if config.KeyFunc == nil {
			panic("zincjwt: KeyFunc or ParseTokenFunc is required")
		}
		claimsFactory := config.NewClaims
		if claimsFactory == nil {
			claimsFactory = func(*zinc.Context) jwtgo.Claims {
				return jwtgo.MapClaims{}
			}
		}
		keyFunc := config.KeyFunc
		options := append([]jwtgo.ParserOption(nil), config.ParserOptions...)
		cfg.ParseTokenFunc = func(c *zinc.Context, tokenString string) (*jwtgo.Token, error) {
			claims := claimsFactory(c)
			if claims == nil {
				claims = jwtgo.MapClaims{}
			}
			return jwtgo.ParseWithClaims(tokenString, claims, func(token *jwtgo.Token) (any, error) {
				return keyFunc(c, token)
			}, options...)
		}
	}

	if config.ErrorHandler != nil {
		cfg.ErrorHandler = config.ErrorHandler
	} else {
		realm := cfg.Realm
		cfg.ErrorHandler = func(c *zinc.Context, err error) error {
			var httpErr *zinc.HTTPError
			if errors.As(err, &httpErr) && httpErr.Code != http.StatusUnauthorized {
				return err
			}

			c.SetHeader(zinc.HeaderWWWAuthenticate, buildBearerChallenge(realm, err))

			if errors.As(err, &httpErr) {
				return err
			}
			return zinc.ErrUnauthorized
		}
	}

	return cfg
}

func buildBearerChallenge(realm string, err error) string {
	params := make([]string, 0, 2)
	if realm != "" {
		params = append(params, fmt.Sprintf(`realm=%q`, realm))
	}
	if code := challengeErrorCode(err); code != "" {
		params = append(params, fmt.Sprintf(`error=%q`, code))
	}
	if len(params) == 0 {
		return "Bearer"
	}
	return "Bearer " + strings.Join(params, ", ")
}

func challengeErrorCode(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrTokenMissing):
		return ""
	case errors.Is(err, ErrTokenMalformed):
		return "invalid_request"
	default:
		return "invalid_token"
	}
}
