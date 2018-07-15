package my9ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Session struct {
	svc *ec2.EC2
}
type My9AWSSession session.Session

type EC2Instance struct {
	ImageId            string `json:"imageid"`
	InstanceType       string `json:"instancetype"`
	MinCount           int64  `json:"mincount"`
	MaxCount           int64  `json:"maxcount"`
	IamInstanceProfile string `json:"iaminstanceprofile"`
	KeyName            string `json:"keyname"`
	SubnetId           string `json:"subnetid"`
	Owner              string `json:"owner"`
	Env                string `json:"env"`
	UserData           string `json:"userdata"`
}

type EC2InstIn struct {
	InstanceId string
}

type CreateTagIn struct {
	Resource string
	Key      string
	Value    string
}

type EC2DescRoutesIn struct {
	RouteTableId string
}

func NewEc2Session(sess client.ConfigProvider, region string) (e2 EC2Session, err error) {
	e2.svc = ec2.New(sess, aws.NewConfig().WithRegion(region))
	return e2, err
}

func (e2 *EC2Session) CreateEC2Instance(ec2In EC2Instance) (runResult *ec2.Reservation, err error) {
	// Specify the details of the instance that you want to create.
	log.Println("Created instance, using role", ec2In.IamInstanceProfile)
	runResult, err = e2.svc.RunInstances(&ec2.RunInstancesInput{
		// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
		ImageId:      aws.String(ec2In.ImageId),
		InstanceType: aws.String(ec2In.InstanceType),
		MinCount:     aws.Int64(ec2In.MinCount),
		MaxCount:     aws.Int64(ec2In.MaxCount),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(ec2In.IamInstanceProfile),
		},
		KeyName:  aws.String(ec2In.KeyName),
		SubnetId: aws.String(ec2In.SubnetId),
		UserData: aws.String(ec2In.UserData),
	})

	if err != nil {
		log.Println("Could not create instance", err)
		return
	}

	log.Println("Created instance", *runResult.Instances[0].InstanceId)

	return runResult, err
}

func (e2 *EC2Session) DescribeEC2InstStatus(ec2TermIn EC2InstIn) (resp *ec2.DescribeInstanceStatusOutput, err error) {
	params := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{
			aws.String(ec2TermIn.InstanceId), // Required
			// More values...
		},
	}
	resp, err = e2.svc.DescribeInstanceStatus(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return resp, err
}

func (e2 *EC2Session) DescribeEC2Instances() (resp *ec2.DescribeInstancesOutput, err error) {
	resp, err = e2.svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return resp, err
}

func (e2 *EC2Session) CreateTag(ec2TagIn CreateTagIn) (err error) {
	params := &ec2.CreateTagsInput{
		Resources: []*string{ // Required
			aws.String(ec2TagIn.Resource), // Required
		},
		Tags: []*ec2.Tag{ // Required
			{ // Required
				Key:   aws.String(ec2TagIn.Key),
				Value: aws.String(ec2TagIn.Value),
			},
		},
	}
	_, err = e2.svc.CreateTags(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return
}

func (e2 *EC2Session) Ec2DescribeTags(instanceId string) (resp *ec2.DescribeTagsOutput, err error) {
	params := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("resource-id"),
				Values: []*string{
					aws.String(instanceId),
				},
			},
		},
	}
	resp, err = e2.svc.DescribeTags(params)

	if err != nil {
		fmt.Println(err.Error())
		return resp, err
	}
	return resp, err
}

func (e2 *EC2Session) TerminateEC2Instance(ec2TermIn EC2InstIn) (err error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{ // Required
			aws.String(ec2TermIn.InstanceId), // Required
		},
	}
	resp, err := e2.svc.TerminateInstances(params)

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

func (e2 *EC2Session) DescribeEC2RouteTables(ec2DescRoutesIn EC2DescRoutesIn) (resp *ec2.DescribeRouteTablesOutput, err error) {
	params := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []*string{
			aws.String(ec2DescRoutesIn.RouteTableId), // Required
			// More values...
		},
	}
	resp, err = e2.svc.DescribeRouteTables(params)
	if err != nil {
		fmt.Println(err.Error())
		return resp, err
	}
	return resp, err
}
