import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as sqs from 'aws-cdk-lib/aws-sqs';

type CloudWatchProps = cdk.StackProps & {
    queue: sqs.Queue;
};
export class CloudWatchStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: CloudWatchProps) {
        super(scope, id, props);
        props?.queue.metricNumberOfMessagesDeleted().
            createAlarm(this, 'AlarmNumberOfMessagesDeleted', {
                threshold: 1,
                evaluationPeriods: 1,
            });
    }
}
