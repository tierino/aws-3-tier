package main

import (
	"os"

	"multi-tier-ha/stacks"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	defer jsii.Close()

	app := awscdk.NewApp(&awscdk.AppProps{})

	containerRegistryStack := stacks.NewContainerRegistryStack(app, "ContainerRegistryStack", &stacks.ContainerRegistryStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	vpcStack := stacks.NewVpcStack(app, "VpcStack", &stacks.VpcStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	dbStack := stacks.NewDataStack(app, "DataStack", &stacks.DataStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Vpc: vpcStack.Vpc,
	})

	appStack := stacks.NewAppStack(app, "AppStack", &stacks.AppStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Vpc:             vpcStack.Vpc,
		DbSecurityGroup: dbStack.DbSecurityGroup,
		Repo:            containerRegistryStack.Repo,
	})

	stacks.NewFrontendStack(app, "FrontendStack", &stacks.FrontendStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Alb: appStack.Alb,
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Region:  jsii.String(os.Getenv("REGION")),
		Account: jsii.String(os.Getenv("ACCOUNT_ID")),
	}
}
