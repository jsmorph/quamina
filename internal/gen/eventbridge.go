package gen

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

func EventBridgeMatches(ctx context.Context, svc *eventbridge.Client, pattern, event string) (bool, error) {
	in := &eventbridge.TestEventPatternInput{
		Event:        aws.String(event),
		EventPattern: aws.String(pattern),
	}
	out, err := svc.TestEventPattern(ctx, in)
	if err != nil {
		log.Printf("aws error %v", err)
		return false, err
	}
	return out.Result, nil
}
