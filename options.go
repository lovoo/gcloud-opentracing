package gcloudtracer

import (
	"errors"

	"google.golang.org/api/option"
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

// WithLogger returns an Option that specifies a logger of the Recorder.
func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.log = logger
	}
}

// WithClientOption retuns an option that specifies GRPC client Options.
func WithClientOption(opts ...option.ClientOption) Option {
	return func(o *Options) {
		o.external = append(o.external, opts...)
	}
}
