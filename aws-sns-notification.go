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
	IsFIFO                 bool                                  `json:"is_fifo"`
	Topic                  string                                `json:"topic" validate:"required"`
	Message                string                                `json:"message" validate:"required"`
	Subject                string                                `json:"subject"`
	Recipients             any                                   `json:"recipients" validate:"required"`
	Body                   any                                   `json:"body"`
	Type                   string                                `json:"type"`
	TypeID                 string                                `json:"type_id"`
	MessageGroupID         string                                `json:"message_group_id"`
	IsServiceToService     bool                                  `json:"is_service_to_service"`
	ExtraMessageAttributes map[string]*sns.MessageAttributeValue `json:"extra_message_attributes"`
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
	}

	if w.Subject != "" {
		input.Subject = aws.String(w.Subject)
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
			if w.MessageGroupID != "" {
				input.MessageGroupId = aws.String(w.MessageGroupID)
			}
		}

		attributes["typeId"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(w.TypeID),
		}
	}

	if recipients, ok := w.Recipients.(datatypes.JSON); ok {
		attributes["recipients"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(recipients.String()),
		}
	}

	if body, ok := w.Body.(datatypes.JSON); ok {
		attributes["body"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(body.String()),
		}
	}

	// Add extra message attributes before setting input.MessageAttributes
	if w.ExtraMessageAttributes != nil {
		for k, v := range w.ExtraMessageAttributes {
			attributes[k] = v
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
	var bodySize, recipientsSize int

	if w.Body != nil {
		if body, ok := w.Body.(datatypes.JSON); ok {
			bodySize = len(body.String())
		}
	}

	if w.Recipients != nil {
		if recipients, ok := w.Recipients.(datatypes.JSON); ok {
			recipientsSize = len(recipients.String())
		}
	}

	// Include all message attribute sizes
	attributeSize := len(w.Type) + len(w.TypeID) + bodySize + recipientsSize
	if w.ExtraMessageAttributes != nil {
		for k, v := range w.ExtraMessageAttributes {
			attributeSize += len(k)
			if v.StringValue != nil {
				attributeSize += len(*v.StringValue)
			}
		}
	}

	msgSize := len(w.Message) + len(w.Subject) + attributeSize
	if msgSize > maxSNSMessageSize {
		return fmt.Errorf("notification exceeds maximum SNS message size of %d bytes (current: %d)", maxSNSMessageSize, msgSize)
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
	if w.IsFIFO && w.TypeID == "" && w.MessageGroupID == "" {
		return errors.New("FIFO topics require either TypeID or MessageGroupID")
	}
	if !w.IsServiceToService {
		if err := w.parseRecipients(); err != nil {
			return err
		}
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
