package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"my9awsgo/my9ec2"
	"my9awsgo/my9ecs"
	"my9awsgo/my9sfn"
)

func (swlRun *SwlRun) RunStepWorker() (err error) {
	var stepName string
	var getExecHistIn my9sfn.GetExecHistIn
	getExecHistIn.ExecutionArn = swlRun.SwlConf.ExecutionArn
	getExecHistIn.ReverseOrder = true

	getExecHistOut, err := swlRun.SFNSession.SfnGetExecutionHistory(getExecHistIn)
	if err != nil {
		fmt.Println("RunStepWorker: Error in SfnGetExecutionHistory :: Error=%s", err)
		panic(err)
	}

	for _, histEvent := range getExecHistOut.Events {
		fmt.Println("RunStepWorker: histEvent :: ", histEvent)
		if *histEvent.Type == "TaskStateEntered" {

			stepName = *histEvent.StateEnteredEventDetails.Name
			break
		}
	}
	stepNameParts := strings.Split(stepName, "_")
	stepPath := stepNameParts[1]

	swlRun.EcsTaskConfigFilePath = swlRun.SwlConf.ConfigPrefix + "/" + swlRun.SwlConf.Project + "/" + swlRun.SwlConf.Env + "/" + stepPath + "/" + swlRun.SwlConf.EcsContConfig
	swlRun.Ec2InstanceConfigFilePath = swlRun.SwlConf.ConfigPrefix + "/" + swlRun.SwlConf.Project + "/" + swlRun.SwlConf.Env + "/" + stepPath + "/" + swlRun.SwlConf.Ec2InstConfig

	/*
		Create ECS Task Definition
	*/

	var ecsTaskDef my9ecs.ECSTaskDefinition
	ecsTaskConfData, err := readConfFile(swlRun.S3Session, swlRun.ConfigBucket, swlRun.EcsTaskConfigFilePath)
	err = json.Unmarshal(ecsTaskConfData, &ecsTaskDef)
	if err != nil {
		fmt.Println("error opening ECS Task config file")
		panic(err)
	}

	configbucketkey_worker := swlRun.SwlConf.ConfigPrefix + "/" + swlRun.SwlConf.Project + "/" + swlRun.SwlConf.Env + "/" + stepPath + "/" + swlRun.SwlConf.WorkerConfig

	swlRun.NewEc2InstanceId = swlRun.Result
	swlRun.Result = "UNKNOWN"

	ecsTaskDef.ContainerDefinitions[0].Environment[0].Name = "PROJECT"
	ecsTaskDef.ContainerDefinitions[0].Environment[0].Value = swlRun.SwlConf.Project
	ecsTaskDef.ContainerDefinitions[0].Environment[1].Name = "ENV"
	ecsTaskDef.ContainerDefinitions[0].Environment[1].Value = swlRun.SwlConf.Env
	ecsTaskDef.ContainerDefinitions[0].Environment[2].Name = "REGION"
	ecsTaskDef.ContainerDefinitions[0].Environment[2].Value = swlRun.SwlConf.Region
	ecsTaskDef.ContainerDefinitions[0].Environment[3].Name = "CONFIGBUCKET"
	ecsTaskDef.ContainerDefinitions[0].Environment[3].Value = swlRun.ConfigBucket
	ecsTaskDef.ContainerDefinitions[0].Environment[4].Name = "CONFIGBUCKETKEY"
	ecsTaskDef.ContainerDefinitions[0].Environment[4].Value = configbucketkey_worker
	ecsTaskDef.ContainerDefinitions[0].Environment[5].Name = "EXECUTIONARN"
	ecsTaskDef.ContainerDefinitions[0].Environment[5].Value = swlRun.SwlConf.ExecutionArn
	ecsTaskDef.ContainerDefinitions[0].Environment[6].Name = "RESULT"
	ecsTaskDef.ContainerDefinitions[0].Environment[6].Value = swlRun.Result

	fmt.Println("Registering ECS task")
	fmt.Println("ECS task definition : ", ecsTaskDef)
	resp := swlRun.EcsSession.RegisterECSTask(ecsTaskDef)
	fmt.Println("ECS Task Response : ... ")
	fmt.Println(resp)

	/*
		Run ECS Task
	*/
	fmt.Println("Wait for ECS agent to register before RunTask ...")
	time.Sleep(180 * 1000 * time.Millisecond) // sleep for 3 mins

	taskrunning := false
	count := 1

	var ecsTaskRun my9ecs.ECSRunTaskIn
	ecsTaskRun.TaskDefinition = ecsTaskDef.Family
	ecsTaskRun.Cluster = swlRun.SwlConf.EcsTaskCluster
	ecsTaskRun.Count = 1

	for taskrunning == false && count <= swlRun.SwlConf.LswRetryCount {
		fmt.Println("Running ECS task")
		ecsTaskRunResp, err := swlRun.EcsSession.RunECSTask(ecsTaskRun)
		if err != nil {
			if strings.Contains(err.Error(), "No Container Instances were found in your cluster") {
				fmt.Println("Error in running EcsTask :: No container instances found ::  Spinning up EC2 container instance")
				_, err = swlRun.CreateEcsContainerInstance(true)
				continue
			}
			fmt.Println("RunStepWorker:  RunEcsTask fatal error :: ", err)
			panic(err)
		}
		fmt.Println("RunStepWorker:  ecsTaskRunResp :: ", ecsTaskRunResp)
		fmt.Println("RunStepWorker:  ecsTaskRunResp->err :: ", err)
		if len(ecsTaskRunResp.Failures) != 0 {
			fmt.Println("Error in running EcsTask :: Spinning up EC2 container instance")
			if *ecsTaskRunResp.Failures[0].Reason == "PlatformTaskDefinitionIncompatibilityException" ||
				*ecsTaskRunResp.Failures[0].Reason == "RESOURCE:MEMORY" ||
				*ecsTaskRunResp.Failures[0].Reason == "RESOURCE:CPU" {
				_, err = swlRun.CreateEcsContainerInstance(true)
			} else {
				fmt.Println("Fatal error running ECS Task ")
				panic(err)
			}
		} else {
			taskArn := *ecsTaskRunResp.Tasks[0].TaskArn
			fmt.Println("RunStepWorker : ECS Task ARN:  ", taskArn)
			time.Sleep(2 * 1000 * time.Millisecond) // sleep for 2 secs
			taskrunning, err = swlRun.CheckEcsTaskRun(taskArn)
		}
		fmt.Println(resp)
		count++
	}
	return
}

