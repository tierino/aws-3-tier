package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
)

type FrontendStackProps struct {
	awscdk.StackProps
	Alb awselasticloadbalancingv2.ApplicationLoadBalancer
}

type FrontendStack struct {
	awscdk.Stack
}

func NewFrontendStack(scope constructs.Construct, id string, props *FrontendStackProps) *FrontendStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	websiteBucket := awss3.NewBucket(stack, S("WebsiteBucket"), &awss3.BucketProps{
		Versioned:         B(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: B(true),
	})

	// Deploy website files to S3
	awss3deployment.NewBucketDeployment(stack, S("DeployWebsite"), &awss3deployment.BucketDeploymentProps{
		Sources:           &[]awss3deployment.ISource{awss3deployment.Source_Asset(S("./website"), nil)},
		DestinationBucket: websiteBucket,
	})

	// CloudFront distribution with S3 and ALB origins
	distribution := awscloudfront.NewDistribution(stack, S("Distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(websiteBucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{}),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			AllowedMethods:       awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS(),
			CachedMethods:        awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS(),
		},
		AdditionalBehaviors: &map[string]*awscloudfront.BehaviorOptions{
			"/api/*": {
				Origin: awscloudfrontorigins.NewLoadBalancerV2Origin(props.Alb, &awscloudfrontorigins.LoadBalancerV2OriginProps{
					ProtocolPolicy: awscloudfront.OriginProtocolPolicy_HTTP_ONLY,
				}),
				ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
				AllowedMethods:       awscloudfront.AllowedMethods_ALLOW_ALL(),
				CachePolicy:          awscloudfront.CachePolicy_CACHING_DISABLED(),
			},
		},
		DefaultRootObject: S("index.html"),
		EnableLogging:     B(true),
	})

	// todo: Route 53 configuration.

	awscdk.NewCfnOutput(stack, S("CloudFrontUrl"), &awscdk.CfnOutputProps{
		Value: distribution.DistributionDomainName(),
	})

	awscdk.NewCfnOutput(stack, S("WebsiteBucketName"), &awscdk.CfnOutputProps{
		Value: websiteBucket.BucketName(),
	})

	return &FrontendStack{
		Stack: stack,
	}
}
