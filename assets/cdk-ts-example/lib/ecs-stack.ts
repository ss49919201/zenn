import * as cdk from "aws-cdk-lib/core";
import * as ecr from "aws-cdk-lib/aws-ecr";
import * as ecs from "aws-cdk-lib/aws-ecs";
import { Construct } from 'constructs';

type EcsStackProps = cdk.StackProps & {
    repository: ecr.Repository;
};

export class EcsStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: EcsStackProps) {
        super(scope, id, props);

        const cluster = new ecs.Cluster(this, "Cluster");

        const taskDefinition = new ecs.FargateTaskDefinition(this, 'TaskDef', {
            runtimePlatform: {
                operatingSystemFamily: ecs.OperatingSystemFamily.LINUX,
                cpuArchitecture: ecs.CpuArchitecture.ARM64,
            },
        });

        taskDefinition.addContainer('Arm64', {
            logging: ecs.LogDriver.awsLogs({ streamPrefix: 'graviton2-on-fargate' }),
            portMappings: [{ containerPort: 80 }],
            image: ecs.ContainerImage.fromEcrRepository(props!.repository, 'server'),
        });

        new ecs.Ec2Service(this, 'Service', {
            cluster,
            taskDefinition,
        });
    }
}