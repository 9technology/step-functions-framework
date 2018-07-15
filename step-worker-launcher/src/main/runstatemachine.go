package main

import (
	"encoding/json"
	"fmt"
	"my9awsgo/my9sfn"
	"time"
)

const GENERIC_STEP_PATH = "generic"

func (swlRun *SwlRun) RunStateMachine() (err error) {
	var smInput StateMachineInput
	smInput.Project = swlRun.SwlConf.Project
	smInput.Env = swlRun.SwlConf.Env
	smInput.Region = swlRun.SwlConf.Region
	smInput.ConfigBucket = swlRun.ConfigBucket
	smInput.ConfigBucketKey = swlRun.SwlConf.ConfigPrefix

	instance_id, err := swlRun.LaunchCompute()
	smInput.Result = instance_id

	var runSmIn my9sfn.RunSmIn
	runSmIn.Name = swlRun.SwlConf.Project + "_" + getUniqueExecutionName()

	smInput.ExecutionArn = "arn:aws:states:" + swlRun.SwlConf.Region + ":" + swlRun.SwlConf.AwsAccountId + ":execution:" + swlRun.SwlConf.StateMachineName + ":" + runSmIn.Name
	smInput.Mode = "launch_stepworker"

	runSmIn.StateMachineArn = swlRun.SwlConf.StateMachineArn
	inputData, err := json.Marshal(smInput)
	if err != nil {
		fmt.Println("RunStateMachine: Error in reading JSON marshalling State Machine Input :: Error=%s", err)
		panic(err)
	}
	runSmIn.Input = string(inputData)

	runSmOut, err := swlRun.SFNSession.SfnRunStateMachine(runSmIn)
	if err != nil {
		fmt.Println("RunStateMachine: Error in SfnRunStateMachine :: Error=%s", err)
		panic(err)
	}
	fmt.Println("RunStateMachine: running StateMachine ", runSmOut)
	return err
}

func (swlRun *SwlRun) LaunchCompute() (instance_id string, err error) {

	swlRun.Ec2InstanceConfigFilePath = swlRun.SwlConf.ConfigPrefix + "/" + swlRun.SwlConf.Project + "/" + swlRun.SwlConf.Env + "/" + GENERIC_STEP_PATH + "/" + swlRun.SwlConf.Ec2InstConfig
	fmt.Println("Launch compute for this State Machine run ... ")

	instance_id, err = swlRun.CreateEcsContainerInstance(false)
	fmt.Println("Wait for compute spin up ...")
	time.Sleep(240 * 1000 * time.Millisecond) // sleep for 4 mins
	return instance_id, err
}
