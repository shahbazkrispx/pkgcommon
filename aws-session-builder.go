package pkgcommon

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
)

type Credentials struct {
	AccessKey string
	SecretKey string
	Region    string
}

func BuildSession() (*session.Session, error) {
	creds := GetCredentials()

	sessionConfig := aws.Config{
		Region:      aws.String(creds.Region),
		Credentials: credentials.NewStaticCredentials(creds.AccessKey, creds.SecretKey, ""),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: sessionConfig,
	}))
	return sess, nil
}

func GetCredentials() Credentials {
	return Credentials{
		AccessKey: os.Getenv("AWS_ACCESS_KEY"),
		SecretKey: os.Getenv("AWS_SECRET"),
		Region:    os.Getenv("AWS_REGION"),
	}
}
