package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/constructs-go/constructs/v10"
)

type ContainerRegistryStackProps struct {
	awscdk.StackProps
}

type ContainerRegistryStack struct {
	awscdk.Stack
	Repo awsecr.Repository
}

func NewContainerRegistryStack(scope constructs.Construct, id string, props *ContainerRegistryStackProps) *ContainerRegistryStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	repo := awsecr.NewRepository(stack, S("Repo"), &awsecr.RepositoryProps{
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.NewCfnOutput(stack, S("ECRRepositoryUri"), &awscdk.CfnOutputProps{
		Value:       repo.RepositoryUri(),
		Description: S("ECR repository URI"),
	})

	return &ContainerRegistryStack{
		Stack: stack,
		Repo:  repo,
	}
}
