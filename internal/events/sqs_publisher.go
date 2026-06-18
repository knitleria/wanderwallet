package events

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSPublisher struct {
	sqsClient *sqs.Client
	queueURL  string
}

func NewSQSPublisher(context context.Context, queueURL string) (*SQSPublisher, error) {
	cfg, err := config.LoadDefaultConfig(context)
	if err != nil {
		return nil, err
	}
	return &SQSPublisher{
		sqsClient: sqs.NewFromConfig(cfg),
		queueURL:  queueURL,
	}, nil
}

func (p *SQSPublisher) PublishExpense(ctx context.Context, event ExpenseCreatedEvent) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = p.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(json)),
	})
	return err
}
