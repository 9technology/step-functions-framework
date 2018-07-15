package main

import (
	"fmt"
	"my9awsgo/my9ec2"
	"my9awsgo/my9ecs"
	"my9awsgo/my9sns"
	"strings"
)

func (swlRun *SwlRun) CleanupStateMachine() (err error) {

	// SNS : send email with result
	sub := "StepFunctionsResult: " + swlRun.SwlConf.Project + " " + swlRun.SwlConf.Env + " :: " + swlRun.Result
	mesg := sub + " :: EXECUTIONARN : " + swlRun.SwlConf.ExecutionArn
	var publishIn my9sns.SNSPublishInput
	publishIn.Message = mesg
	publishIn.TopicArn = swlRun.SwlConf.SnsTopicArn
	publishIn.Subject = sub

	swlRun.SnsSession.SnsPublish(publishIn)

	// Deregister unused compute from EcsTaskCluster

	// List instances in the cluster

	var listContInstIn my9ecs.ListContInstIn
	listContInstIn.Cluster = swlRun.SwlConf.EcsTaskCluster
	listContInstIn.MaxResults = 100

	listContInstOut, err := swlRun.EcsSession.ListContInstances(listContInstIn)
	if err != nil {
		fmt.Println("CleanupStateMachine: Error in ListContInstances :: Error=%s", err)
		panic(err)
	}

	// Determine which instances do not have tasks running

	var descContInstIn my9ecs.DescContInstIn
	descContInstIn.Cluster = swlRun.SwlConf.EcsTaskCluster
	descContInstIn.ContainerInstances = listContInstOut.ContainerInstanceArns

	fmt.Println("CleanupStateMachine: Instance list for cluster ", descContInstIn.ContainerInstances)
	if len(descContInstIn.ContainerInstances) == 0 {
		fmt.Println("CleanupStateMachine: No instances in this cluster. Nothing to do ! ")
		return err
	}

	descContInstOut, err := swlRun.EcsSession.DescribeEC2ContInstance(descContInstIn)
	if err != nil {
		fmt.Println("CleanupStateMachine: Error in DescribeEC2ContInstance :: Error=%s", err)
		panic(err)
	}

	var InstanceDeRegisterList []*string
	used := false

	for _, instDsc := range descContInstOut.ContainerInstances {
		instDeReg := new(string)
		if *instDsc.PendingTasksCount == 0 && *instDsc.RunningTasksCount == 0 {
			DescTagsOut, _ := swlRun.EC2Session.Ec2DescribeTags(*instDsc.Ec2InstanceId)
			if swlRun.Mode == "cleanup_used_and_new" {
				used = true
			} else {
				for _, tag := range DescTagsOut.Tags {
					if *tag.Key == "status" && *tag.Value == "used" {
						fmt.Println("CleanupStateMachine: Instance tags :: status : used ", instDsc.Ec2InstanceId)
						used = true
					}
				}
			}
			fmt.Println("CleanupStateMachine: Instance tags :: status :: value : used? =>  ", used, instDsc.Ec2InstanceId)
			if used {
				instDeReg = instDsc.ContainerInstanceArn
				fmt.Println("CleanupStateMachine: DeRegister List: No Task running or pending on this instance : ", *instDsc.ContainerInstanceArn)
				InstanceDeRegisterList = append(InstanceDeRegisterList, instDeReg)
			}
		}
		used = false
	}

	fmt.Println("CleanupStateMachine: DeRegister List ", InstanceDeRegisterList)

	if len(InstanceDeRegisterList) == 0 {
		fmt.Println("CleanupStateMachine: DeRegister List Empty :: Nothing to DeRegister and/or Cleanup. Exiting ... !")
		return err
	}

	// De Register those container instances

	var deRegInstIn my9ecs.ContainerInstance
	deRegInstIn.Cluster = swlRun.SwlConf.EcsTaskCluster

	for _, instDrl := range InstanceDeRegisterList {
		deRegInstIn.ContainerInstanceArn = *instDrl
		swlRun.EcsSession.DeRegisterEC2Instance(deRegInstIn)
	}

	// Check again with previous list if any tasks are running or pending

	var InstanceCleanupList []string

	var descContInstInClean my9ecs.DescContInstIn
	descContInstInClean.Cluster = swlRun.SwlConf.EcsTaskCluster
	descContInstInClean.ContainerInstances = InstanceDeRegisterList

	fmt.Println("CleanupStateMachine: DeRegister List2 ", descContInstInClean.ContainerInstances)

	descContInstOutClean, err := swlRun.EcsSession.DescribeEC2ContInstance(descContInstInClean)
	if err != nil {
		fmt.Println("CleanupStateMachine: Error in DescribeEC2ContInstance (Clean) :: Error=%s", err)
		panic(err)
	}

	for _, instCl := range descContInstOutClean.ContainerInstances {
		instClean := new(string)
		if *instCl.PendingTasksCount == 0 && *instCl.RunningTasksCount == 0 {
			instClean = instCl.Ec2InstanceId
			fmt.Println("CleanupStateMachine: Cleanup List: This instance confirmed for cleanup : ", *instCl.Ec2InstanceId)
			InstanceCleanupList = append(InstanceCleanupList, *instClean)
		}
	}

	fmt.Println("CleanupStateMachine: Instance Cleanup List ", InstanceCleanupList)

	if len(InstanceCleanupList) == 0 {
		fmt.Println("CleanupStateMachine: CleanUpList List Empty :: Nothing to Cleanup. Exiting ... !")
		return err
	}

	// Terminate unused compute

	var ec2TermIn my9ec2.EC2InstIn
	for _, instTerm := range InstanceCleanupList {
		ec2TermIn.InstanceId = instTerm
		err := swlRun.EC2Session.TerminateEC2Instance(ec2TermIn)
		if err != nil {
			fmt.Println("CleanupStateMachine: Error in TerminateEC2Instance  :: Error=%s", err)
		}
	}

	fmt.Println("CleanupStateMachine: End of CleanUpStateMachine run ! ")
	if strings.Contains(swlRun.Result, "PASS") {
		fmt.Println("CleanupStateMachine: Final Result is PASS ! ")
	} else {
		fmt.Println("CleanupStateMachine: Final Result is FAIL ! ")
		panic(err)
	}

	return err
}
