package my9client

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
)

type My9AWSSession session.Session

func My9AWSNewClient() (sess *session.Session, err error) {
	sess = session.New()
	if err != nil {
		fmt.Println("Error creating AWS Session")
	}
	return sess, err
}
