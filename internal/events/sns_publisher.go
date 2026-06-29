package events

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublisher struct {
	snsClient *sns.Client
	topicARN  string
}

func NewSNSPublisher(context context.Context, topicARN string) (*SNSPublisher, error) {
	cfg, err := config.LoadDefaultConfig(context)
	if err != nil {
		return nil, err
	}
	return &SNSPublisher{
		snsClient: sns.NewFromConfig(cfg),
		topicARN:  topicARN,
	}, nil
}

func (p *SNSPublisher) PublishExpense(ctx context.Context, event ExpenseCreatedEvent) error {
	json, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = p.snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(p.topicARN),
		Message:  aws.String(string(json)),
	})

	return err
}
