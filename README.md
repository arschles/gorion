# gorion

[![GoDoc](https://godoc.org/github.com/arschles/gorion?status.svg)](https://godoc.org/github.com/arschles/gorion)

A Go client library for [Iron.io](http://www.iron.io/) services. This library provides
interfaces that you can use in your code that have in-memory implementations that you
can use in your code's unit tests. Currently supports a small subset of the IronMQ API.

Sample usage:

```go
import (
  "github.com/arschles/gorion/mq"
  "golang.org/x/net/context"
)

// DoEnqueue enqueues some messages and returns all the message IDs of the new
// messages. Returns a non-nil error on any enqueue failure
func DoEnqueue(client mq.Client) ([]string, error) {
  enq, err := client.Enqueue(context.Background(), "myQueue", []NewMessage{
    {Body: "my first message", Delay: 0, PushHeaders: make(map[string]string)},
  })
  if err != nil {
    return nil, err
  }
  return enq.IDs, nil
}
```
