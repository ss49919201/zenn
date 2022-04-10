package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkGoExampleStackProps struct {
	awscdk.StackProps
}

func NewCdkGoExampleStack(scope constructs.Construct, id string, props *CdkGoExampleStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// デフォルトのVPCをインポート
	defaultVpc := awsec2.Vpc_FromLookup(stack, jsii.String("DefaultVPC"), &awsec2.VpcLookupOptions{
		IsDefault: jsii.Bool(true),
	})

	// EC2インスタンスを作成
	awsec2.NewInstance(stack, jsii.String("ExampleInstance"), &awsec2.InstanceProps{
		InstanceType: awsec2.NewInstanceType(jsii.String("t3.micro")),
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2,
		}),
		Vpc: defaultVpc,
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewCdkGoExampleStack(app, "CdkGoExampleStack", &CdkGoExampleStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("ACCOUNT_ID")),
		Region:  jsii.String(os.Getenv("REGION")),
	}
}
