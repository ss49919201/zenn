#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { SqsStack } from '../lib/sqs-stack';
import { CloudWatchStack } from '../lib/cloudwatch-stack';

const app = new cdk.App();
const env = { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION }
const sqs = new SqsStack(app, 'SqsStack', { env: env });
new CloudWatchStack(app, 'CloudWatchStack', { env: env, queue: sqs.queue });
