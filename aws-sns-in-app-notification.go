package pkgcommon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

// InAppNotificationSchema is the "schema" message attribute of an InAppNotification.
// A consumer checks it to know the message body is the notification document,
// before it parses anything.
const InAppNotificationSchema = "in-app-notification/v1"

// InAppNotification is an in-app notification published as one JSON document in
// the SNS message body (KX-1220). AWS treats the body as a message's payload and
// attributes as metadata for deciding how to handle it, so only routing metadata
// (schema, type, event) travels as attributes. A new field is a new JSON key,
// never an attribute — SQS delivers at most 10 of those, and SNS silently drops
// anything above that on a raw-delivery subscription (KX-1217).
//
// Producers still on SNSNotification keep working: notification-service reads
// both shapes.
type InAppNotification struct {
	Topic string `json:"-"`

	Text       string          `json:"text"`            // the card's text; the push body
	Title      string          `json:"title,omitempty"` // push title; the consumer derives one from Type when empty
	Type       string          `json:"type"`
	TypeID     string          `json:"type_id"`
	Recipients []string        `json:"recipients"`
	Data       json.RawMessage `json:"data,omitempty"` // the structured payload the card renders

	// Event marks a message that updates existing cards instead of creating one
	// (e.g. "action_status_changed", KX-1217). Such a message has no recipients.
	Event string `json:"event,omitempty"`

	// ActorID is the user who triggered the notification. Leave it empty when no
	// person did; it must never be the recipient.
	ActorID     string `json:"actor_id,omitempty"`
	CollapseKey string `json:"collapse_key,omitempty"`
	Persona     string `json:"persona,omitempty"`
	ContextType string `json:"context_type,omitempty"`
	ContextID   string `json:"context_id,omitempty"`

	ActionType       string          `json:"action_type,omitempty"`
	ActionStatus     string          `json:"action_status,omitempty"` // only on an Event message
	ActionDescriptor json.RawMessage `json:"action_descriptor,omitempty"`
	ActionExpiresAt  *time.Time      `json:"action_expires_at,omitempty"`
}

// Send validates the notification and publishes it.
func (n *InAppNotification) Send(ctx context.Context) error {
	input, err := n.build()
	if err != nil {
		return err
	}
	return PublishWithContext(ctx, input)
}

// build returns the exact PublishInput Send publishes.
func (n *InAppNotification) build() (*sns.PublishInput, error) {
	if n.Topic == "" {
		return nil, errors.New("topic is required")
	}
	if n.Type == "" || n.TypeID == "" {
		return nil, errors.New("type and type_id are required")
	}
	if n.Event == "" && n.Text == "" {
		return nil, errors.New("text is required")
	}
	if len(n.Data) > 0 && !json.Valid(n.Data) {
		return nil, errors.New("data is not valid JSON")
	}
	if len(n.ActionDescriptor) > 0 && !json.Valid(n.ActionDescriptor) {
		return nil, errors.New("action_descriptor is not valid JSON")
	}

	doc := *n
	if doc.Recipients == nil {
		doc.Recipients = []string{} // "[]", not null
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	attributes := map[string]*sns.MessageAttributeValue{
		"schema": {DataType: aws.String("String"), StringValue: aws.String(InAppNotificationSchema)},
		"type":   {DataType: aws.String("String"), StringValue: aws.String(n.Type)},
	}
	if n.Event != "" {
		attributes["event"] = &sns.MessageAttributeValue{DataType: aws.String("String"), StringValue: aws.String(n.Event)}
	}

	size := len(body)
	for k, v := range attributes {
		size += len(k) + len(*v.StringValue)
	}
	if size > maxSNSMessageSize {
		return nil, fmt.Errorf("notification exceeds maximum SNS message size of %d bytes (current: %d)", maxSNSMessageSize, size)
	}

	return &sns.PublishInput{
		TopicArn:          aws.String(GetSNSArn(n.Topic)),
		Message:           aws.String(string(body)),
		MessageAttributes: attributes,
	}, nil
}
