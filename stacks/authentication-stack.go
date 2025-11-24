package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/constructs-go/constructs/v10"
)

type AuthenticationStackProps struct {
	awscdk.StackProps
}

type AuthenticationStack struct {
	awscdk.Stack
}

func NewAuthenticationStack(scope constructs.Construct, id string, props *AuthenticationStackProps) *AuthenticationStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	userPool := awscognito.NewUserPool(stack, S("UserPool"), &awscognito.UserPoolProps{
		SelfSignUpEnabled: B(true),
		SignInAliases: &awscognito.SignInAliases{
			Email: B(true),
		},
		AutoVerify: &awscognito.AutoVerifiedAttrs{
			Email: B(true),
		},
		PasswordPolicy: &awscognito.PasswordPolicy{
			MinLength:        N(8),
			RequireLowercase: B(true),
			RequireUppercase: B(true),
			RequireDigits:    B(true),
			RequireSymbols:   B(true),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	userPoolClient := userPool.AddClient(S("UserPoolClient"), &awscognito.UserPoolClientOptions{
		AuthFlows: &awscognito.AuthFlow{
			UserPassword: B(true),
			UserSrp:      B(true),
		},
	})

	awscdk.NewCfnOutput(stack, S("UserPoolId"), &awscdk.CfnOutputProps{
		Value: userPool.UserPoolId(),
	})

	awscdk.NewCfnOutput(stack, S("UserPoolClientId"), &awscdk.CfnOutputProps{
		Value: userPoolClient.UserPoolClientId(),
	})

	return &AuthenticationStack{
		Stack: stack,
	}
}
