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
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	LoadEnvFile()
}

func isProductionEnv() bool {
	env := os.Getenv("APP_ENV")
	return env == "prod" || env == "production"
}

const (
	maxWorkers        = 100
	maxMessages       = 10
	waitTimeSeconds   = 20
	visibilityTimeout = 30
)

// SQSSubscriber provides a worker pool for processing SQS messages
type SQSSubscriber struct {
	client      *sqs.Client
	queueURL    *string
	workerCount int
	handlers    []MessageHandler
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	started     bool
	mu          sync.Mutex
}

// MessageHandler defines the function signature for processing SQS messages
type MessageHandler func(msg *types.Message) error

func NewSQSSubscriber(queueName string, workerCount int) (*SQSSubscriber, error) {
	if workerCount <= 0 || workerCount > maxWorkers {
		return nil, fmt.Errorf("worker count must be between 1 and %d", maxWorkers)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	if !isProductionEnv() {
		queueName = fmt.Sprintf("%s_%s", os.Getenv("APP_ENV"), queueName)
	}

	accountID := os.Getenv("AWS_ACCOUNT_ID")
	if accountID == "" {
		cancel()
		return nil, fmt.Errorf("AWS_ACCOUNT_ID environment variable is required")
	}

	result, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueOwnerAWSAccountId: &accountID,
		QueueName:              &queueName,
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if handler == nil {
		return
	}
	s.handlers = append(s.handlers, handler)
}

func (s *SQSSubscriber) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return // Already started
	}

	if len(s.handlers) == 0 {
		log.Println("Warning: No message handlers registered")
	}

	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.startWorker()
	}
	s.started = true
}

func (s *SQSSubscriber) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	s.cancel()
	s.wg.Wait()
	s.started = false
}

func (s *SQSSubscriber) Close() error {
	s.Stop()
	// Any additional cleanup if needed
	return nil
}

func (s *SQSSubscriber) startWorker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Worker shutting down gracefully")
			return
		default:
			messages, err := s.receiveMessages()
			if err != nil {
				log.Printf("Error receiving messages: %v", err)
				select {
				case <-s.ctx.Done():
					return
				case <-time.After(time.Second):
					continue
				}
			}

			for _, msg := range messages.Messages {
				select {
				case <-s.ctx.Done():
					return
				default:
					s.processMessage(&msg)
				}
			}
		}
	}
}

func (s *SQSSubscriber) receiveMessages() (*sqs.ReceiveMessageOutput, error) {
	return s.client.ReceiveMessage(s.ctx, &sqs.ReceiveMessageInput{
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
		if msg.ReceiptHandle != nil {
			_, err := s.client.DeleteMessage(s.ctx, &sqs.DeleteMessageInput{
				QueueUrl:      s.queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
		}
	}
}
