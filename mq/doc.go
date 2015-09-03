// Package mq provides a Client interface for code that needs an IronMQ v3 (http://dev.iron.io/mq/3/reference/api/) client.
//
// Example usage:
//
//  import (
//    "github.com/arschles/gorion/mq"
//    "golang.org/x/net/context"
//  )
//
//  // DoEnqueue enqueues some messages and returns all the message IDs of the new
//  // messages. Returns a non-nil error on any enqueue failure
//  func DoEnqueue(client mq.Client) ([]string, error) {
//    enq, err := client.Enqueue(context.Background(), "myQueue", []NewMessage{
//      {Body: "my first message", Delay: 0, PushHeaders: make(map[string]string)},
//    })
//    if err != nil {
//      return nil, err
//    }
//    return enq.IDs, nil
//  }
package mq
