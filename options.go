package gcloudtracer

import (
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

var (
	// ErrInvalidProjectID occurs if project identifier is invalid.
	ErrInvalidProjectID = errors.New("invalid project id")
)

// Options containes options for recorder and StackDriver client.
type Options struct {
	external []option.ClientOption

	log       Logger
	projectID string
}

// Valid validates Options.
func (o *Options) Valid() error {
	if o.projectID == "" {
		return ErrInvalidProjectID
	}
	if o.log == nil {
		o.log = &defaultLogger{}
	}
	return nil
}

// Option defines an recorder option.
type Option func(o *Options)

// WithProject returns a Option that specifies a project identifier.
func WithProject(pid string) Option {
	return func(o *Options) {
		o.projectID = pid
	}
}

// WithLogger returns a Option that specifies a logger of the Recorder.
func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.log = logger
	}
}

// WithTokenSource returns a ClientOption that specifies an OAuth2 token
// source to be used as the basis for authentication.
func WithTokenSource(s oauth2.TokenSource) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithTokenSource(s))
	}
}

// WithServiceAccountFile returns a Option that uses a Google service
// account credentials file to authenticate.
// Use WithTokenSource with a token source created from
// golang.org/x/oauth2/google.JWTConfigFromJSON
// if reading the file from disk is not an option.
func WithServiceAccountFile(filename string) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithServiceAccountFile(filename))
	}
}

// WithEndpoint returns a Option that overrides the default endpoint
// to be used for a service.
func WithEndpoint(url string) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithEndpoint(url))
	}
}

// WithScopes returns a Option that overrides the default OAuth2 scopes
// to be used for a service.
func WithScopes(scope ...string) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithScopes(scope...))
	}
}

// WithUserAgent returns a Option that sets the User-Agent.
func WithUserAgent(ua string) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithUserAgent(ua))
	}
}

// WithHTTPClient returns a Option that specifies the HTTP client to use
// as the basis of communications. This option may only be used with services
// that support HTTP as their communication transport. When used, the
// WithHTTPClient option takes precedent over all other supplied options.
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithHTTPClient(client))
	}
}

// WithGRPCConn returns a Option that specifies the gRPC client
// connection to use as the basis of communications. This option many only be
// used with services that support gRPC as their communication transport. When
// used, the WithGRPCConn option takes precedent over all other supplied
// options.
func WithGRPCConn(conn *grpc.ClientConn) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithGRPCConn(conn))
	}
}

// WithGRPCDialOption returns a Option that appends a new grpc.DialOption
// to an underlying gRPC dial. It does not work with WithGRPCConn.
func WithGRPCDialOption(opt grpc.DialOption) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithGRPCDialOption(opt))
	}
}

// WithGRPCConnectionPool returns a Option that creates a pool of gRPC
// connections that requests will be balanced between.
// This is an EXPERIMENTAL API and may be changed or removed in the future.
func WithGRPCConnectionPool(size int) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithGRPCConnectionPool(size))
	}
}

// WithAPIKey returns a Option that specifies an API key to be used
// as the basis for authentication.
func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.external = append(o.external, option.WithAPIKey(apiKey))
	}
}
