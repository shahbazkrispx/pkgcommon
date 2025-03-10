package pkgcommon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"gorm.io/datatypes"
)

// SNSNotification represents a web SNSNotification message with the following fields:
// - Topic: The SNS topic to publish to
// - Message: The SNSNotification message content
// - Recipients: The recipients of the SNSNotification (must be JSON serializable)
// - Body: Optional additional data to include (must be JSON serializable)
// - Type: The SNSNotification type identifier
// - TypeID: A unique ID for deduplication
type SNSNotification struct {
	IsFIFO             bool   `json:"is_fifo"`
	Topic              string `json:"topic" validate:"required"`
	Message            string `json:"message" validate:"required"`
	Subject            string `json:"subject"`
	Recipients         any    `json:"recipient" validate:"required"`
	Body               any    `json:"body"`
	Type               string `json:"type"`
	TypeID             string `json:"type_id"`
	IsServiceToService bool   `json:"is_service_to_service"`
}

// maxSNSMessageSize defines the maximum size in bytes for an SNS message (256KB)
const maxSNSMessageSize = 256 * 1024

// Send publishes the notification to SNS after validating the message
// Returns an error if validation fails or the publish fails
func (w *SNSNotification) Send(ctx context.Context) error {
	if err := w.validate(); err != nil {
		return err
	}
	return PublishWithContext(ctx, w.build())
}

// parseBody serializes the Body field to JSON if present
// Returns an error if JSON serialization fails
func (w *SNSNotification) parseBody() error {
	if w.Body == nil {
		return nil
	}
	b, err := json.Marshal(w.Body)
	if err != nil {
		return err
	}
	w.Body = datatypes.JSON(b)
	return nil
}

// parseRecipients serializes the Recipients field to JSON
// Returns an error if Recipients is nil or JSON serialization fails
func (w *SNSNotification) parseRecipients() error {
	if w.Recipients == nil {
		return fmt.Errorf("at least one recipient is required")
	}
	b, err := json.Marshal(w.Recipients)
	if err != nil {
		return err
	}
	w.Recipients = datatypes.JSON(b)
	return nil
}

// build creates SNS message attributes from the notification
// Returns an SNS PublishInput with the notification data and attributes
func (w *SNSNotification) build() *sns.PublishInput {
	attributes := make(map[string]*sns.MessageAttributeValue)
	input := &sns.PublishInput{
		TopicArn: aws.String(w.getTopic()),
		Message:  aws.String(w.Message),
		Subject:  aws.String(w.Subject),
	}
	// Add non-empty attributes only
	if w.Type != "" {
		attributes["type"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(w.Type),
		}
	}

	if w.TypeID != "" {
		if w.IsFIFO {
			input.MessageDeduplicationId = aws.String(w.TypeID)
		}

		attributes["typeId"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(w.TypeID),
		}
	}

	if recipients, ok := w.Recipients.(datatypes.JSON); ok {
		attributes["recipients"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String.Array"),
			StringValue: aws.String(recipients.String()),
		}
	}

	if body, ok := w.Body.(datatypes.JSON); ok {
		attributes["body"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String.Array"),
			StringValue: aws.String(body.String()),
		}
	}
	if len(attributes) > 0 {
		input.MessageAttributes = attributes
	}

	return input
}

// validateMessageSize checks if the total message size is within SNS limits
// Returns an error if the message exceeds maxSNSMessageSize
func (w *SNSNotification) validateMessageSize() error {
	msgSize := len(w.Message) + len(w.Subject) + len(w.Body.(datatypes.JSON).String())
	if msgSize > maxSNSMessageSize {
		return fmt.Errorf("notification exceeds maximum SNS message size of %d bytes", maxSNSMessageSize)
	}
	return nil
}

// validate checks if all required fields are present and valid
// Returns an error if validation fails
func (w *SNSNotification) validate() error {
	if w.Topic == "" {
		return errors.New("topic is required")
	}
	if w.Message == "" {
		return errors.New("message is required")
	}
	if err := w.parseRecipients(); err != nil && !w.IsServiceToService {
		return err
	}
	if err := w.parseBody(); err != nil {
		return fmt.Errorf("invalid JSON body: %v", err)
	}
	if err := w.validateMessageSize(); err != nil {
		return err
	}
	return nil
}

// getTopic returns the full SNS ARN for the notification's topic
// by calling GetSNSArn with the Topic field value.
// This converts the topic name into a complete SNS topic ARN.
func (w *SNSNotification) getTopic() string {
	return GetSNSArn(w.Topic)
}
