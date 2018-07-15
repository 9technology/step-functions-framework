package my9sns

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSSession struct {
	svc *sns.SNS
}

type SNSPublishInput struct {
	Message  string
	Subject  string
	TopicArn string
}

func NewSnsSession(sess client.ConfigProvider, region string) (sn SNSSession, err error) {
	sn.svc = sns.New(sess, aws.NewConfig().WithRegion(region))
	return sn, err
}

func (sn *SNSSession) SnsPublish(publishIn SNSPublishInput) (err error) {

	params := &sns.PublishInput{
		Message:  &publishIn.Message,
		Subject:  &publishIn.Subject,
		TopicArn: &publishIn.TopicArn,
	}
	resp, err := sn.svc.Publish(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}
	// Pretty-print the response data.
	fmt.Println(resp)
	return
}
