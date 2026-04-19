package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

func hello(ctx context.Context) error {
	log.Println("Hello, manga notifier")
	return nil
}

func main() {
	lambda.Start(hello)
}
