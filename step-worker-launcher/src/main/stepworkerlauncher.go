package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"my9awsgo/my9client"
	"my9awsgo/my9ec2"
	"my9awsgo/my9ecs"
	"my9awsgo/my9s3"
	"my9awsgo/my9sfn"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StepWorkerLauncherConfig struct {
	Project          string `json:"project"`
	Env              string `json:"env"`
	Region           string `json:"region"`
	ConfigPrefix     string `json:"configprefix"`
	Ec2InstConfig    string `json:"ec2instanceconfig"`
	EcsContConfig    string `json:"containerconfig"`
	WorkerConfig     string `json:"workerconfig"`
	LswRetryCount    int    `json:"lswretrycount"`
	AwsAccountId     string `json:"accountid"`
	StateMachineArn  string `json:"stateMachineArn"`
	StateMachineName string `json:"stateMachineName"`
	ExecutionArn     string `json:"executionarn"`
	EcsTaskConfig    string `json:"ecstaskconfig"`
	EcsTaskCluster   string `json:"ecstaskcluster"`
}

type StateMachineInput struct {
	Project         string `json:"project"`
	Env             string `json:"env"`
	Region          string `json:"region"`
	Mode            string `json:"mode"`
	ExecutionArn    string `json:"executionarn"`
	ConfigBucket    string `json:"configbucket"`
	ConfigBucketKey string `json:"configbucketkey"`
	Result          string `json:"result"`
}

type SwlRun struct {
	S3Session                 my9s3.S3Session
	EC2Session                my9ec2.EC2Session
	SFNSession                my9sfn.SFNSession
	EcsSession                my9ecs.ECSSession
	Mode                      string
	ConfigBucket              string
	ConfigBucketKey           string
	SwlConf                   StepWorkerLauncherConfig
	SmInput                   StateMachineInput
	Ec2InstanceConfigFilePath string
	EcsTaskConfigFilePath     string
	Result                    string
	NewEc2InstanceId          string
	//SwlDateTime           SwlRunDateTime
}

func (swlRun *SwlRun) StepWorkerLauncherRun() (err error) {

	switch swlRun.Mode {
	case "run_state_machine":
		swlRun.RunStateMachine()
	case "launch_stepworker":
		swlRun.RunStepWorker()
	default:
		fmt.Println("Invalid mode for Step Worker Launcher run ...  ")
	}

	return err
}

func main() {

	project := os.Args[1]
	env := os.Args[2]
	region := os.Args[3]
	mode := os.Args[4]
	configbucket := os.Args[5]
	configbucketkey := os.Args[6]
	executionarn := os.Args[7]
	result := os.Args[8]

	var swlRun SwlRun
	swlRun.Mode = mode
	swlRun.Result = result
	swlRun.SwlConf.ConfigPrefix = configbucketkey
	swlRun.ConfigBucket = configbucket
	swlRun.ConfigBucketKey = configbucketkey + "/" + project + "/" + env + "/master.json"
	/*keyname
	Create AWS Service Sessions
	*/

	sess, err := my9client.My9AWSNewClient()
	if err != nil {
		fmt.Println("Error creating AWS Session")
	}

	ecs_session, err := my9ecs.NewEcsSession(sess, region)
	if err != nil {
		fmt.Println("Error creating ECS Session")
	}

	s3_session, err := my9s3.NewS3Session(sess, region)
	if err != nil {
		fmt.Println("Error creating S3 Session")
	}

	ec2_session, err := my9ec2.NewEc2Session(sess, region)
	if err != nil {
		fmt.Println("Error creating EC2 Session")
	}

	sfn_session, err := my9sfn.NewSfnSession(sess, region)
	if err != nil {
		fmt.Println("Error creating SFN Session")
	}

	swlRun.S3Session = s3_session
	swlRun.EC2Session = ec2_session
	swlRun.SFNSession = sfn_session
	swlRun.EcsSession = ecs_session

	err = readConfig(&swlRun)
	fmt.Println("Reading SWL config from S3 ")

	//ecsTaskConfigFile := swlRun.SwlConf.EcsTaskConfig + "/" + project + "/" + env + ".json"

	swlRun.SwlConf.Project = project
	swlRun.SwlConf.Env = env
	swlRun.SwlConf.Region = region
	swlRun.SwlConf.ExecutionArn = executionarn

	err = swlRun.StepWorkerLauncherRun()
	if err != nil {
		fmt.Println("Error during StepWorkerLauncherRun")
	}

	fmt.Println("End of StepWorkerLauncherRun ... ! ")

}

func readConfig(swlRun *SwlRun) (err error) {
	confData, err := readConfFile(swlRun.S3Session, swlRun.ConfigBucket, swlRun.ConfigBucketKey)

	err = json.Unmarshal(confData, &swlRun.SwlConf)
	if err != nil {
		fmt.Println("Error in reading APN conf file :: Error=%s", err)
		return err
	}
	return err
}

func readConfFile(s3_session my9s3.S3Session, bucketname string, bucketkey string) (content []byte, err error) {
	confFile, err := s3_session.Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(bucketkey),
	})
	if err != nil {
		fmt.Println("Error in obtaining S3 config file object :: Error=%s", err)
		return content, err
	}
	defer confFile.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, confFile.Body)
	if err != nil {
		fmt.Println("Error in reading S3 config file :: Error=%s", err)
		return content, err
	}
	content = buf.Bytes()
	return content, err
}
