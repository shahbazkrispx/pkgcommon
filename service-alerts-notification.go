package pkgcommon

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

type ServiceAlert struct {
	UserID         string
	CarID          string
	FallbackDetail string
	Detail         string
	Title          string
	Data           interface{}
	TopicName      string
	Service        string
}

func ServiceAlertNotification(notification ServiceAlert) error {
	data := []byte("")
	if notification.Data != nil {
		data, _ = json.Marshal(notification.Data)
	}

	msgData := map[string]*sns.MessageAttributeValue{
		"Service": {
			DataType:    aws.String("String"),
			StringValue: aws.String(notification.Service),
		},
		"Detail": {
			DataType:    aws.String("String"),
			StringValue: aws.String(notification.Detail),
		},
		"Title": {
			DataType:    aws.String("String"),
			StringValue: aws.String(notification.Title),
		},
		"FallbackDetail": {
			DataType:    aws.String("String"),
			StringValue: aws.String(notification.FallbackDetail),
		},
		"Data": {
			DataType:    aws.String("String"),
			StringValue: aws.String(string(data)),
		},
	}

	if notification.CarID != "" {
		msgData["CarID"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(notification.CarID),
		}
	}
	if notification.UserID != "" {
		msgData["UserID"] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(notification.UserID),
		}
	}

	return PublishMessageToSNS(notification.TopicName, "Service alert", msgData)

}
