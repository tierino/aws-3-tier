package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
)

type DataStackProps struct {
	awscdk.StackProps
	Vpc awsec2.Vpc
}

type DataStack struct {
	awscdk.Stack
	DbSecurityGroup awsec2.SecurityGroup
}

func NewDataStack(scope constructs.Construct, id string, props *DataStackProps) *DataStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	dbSecurityGroup := awsec2.NewSecurityGroup(stack, S("DbSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              props.Vpc,
		Description:      S("Security group for Aurora database"),
		AllowAllOutbound: B(true),
	})

	dbCluster := awsrds.NewDatabaseCluster(stack, S("AuroraCluster"), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
			Version: awsrds.AuroraPostgresEngineVersion_VER_17_5(),
		}),
		Credentials: awsrds.Credentials_FromGeneratedSecret(S("admin"), &awsrds.CredentialsBaseOptions{}),
		Writer: awsrds.ClusterInstance_Provisioned(S("Writer"), &awsrds.ProvisionedClusterInstanceProps{
			InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MEDIUM),
		}),
		Readers: &[]awsrds.IClusterInstance{
			awsrds.ClusterInstance_Provisioned(S("Reader"), &awsrds.ProvisionedClusterInstanceProps{
				InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MEDIUM),
			}),
		},
		Vpc: props.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups:      &[]awsec2.ISecurityGroup{dbSecurityGroup},
		DefaultDatabaseName: S("appdb"),
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.NewCfnOutput(stack, S("DbSecretName"), &awscdk.CfnOutputProps{
		Value: dbCluster.Secret().SecretName(),
	})

	awscdk.NewCfnOutput(stack, S("DatabaseEndpoint"), &awscdk.CfnOutputProps{
		Value: dbCluster.ClusterEndpoint().Hostname(),
	})

	return &DataStack{
		Stack:           stack,
		DbSecurityGroup: dbSecurityGroup,
	}
}
