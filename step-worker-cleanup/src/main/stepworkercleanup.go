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
	"my9awsgo/my9sns"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StepWorkerCleanupConfig struct {
	Project           string   `json:"project"`
	Env               string   `json:"env"`
	Region            string   `json:"region"`
	ConfigPrefix      string   `json:"configprefix"`
	Ec2InstConfig     string   `json:"ec2instanceconfig"`
	EcsContConfig     string   `json:"containerconfig"`
	WorkerConfig      string   `json:"workerconfig"`
	AwsAccountId      string   `json:"accountid"`
	StateMachineArn   string   `json:"stateMachineArn"`
	StateMachineName  string   `json:"stateMachineName"`
	ExecutionArn      string   `json:"executionarn"`
	EcsTaskConfig     string   `json:"ecstaskconfig"`
	EcsTaskCluster    string   `json:"ecstaskcluster"`
	EcsSfnClusterList []string `json:"ecssfnclusterlist"`
	SnsTopicArn       string   `json:"snstopicarn"`
}

type StateMachineInput struct {
	Project         string `json:"project"`
	Env             string `json:"env"`
	Region          string `json:"region"`
	Mode            string `json:"mode"`
	ExecutionArn    string `json:"executionarn"`
	ConfigBucket    string `json:"configbucket"`
	ConfigBucketKey string `json:"configbucketkey"`
}

type SwlRun struct {
	S3Session                 my9s3.S3Session
	EC2Session                my9ec2.EC2Session
	SFNSession                my9sfn.SFNSession
	EcsSession                my9ecs.ECSSession
	SnsSession                my9sns.SNSSession
	Mode                      string
	ConfigBucket              string
	ConfigBucketKey           string
	SwlConf                   StepWorkerCleanupConfig
	SmInput                   StateMachineInput
	Ec2InstanceConfigFilePath string
	EcsTaskConfigFilePath     string
	Result                    string
	//SwlDateTime           SwlRunDateTime
}

const CONF_PREFIX = "projects"

func (swlRun *SwlRun) StepWorkerCleanupRun() (err error) {

	switch swlRun.Mode {
	case "cleanup_state_machine":
		swlRun.CleanupStateMachine()
	case "periodic_cleanup": // periodic cleanup every hour
		for _, cluster := range swlRun.SwlConf.EcsSfnClusterList {
			fmt.Println("StepWorkerCleanupRun: Cleaning up cluster :", cluster)
			swlRun.SwlConf.EcsTaskCluster = cluster
			swlRun.CleanupStateMachine()
		}
	case "cleanup_used_and_new": // once every 6 hours cleanup both used and new
		for _, cluster := range swlRun.SwlConf.EcsSfnClusterList {
			fmt.Println("StepWorkerCleanupRun: Cleaning up cluster :", cluster)
			swlRun.SwlConf.EcsTaskCluster = cluster
			swlRun.CleanupStateMachine()
		}
	default:
		fmt.Println("No mode supplied for Step Worker Cleanup run ... Will still run cleanup ...  ")
		swlRun.CleanupStateMachine()
	}

	return err
}

var def_project, def_env, def_region, def_mode, def_confbucket, def_confbucketkey, def_result string

func main() {

	var project, env, region, mode, configbucket, configbucketkey, executionarn, result string
	region = os.Args[3]
	fmt.Println("Main : Len OS args => ", len(os.Args))
	fmt.Println("Main : region =>  ", region)
	if (len(os.Args) == 9) && (region != "undefined") {
		project = os.Args[1]
		env = os.Args[2]
		region = os.Args[3]
		mode = os.Args[4]
		configbucket = os.Args[5]
		configbucketkey = os.Args[6]
		executionarn = os.Args[7]
		result = os.Args[8]
		fmt.Println("Main : Obtained normal params ... ")
	} else {
		project = def_project
		env = def_env
		region = def_region
		mode = def_mode
		configbucket = def_confbucket
		configbucketkey = def_confbucketkey
		executionarn = "UNKNOWN"
		result = def_result
		fmt.Println("Main : Using default params ... ")
	}

	var swlRun SwlRun
	swlRun.Mode = mode
	swlRun.Result = result
	swlRun.SwlConf.ConfigPrefix = configbucketkey
	swlRun.ConfigBucket = configbucket
	swlRun.ConfigBucketKey = CONF_PREFIX + "/" + project + "/" + env + "/master.json"
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

	sns_session, err := my9sns.NewSnsSession(sess, region)
	if err != nil {
		fmt.Println("Error creating SNS Session")
	}

	swlRun.S3Session = s3_session
	swlRun.EC2Session = ec2_session
	swlRun.SFNSession = sfn_session
	swlRun.EcsSession = ecs_session
	swlRun.SnsSession = sns_session

	err = readConfig(&swlRun)
	fmt.Println("Reading SWL config from S3 ")

	//ecsTaskConfigFile := swlRun.SwlConf.EcsTaskConfig + "/" + project + "/" + env + ".json"
	fmt.Println("Main: Ecs cluster :", swlRun.SwlConf.EcsTaskCluster)

	swlRun.SwlConf.Project = project
	swlRun.SwlConf.Env = env
	swlRun.SwlConf.Region = region
	swlRun.SwlConf.ExecutionArn = executionarn

	err = swlRun.StepWorkerCleanupRun()
	if err != nil {
		fmt.Println("Error during StepWorkerCleanupRun")
	}

	fmt.Println("End of StepWorkerCleanupRun ... ! ")

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
