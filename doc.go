// Package gorion is a Go client library for Iron.io (http://www.iron.io/) services.
// It provides interfaces (see the mq.Client for example) for use in your code
// and in-memory "stubs" for those interfaces that you can use in your code's
// unit tests.
//
// Additionally, all interface funcs take a net.Context (http://godoc.org/golang.org/x/net/context)
// which your code can use to control timeouts, cancellation, and more. See https://blog.golang.org/context
// for more information on how to use Contexts.
//
// Currently supports a small subset of the IronMQ API. See the mq package
// for more usage details.
package gorion
