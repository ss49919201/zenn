import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as sqs from 'aws-cdk-lib/aws-sqs';

export class SqsStack extends cdk.Stack {
  queue: sqs.Queue;
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    this.queue = new sqs.Queue(this, 'CdkTsExampleQueue', {
      visibilityTimeout: cdk.Duration.seconds(300)
    });
  }
}
