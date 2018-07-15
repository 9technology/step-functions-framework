# step-functions-framework
Step Functions Framework with ECS Cluster Orchestration

Step Functions Framework provides ECS Cluster Orchestration for running jobs as Step Functions State Machines.

This is composed of two generic lambda functions (Launcher and Cleanup)

Build Launcher Lambda:  

cd step-worker-launcher
make build


Build Cleanup Lambda:  

cd step-worker-cleanup
make build


For more details refer to the related blog posts.
