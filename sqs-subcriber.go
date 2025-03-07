package pkgcommon

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go/aws"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	LoadEnvFile()
}

func isNonProductionEnv() bool {
	env := os.Getenv("APP_ENV")
	return env != "prod" && env != "production"
}

const (
	maxWorkers        = 100
	maxMessages       = 10
	waitTimeSeconds   = 20
	visibilityTimeout = 30
)

type SQSSubscriber struct {
	client      *sqs.Client
	queueURL    *string
	workerCount int
	handlers    []MessageHandler
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

type MessageHandler func(msg *types.Message) error

func NewSQSSubscriber(queueName string, workerCount int) (*SQSSubscriber, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	if isNonProductionEnv() {
		queueName = fmt.Sprintf("%s_%s", os.Getenv("APP_ENV"), queueName)
	}

	result, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueOwnerAWSAccountId: aws.String(os.Getenv("AWS_ACCOUNT_ID")),
		QueueName:              aws.String(queueName),
	})
	if err != nil {
		cancel()
		return nil, err
	}

	return &SQSSubscriber{
		client:      client,
		queueURL:    result.QueueUrl,
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (s *SQSSubscriber) AddHandler(handler MessageHandler) {
	s.handlers = append(s.handlers, handler)
}

func (s *SQSSubscriber) Start() {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.startWorker()
	}
}

func (s *SQSSubscriber) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *SQSSubscriber) startWorker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			messages, err := s.receiveMessages()
			if err != nil {
				log.Printf("Error receiving messages: %v", err)
				time.Sleep(time.Second)
				continue
			}

			for _, msg := range messages.Messages {
				s.processMessage(&msg)
			}
		}
	}
}

func (s *SQSSubscriber) receiveMessages() (*sqs.ReceiveMessageOutput, error) {
	output, err := s.client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
		MessageAttributeNames: []string{
			"All",
		},
		QueueUrl:            s.queueURL,
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTimeSeconds,
		VisibilityTimeout:   visibilityTimeout,
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *SQSSubscriber) processMessage(msg *types.Message) {
	var processError error

	// Execute all handlers for the message
	for _, handler := range s.handlers {
		if err := handler(msg); err != nil {
			processError = err
			log.Printf("Error processing message: %v", err)
			break
		}
	}

	// If message was processed successfully, delete it
	if processError == nil {
		_, err := s.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl:      s.queueURL,
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			log.Printf("Error deleting message: %v", err)
		}
	}
}
