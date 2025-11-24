package stacks

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
)

type AppStackProps struct {
	awscdk.StackProps
	Vpc             awsec2.Vpc
	DbSecurityGroup awsec2.SecurityGroup
	Repo            awsecr.Repository
}

type AppStack struct {
	awscdk.Stack
	Alb awselasticloadbalancingv2.ApplicationLoadBalancer
}

func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) *AppStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	albSecurityGroup := awsec2.NewSecurityGroup(stack, S("AlbSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              props.Vpc,
		Description:      S("Security group for Application Load Balancer"),
		AllowAllOutbound: B(true),
	})
	albSecurityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(N(80)),
		S("Allow HTTP from anywhere"),
		B(false),
	)
	albSecurityGroup.AddIngressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(N(443)),
		S("Allow HTTPS from anywhere"),
		B(false),
	)

	alb := awselasticloadbalancingv2.NewApplicationLoadBalancer(stack, S("AppLoadBalancer"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:            props.Vpc,
		InternetFacing: B(true),
		SecurityGroup:  albSecurityGroup,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
	})

	listener := alb.AddListener(S("HttpListener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
		Port: N(80),
		Open: B(true),
	})

	awscdk.NewCfnOutput(stack, S("LoadBalancerDns"), &awscdk.CfnOutputProps{
		Value: alb.LoadBalancerDnsName(),
	})

	// Security group for EC2 instances
	ec2SecurityGroup := awsec2.NewSecurityGroup(stack, S("Ec2SecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              props.Vpc,
		Description:      S("Security group for EC2 instances"),
		AllowAllOutbound: B(true),
	})
	ec2SecurityGroup.AddIngressRule(
		albSecurityGroup,
		awsec2.Port_Tcp(N(8080)),
		S("Allow HTTP from ALB"),
		B(false),
	)

	// Allow EC2 instances to connect to Aurora
	props.DbSecurityGroup.AddIngressRule(
		ec2SecurityGroup,
		awsec2.Port_Tcp(N(5432)),
		S("Allow PostgreSQL from EC2 instances"),
		B(false),
	)

	userDataScriptBytes, err := os.ReadFile("userdata/configure.sh")
	if err != nil {
		panic(err)
	}
	userDataScript := string(userDataScriptBytes)

	userData := awsec2.UserData_ForLinux(nil)
	userData.AddCommands(
		S(fmt.Sprintf("REGION=%s", *stack.Region())),
		S(fmt.Sprintf("ACCOUNT=%s", *stack.Account())),
		S(fmt.Sprintf("REPO=%s", *props.Repo.RepositoryUri())),
		S(userDataScript),
	)

	asgRole := awsiam.NewRole(stack, S("AppAutoScalingGroupRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(S("ec2.amazonaws.com"), nil),
	})

	asgRole.AddManagedPolicy(
		awsiam.ManagedPolicy_FromAwsManagedPolicyName(
			S("AmazonSSMManagedInstanceCore"),
		),
	)

	asgRole.AddManagedPolicy(
		awsiam.ManagedPolicy_FromAwsManagedPolicyName(
			S("AmazonEC2ContainerRegistryReadOnly"),
		),
	)

	// Auto Scaling Group with EC2 instances
	asg := awsautoscaling.NewAutoScalingGroup(stack, S("AppAutoScalingGroup"), &awsautoscaling.AutoScalingGroupProps{
		Vpc: props.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MICRO),
		MachineImage: awsec2.MachineImage_LatestAmazonLinux2023(&awsec2.AmazonLinux2023ImageSsmParameterProps{
			CpuType: awsec2.AmazonLinuxCpuType_X86_64,
		}),
		UserData:        userData,
		MinCapacity:     N(2),
		MaxCapacity:     N(6),
		DesiredCapacity: N(2),
		SecurityGroup:   ec2SecurityGroup,
		Role:            asgRole,
	})

	// Attach ASG to ALB target group
	listener.AddTargets(S("AppTargets"), &awselasticloadbalancingv2.AddApplicationTargetsProps{
		Port:    N(8080),
		Targets: &[]awselasticloadbalancingv2.IApplicationLoadBalancerTarget{asg},
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Path:     S("/health"),
			Interval: awscdk.Duration_Seconds(N(30)),
		},
	})

	awscdk.NewCfnOutput(stack, S("AutoScalingGroupName"), &awscdk.CfnOutputProps{
		Value: asg.AutoScalingGroupName(),
	})

	return &AppStack{
		Stack: stack,
		Alb:   alb,
	}
}
