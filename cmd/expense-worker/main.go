package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	ctx := context.Background()
	queueURL := os.Getenv("EXPENSE_EVENTS_QUEUE_URL")
	if queueURL == "" {
		log.Fatalf("EXPENSE_EVENTS_QUEUE_URL is not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	log.Println("SQS client created")

	for {
		output, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 5,
			WaitTimeSeconds:     10,
		})
		if err != nil {
			log.Fatalf("Failed to receive message: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		for _, message := range output.Messages {
			if message.Body != nil {
				log.Printf("Received message: %s", *message.Body)
			}

			if _, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			}); err != nil {
				log.Fatalf("Failed to delete message: %v", err)
			}

		}
	}
}
