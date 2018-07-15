package my9ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ECSSession struct {
	svc *ecs.ECS
}

type ECSTaskDefinition struct {
	Family               string `json:"family"`
	ContainerDefinitions []struct {
		Command           string `json:"command"`
		Cpu               int64  `json:"cpu"`
		Hostname          string `json:"hostname"`
		Image             string `json:"image"`
		Memory            int64  `json:"memory"`
		Name              string `json:"containername"`
		DisableNetworking bool   `json:"disablenetworking"`
		Environment       []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}
		WorkingDirectory string `json:"workingdirectory"`
	} `json:"containerdefinitions"`
	Volumes []struct {
		Host []struct {
			SourcePath *string `json:"sourcepath"`
		} `json:"host"`
		Name string `json:"volname"`
	} `json:"volumes"`
}

type ECSRunTaskIn struct {
	TaskDefinition string
	Cluster        string
	Count          int64
}

type DescEcsTaskIn struct {
	Cluster string
	Tasks   []*string
}

type ContainerInstance struct {
	Attributes []struct {
		Name  string
		Value string
	}
	Cluster              string
	ContainerInstanceArn string
}

type ListContInstIn struct {
	Cluster    string
	MaxResults int64
	NextToken  string
}

type DescContInstIn struct {
	Cluster            string
	ContainerInstances []*string
}

func NewEcsSession(sess client.ConfigProvider, region string) (ec ECSSession, err error) {
	ec.svc = ecs.New(sess, aws.NewConfig().WithRegion(region))
	return ec, err
}

func (ec *ECSSession) CreateCluster(clustername string) (err error) {

	params := &ecs.CreateClusterInput{
		ClusterName: aws.String(clustername),
	}
	resp, err := ec.svc.CreateCluster(params)
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

func (ec *ECSSession) RegisterECSTask(taskDefIn ECSTaskDefinition) (err error) {

	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{ // Required
			{ // Required
				Cpu:               &taskDefIn.ContainerDefinitions[0].Cpu,
				DisableNetworking: &taskDefIn.ContainerDefinitions[0].DisableNetworking,
				Environment: []*ecs.KeyValuePair{
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[0].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[0].Value,
					},
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[1].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[1].Value,
					},
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[2].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[2].Value,
					},
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[3].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[3].Value,
					},
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[4].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[4].Value,
					},
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[5].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[5].Value,
					},
					{ // Required
						Name:  &taskDefIn.ContainerDefinitions[0].Environment[6].Name,
						Value: &taskDefIn.ContainerDefinitions[0].Environment[6].Value,
					},
				},
				Hostname:         &taskDefIn.ContainerDefinitions[0].Hostname,
				Image:            &taskDefIn.ContainerDefinitions[0].Image,
				Memory:           &taskDefIn.ContainerDefinitions[0].Memory,
				Name:             &taskDefIn.ContainerDefinitions[0].Name,
				WorkingDirectory: &taskDefIn.ContainerDefinitions[0].WorkingDirectory,
			},
			// More values...
		},
		Family: &taskDefIn.Family, // Required
		//},
	}
	fmt.Println("container name ", taskDefIn.ContainerDefinitions[0].Name)
	resp, err := ec.svc.RegisterTaskDefinition(params)

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

func (ec *ECSSession) RegisterEC2Instance(contInstance ContainerInstance) (err error) {

	params := &ecs.RegisterContainerInstanceInput{
		Attributes: []*ecs.Attribute{
			{ // Required
				Name:  &contInstance.Attributes[0].Name, // Required
				Value: &contInstance.Attributes[0].Value,
			},
			// More values...
		},
		Cluster:              &contInstance.Cluster,
		ContainerInstanceArn: &contInstance.ContainerInstanceArn,
	}
	resp, err := ec.svc.RegisterContainerInstance(params)

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

func (ec *ECSSession) RunECSTask(ecsTask ECSRunTaskIn) (resp *ecs.RunTaskOutput, err error) {

	params := &ecs.RunTaskInput{
		TaskDefinition: &ecsTask.TaskDefinition, // Required
		Cluster:        &ecsTask.Cluster,
		Count:          &ecsTask.Count,
	}
	resp, err = ec.svc.RunTask(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return resp, err
	}

	// Pretty-print the response data.
	fmt.Println(resp)
	return resp, err
}

// for cleanup

func (ec *ECSSession) DeRegisterEC2Instance(contInstance ContainerInstance) (resp *ecs.DeregisterContainerInstanceOutput, err error) {

	params := &ecs.DeregisterContainerInstanceInput{
		Cluster:           &contInstance.Cluster,
		ContainerInstance: &contInstance.ContainerInstanceArn,
	}
	resp, err = ec.svc.DeregisterContainerInstance(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return resp, err
	}

	// Pretty-print the response data.
	fmt.Println(resp)
	return resp, err
}

func (ec *ECSSession) DescribeEC2ContInstance(descContInstIn DescContInstIn) (resp *ecs.DescribeContainerInstancesOutput, err error) {

	params := &ecs.DescribeContainerInstancesInput{
		Cluster:            &descContInstIn.Cluster,
		ContainerInstances: descContInstIn.ContainerInstances,
	}
	resp, err = ec.svc.DescribeContainerInstances(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return resp, err
	}

	// Pretty-print the response data.
	fmt.Println(resp)
	return resp, err
}

func (ec *ECSSession) ListContInstances(listContInstIn ListContInstIn) (resp *ecs.ListContainerInstancesOutput, err error) {

	params := &ecs.ListContainerInstancesInput{
		Cluster:    &listContInstIn.Cluster,
		MaxResults: &listContInstIn.MaxResults,
	}
	resp, err = ec.svc.ListContainerInstances(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return resp, err
	}

	// Pretty-print the response data.
	fmt.Println(resp)
	return resp, err
}

func (ec *ECSSession) DescribeEcsTask(descEcsTaskIn DescEcsTaskIn) (resp *ecs.DescribeTasksOutput, err error) {

	params := &ecs.DescribeTasksInput{
		Cluster: &descEcsTaskIn.Cluster,
		Tasks:   descEcsTaskIn.Tasks,
	}

	fmt.Println("DescribeEcsTask : params : ", params)
	resp, err = ec.svc.DescribeTasks(params)

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(resp)
	return resp, err
}
