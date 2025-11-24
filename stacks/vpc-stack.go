package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
)

type VpcStackProps struct {
	awscdk.StackProps
}

type VpcStack struct {
	awscdk.Stack
	Vpc awsec2.Vpc
}

func NewVpcStack(scope constructs.Construct, id string, props *VpcStackProps) *VpcStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := awsec2.NewVpc(stack, S("Vpc"), &awsec2.VpcProps{
		MaxAzs:      N(2),
		NatGateways: N(2),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       S("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   N(24),
			},
			{
				Name:       S("PrivateApp"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   N(24),
			},
			{
				Name:       S("PrivateDb"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   N(24),
			},
		},
	})

	vpc.AddInterfaceEndpoint(S("SsmEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service: awsec2.InterfaceVpcEndpointAwsService_SSM(),
	})

	vpc.AddInterfaceEndpoint(S("SsmMessagesEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service: awsec2.InterfaceVpcEndpointAwsService_SSM_MESSAGES(),
	})

	vpc.AddInterfaceEndpoint(S("Ec2MessagesEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service: awsec2.InterfaceVpcEndpointAwsService_EC2_MESSAGES(),
	})

	return &VpcStack{
		Stack: stack,
		Vpc:   vpc,
	}
}