func (swlRun *SwlRun) CreateEcsContainerInstance(wait bool) (instance_id string, err error) {
	/*
		Create EC2 Container Instance
	*/

	var ecsContainerInstance my9ec2.EC2Instance
	ec2Conf, err := readConfFile(swlRun.S3Session, swlRun.ConfigBucket, swlRun.Ec2InstanceConfigFilePath)
	err = json.Unmarshal(ec2Conf, &ecsContainerInstance)
	if err != nil {
		fmt.Println("error opening EC2 Instance config file")
		panic(err)
	}

	ec2userdata := "#!/bin/bash\necho ECS_CLUSTER=" + swlRun.SwlConf.EcsTaskCluster + " >> /etc/ecs/ecs.config"
	ec2userdataEnc := b64.StdEncoding.EncodeToString([]byte(ec2userdata))
	ecsContainerInstance.UserData = ec2userdataEnc

	runResult, err := swlRun.EC2Session.CreateEC2Instance(ecsContainerInstance)
	fmt.Println("Running EC2 container instance", runResult)
	instance_id = *runResult.Instances[0].InstanceId

	/*
		Add Tags to EC2 instance
	*/

	var tagInstOwner my9ec2.CreateTagIn
	tagInstOwner.Resource = instance_id
	tagInstOwner.Key = "owner"
	tagInstOwner.Value = ecsContainerInstance.Owner
	err = swlRun.EC2Session.CreateTag(tagInstOwner)

	var tagInstEnv my9ec2.CreateTagIn
	tagInstEnv.Resource = instance_id
	tagInstEnv.Key = "env"
	tagInstEnv.Value = ecsContainerInstance.Env
	err = swlRun.EC2Session.CreateTag(tagInstEnv)

	var tagInstStatus my9ec2.CreateTagIn
	tagInstStatus.Resource = instance_id
	tagInstStatus.Key = "status"
	tagInstStatus.Value = "new"
	err = swlRun.EC2Session.CreateTag(tagInstStatus)

	/*
		Check EC2 instance status
	*/

	if wait {
		var contInst my9ec2.EC2InstIn
		contInst.InstanceId = instance_id
		bootup := false
		for bootup == false {
			fmt.Println("Waiting for instance to boot up", contInst.InstanceId)
			instStatus, err := swlRun.EC2Session.DescribeEC2InstStatus(contInst)
			if err != nil {
				fmt.Println("Error obtaining EC2 instance status ...")
			}
			if instStatus.InstanceStatuses != nil {
				fmt.Println("Instance Status: ", *instStatus.InstanceStatuses[0].InstanceState.Name)
				bootup = *instStatus.InstanceStatuses[0].InstanceState.Name == "running" &&
					*instStatus.InstanceStatuses[0].InstanceStatus.Status == "ok" &&
					*instStatus.InstanceStatuses[0].SystemStatus.Status == "ok"
			}
			time.Sleep(2 * 1000 * time.Millisecond) // sleep for 2 secs
		}
		fmt.Println("Boot up complete", contInst.InstanceId)
		fmt.Println("Wait 20 secs for ECS agent to register with Cluster ...")
		time.Sleep(20 * 1000 * time.Millisecond) // sleep for 20 secs
	} // end of if wait
	return instance_id, err
}

