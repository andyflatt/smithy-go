package middleware

import (
	"context"
)

// Handler provides the interface for performing the logic to obtain an output,
// or error for the given input.
type Handler interface {
	// Handle performs logic to obtain an output for the given input. Handler
	// should be decorated with middleware to perform input specific behavior.
	Handle(ctx context.Context, input interface{}) (output interface{}, err error)
}

// HandlerFunc provides a wrapper around a function pointer to be used as a
// middleware handler.
type HandlerFunc func(ctx context.Context, input interface{}) (output interface{}, err error)

// Handle invokes the underlying function, returning the result.
func (fn HandlerFunc) Handle(ctx context.Context, input interface{}) (output interface{}, err error) {
	return fn(ctx, input)
}

// Middleware provides the interface to call handlers in a chain.
type Middleware interface {
	// ID provides a unique identifier for the middleware.
	ID() string

	// Performs the middleware's handling of the input, returning the output,
	// or error. The middleware can invoke the next Handler if handling should
	// continue.
	HandleMiddleware(ctx context.Context, input interface{}, next Handler) (
		output interface{}, err error,
	)
}

// decoratedHandler wraps a middleware in order to to call the next handler in
// the chain.
type decoratedHandler struct {
	// The next handler to be called.
	Next Handler

	// The current middleware decorating the handler.
	With Middleware
}

// Handle implements the Handler interface to handle a operation invocation.
func (m decoratedHandler) Handle(ctx context.Context, input interface{}) (
	output interface{}, err error,
) {
	ctx = RecordMiddleware(ctx, m.With.ID())
	return m.With.HandleMiddleware(ctx, input, m.Next)
}

type middlewareIDs struct{}

// RecordMiddleware appends the middleware id to the list of middleware ids to
// the context.
func RecordMiddleware(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, middlewareIDs{}, append(GetMiddlewareIDs(ctx), id))
}

// GetMiddlewareIDs returns a list of middleware IDs in the order they were
// added in.
func GetMiddlewareIDs(ctx context.Context) []string {
	v, ok := ctx.Value(middlewareIDs{}).([]string)
	if !ok {
		return nil
	}
	return v
}

// DecorateHandler decorates a handler with a middleware. Wrapping the handler
// with the middleware.
func DecorateHandler(h Handler, with ...Middleware) Handler {
	for i := len(with) - 1; i >= 0; i-- {
		h = decoratedHandler{
			Next: h,
			With: with[i],
		}
	}

	return h
}
