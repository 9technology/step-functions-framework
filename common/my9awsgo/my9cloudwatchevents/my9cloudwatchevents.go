package my9cloudwatchevents

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

type CwePutRuleIn struct {
	EventPattern       string
	Name               string
	RoleArn            string
	ScheduleExpression string
	State              string
}

type CwePutTargetsIn struct {
	RuleName string
	Targets  []*cloudwatchevents.Target
}

type CloudWatchEventsSession struct {
	Svc *cloudwatchevents.CloudWatchEvents
}

func NewCloudWatchEventsSession(sess client.ConfigProvider, region string) (cwe CloudWatchEventsSession, err error) {
	cwe.Svc = cloudwatchevents.New(sess, aws.NewConfig().WithRegion(region))
	return cwe, err
}

func (cwe *CloudWatchEventsSession) CwePutRule(putRuleIn CwePutRuleIn) (putRuleOut *cloudwatchevents.PutRuleOutput, err error) {
	params := &cloudwatchevents.PutRuleInput{
		EventPattern:       &putRuleIn.EventPattern, // Required
		Name:               &putRuleIn.Name,
		RoleArn:            &putRuleIn.RoleArn,
		ScheduleExpression: &putRuleIn.ScheduleExpression,
		State:              &putRuleIn.State,
	}

	putRuleOut, err = cwe.Svc.PutRule(params)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(putRuleOut)
	return putRuleOut, err

}

func (cwe *CloudWatchEventsSession) CwePutTargets(putTargetsIn CwePutTargetsIn) (putTargetsOut *cloudwatchevents.PutTargetsOutput, err error) {
	params := &cloudwatchevents.PutTargetsInput{
		Rule:    &putTargetsIn.RuleName,
		Targets: putTargetsIn.Targets,
	}

	putTargetsOut, err = cwe.Svc.PutTargets(params)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(putTargetsOut)
	return putTargetsOut, err

}