func (swlRun *SwlRun) CheckEcsTaskRun(taskArn string) (isTaskRunning bool, err error) {
	var descEcsTaskIn my9ecs.DescEcsTaskIn
	var contInstArn string
	descEcsTaskIn.Cluster = swlRun.SwlConf.EcsTaskCluster
	descEcsTask := new(string)
	*descEcsTask = taskArn
	descEcsTaskIn.Tasks = append(descEcsTaskIn.Tasks, descEcsTask)
	count := 0
	isTaskRunning = false
	taskStatus := "UNKNOWN"
	for !isTaskRunning && count < 10 {
		respDescTask, err := swlRun.EcsSession.DescribeEcsTask(descEcsTaskIn)
		if len(respDescTask.Tasks) > 0 {
			contInstArn = *respDescTask.Tasks[0].ContainerInstanceArn
		} else {
			time.Sleep(2 * 1000 * time.Millisecond) // sleep for 2 secs
			count++
			continue
		}
		if err != nil {
			fmt.Println("CheckEcsTaskRun : Error in DescribeEcsTask ...")
			isTaskRunning = true
			return isTaskRunning, err
		}
		if len(respDescTask.Tasks) < 1 {
			isTaskRunning = true
			return isTaskRunning, err
		}

		taskStatus = *respDescTask.Tasks[0].LastStatus
		if taskStatus == "RUNNING" {
			isTaskRunning = true
			swlRun.TagAsUsed(contInstArn)
		} else if taskStatus != "PENDING" {
			isTaskRunning = true
		} else {
			// wait 2 secs
			time.Sleep(2 * 1000 * time.Millisecond) // sleep for 2 secs
			count++
		}
	}
	if taskStatus == "PENDING" {
		swlRun.TagAsUsed(contInstArn)
		isTaskRunning = true
	}
	return isTaskRunning, err
}

func (swlRun *SwlRun) TagAsUsed(contInstArn string) (err error) {

	var descContInstIn my9ecs.DescContInstIn
	descContInstIn.Cluster = swlRun.SwlConf.EcsTaskCluster
	contInst := new(string)
	*contInst = contInstArn
	descContInstIn.ContainerInstances = append(descContInstIn.ContainerInstances, contInst)

	descContInstOut, err := swlRun.EcsSession.DescribeEC2ContInstance(descContInstIn)
	if err != nil {
		fmt.Println("TagAsUsed: Error in DescribeEC2ContInstance :: Error=%s", err)
	}

	var tagInstStatus my9ec2.CreateTagIn
	tagInstStatus.Resource = *descContInstOut.ContainerInstances[0].Ec2InstanceId
	tagInstStatus.Key = "status"
	tagInstStatus.Value = "used"
	err = swlRun.EC2Session.CreateTag(tagInstStatus)
	fmt.Println("TagAsUsed: Instance marked as 'used' InstancId=", tagInstStatus.Resource)

	if tagInstStatus.Resource != swlRun.NewEc2InstanceId {
		tagInstStatus.Resource = swlRun.NewEc2InstanceId
		err = swlRun.EC2Session.CreateTag(tagInstStatus)
		fmt.Println("TagAsUsed: New Instance marked as 'used' InstancId=", tagInstStatus.Resource)
	}
	return err
}
