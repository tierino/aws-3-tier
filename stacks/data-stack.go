package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
)

type DataStackProps struct {
	awscdk.StackProps
	Vpc awsec2.Vpc
}

type DataStack struct {
	awscdk.Stack
	DbSecurityGroupId    *string
	CacheSecurityGroupId *string
	DbClusterSecretName  *string
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
		Credentials: awsrds.Credentials_FromGeneratedSecret(S("dbuser"), &awsrds.CredentialsBaseOptions{}),
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

	cacheSecurityGroup := awsec2.NewSecurityGroup(stack, S("CacheSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              props.Vpc,
		Description:      S("Security group for ElastiCache cluster"),
		AllowAllOutbound: B(true),
	})

	// Same subnets as the ASG
	privateAppSubnets := props.Vpc.SelectSubnets(&awsec2.SubnetSelection{
		SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
	})

	subnetIds := make([]*string, len(*privateAppSubnets.Subnets))
	for i, subnet := range *privateAppSubnets.Subnets {
		subnetIds[i] = subnet.SubnetId()
	}

	cacheSubnetGroup := awselasticache.NewCfnSubnetGroup(stack, S("CacheSubnetGroup"), &awselasticache.CfnSubnetGroupProps{
		Description:          S("Subnet group for Elasticache cluster"),
		SubnetIds:            &subnetIds,
		CacheSubnetGroupName: S("elasticache-subnet-group"),
	})

	cacheCluster := awselasticache.NewCfnReplicationGroup(stack, S("ValkeyCluster"), &awselasticache.CfnReplicationGroupProps{
		ReplicationGroupDescription: S("Valkey replication group across 2 AZs"),
		Engine:                      S("valkey"),
		TransitEncryptionEnabled:    false,
		CacheNodeType:               S("cache.t3.micro"),
		NumCacheClusters:            N(2),
		AutomaticFailoverEnabled:    B(true),
		MultiAzEnabled:              B(true),
		CacheSubnetGroupName:        cacheSubnetGroup.CacheSubnetGroupName(),
		SecurityGroupIds:            &[]*string{cacheSecurityGroup.SecurityGroupId()},
		PreferredCacheClusterAZs: &[]*string{
			S("ap-southeast-2a"),
			S("ap-southeast-2b"),
		},
	})

	cacheCluster.AddDependency(cacheSubnetGroup)

	awscdk.NewCfnOutput(stack, S("AuroraClusterHostname"), &awscdk.CfnOutputProps{
		Value: dbCluster.ClusterEndpoint().Hostname(),
	})

	awscdk.NewCfnOutput(stack, S("CacheEndpoint"), &awscdk.CfnOutputProps{
		Value:       cacheCluster.AttrPrimaryEndPointAddress(),
		Description: S("Valkey primary endpoint address"),
	})

	awscdk.NewCfnOutput(stack, S("CachePort"), &awscdk.CfnOutputProps{
		Value:       cacheCluster.AttrPrimaryEndPointPort(),
		Description: S("Valkey primary endpoint port"),
	})

	return &DataStack{
		Stack:                stack,
		DbSecurityGroupId:    dbSecurityGroup.SecurityGroupId(),
		CacheSecurityGroupId: cacheSecurityGroup.SecurityGroupId(),
		DbClusterSecretName:  dbCluster.Secret().SecretName(),
	}
}
