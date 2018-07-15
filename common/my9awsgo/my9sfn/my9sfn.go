package my9sfn

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sfn"
)

type SFNSession struct {
	Svc *sfn.SFN
}

type CreateActIn struct {
	Name string
}

type CreateSmIn struct {
	Definition string
	Name       string
	RoleArn    string
}

type UpdateSmIn struct {
	Definition      string
	RoleArn         string
	StateMachineArn string
}

type RunSmIn struct {
	Input           string
	Name            string
	StateMachineArn string
}

type GetExecHistIn struct {
	ExecutionArn string
	NextToken    string
	ReverseOrder bool
}

func NewSfnSession(sess client.ConfigProvider, region string) (sf SFNSession, err error) {
	sf.Svc = sfn.New(sess, aws.NewConfig().WithRegion(region))
	return sf, err
}

func (sf *SFNSession) GetActivityTaskToken(activity_arn string) (token string, err error) {

	params := &sfn.GetActivityTaskInput{
		ActivityArn: aws.String(activity_arn), // Required
	}

	for i := 0; i < 10; i++ {
		resp, err := sf.Svc.GetActivityTask(params)
		if err != nil {
			fmt.Println(err.Error())
			continue
		} else if resp.TaskToken != nil {
			fmt.Println("Obtained task ...")
			token = *resp.TaskToken
			return token, err
		}
		time.Sleep(2 * 1000 * time.Millisecond)
	}
	return token, err
}

func (sf *SFNSession) SfnSendTaskSuccess(mesg string, token string) (err error) {
	params := &sfn.SendTaskSuccessInput{
		Output:    aws.String(mesg), // Required
		TaskToken: &token,           // Required
	}
	_, err = sf.Svc.SendTaskSuccess(params)
	return err
}

func (sf *SFNSession) SfnSendTaskFailure(mesg string, token string) (err error) {
	params := &sfn.SendTaskFailureInput{
		Cause:     aws.String(mesg), // Required
		TaskToken: &token,           // Required
	}
	_, err = sf.Svc.SendTaskFailure(params)
	return err
}

func (sf *SFNSession) SfnCreateActivity(createActIn CreateActIn) (createActOut *sfn.CreateActivityOutput, err error) {
	params := &sfn.CreateActivityInput{
		Name: &createActIn.Name, // Required
	}
	createActOut, err = sf.Svc.CreateActivity(params)
	return createActOut, err
}

func (sf *SFNSession) SfnDescribeActivity(activityArn string) (descActOut *sfn.DescribeActivityOutput, err error) {
	params := &sfn.DescribeActivityInput{
		ActivityArn: &activityArn, // Required
	}
	descActOut, err = sf.Svc.DescribeActivity(params)
	return descActOut, err
}

func (sf *SFNSession) SfnCreateStateMachine(createSmIn CreateSmIn) (createSmOut *sfn.CreateStateMachineOutput, err error) {
	params := &sfn.CreateStateMachineInput{
		Definition: &createSmIn.Definition,
		Name:       &createSmIn.Name,
		RoleArn:    &createSmIn.RoleArn,
	}
	createSmOut, err = sf.Svc.CreateStateMachine(params)
	return createSmOut, err
}

func (sf *SFNSession) SfnUpdateStateMachine(updateSmIn UpdateSmIn) (updateSmOut *sfn.UpdateStateMachineOutput, err error) {
	params := &sfn.UpdateStateMachineInput{
		Definition:      &updateSmIn.Definition,
		RoleArn:         &updateSmIn.RoleArn,
		StateMachineArn: &updateSmIn.StateMachineArn,
	}
	updateSmOut, err = sf.Svc.UpdateStateMachine(params)
	return updateSmOut, err
}

func (sf *SFNSession) SfnRunStateMachine(runSmIn RunSmIn) (runSmOut *sfn.StartExecutionOutput, err error) {
	params := &sfn.StartExecutionInput{
		Input:           &runSmIn.Input,
		Name:            &runSmIn.Name,
		StateMachineArn: &runSmIn.StateMachineArn,
	}
	runSmOut, err = sf.Svc.StartExecution(params)
	return runSmOut, err
}

func (sf *SFNSession) SfnGetExecutionHistory(getExecHistIn GetExecHistIn) (getExecHistOut *sfn.GetExecutionHistoryOutput, err error) {
	params := &sfn.GetExecutionHistoryInput{
		ExecutionArn: &getExecHistIn.ExecutionArn,
		//NextToken:    &getExecHistIn.NextToken,
		ReverseOrder: &getExecHistIn.ReverseOrder,
	}
	getExecHistOut, err = sf.Svc.GetExecutionHistory(params)
	return getExecHistOut, err
}
