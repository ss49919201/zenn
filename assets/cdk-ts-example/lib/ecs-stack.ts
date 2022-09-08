import * as cdk from "aws-cdk-lib";
import * as ecr from "aws-cdk-lib/aws-ecr";
import * as ecs from "aws-cdk-lib/aws-ecs";
import { Construct } from 'constructs';

type EcsStackProps = cdk.StackProps & {
    repository: ecr.Repository;
};

export class EcsStack extends cdk.Stack {
    private taskDefinition(arch: ecs.CpuArchitecture): ecs.TaskDefinition {
        return new ecs.FargateTaskDefinition(this, 'TaskDef', {
            runtimePlatform: {
                cpuArchitecture: arch
            },
        });
    }

    constructor(scope: Construct, id: string, props?: EcsStackProps) {
        super(scope, id, props);

        const cluster = new ecs.Cluster(this, "Cluster");

        const taskDefinition = this.taskDefinition(ecs.CpuArchitecture.X86_64);

        taskDefinition.addContainer('Container', {
            logging: ecs.LogDriver.awsLogs({ streamPrefix: 'graviton2-on-fargate' }),
            portMappings: [{ containerPort: 80 }],
            image: ecs.ContainerImage.fromEcrRepository(props!.repository, 'server'),
        });

        new ecs.FargateService(this, 'Service', {
            cluster,
            taskDefinition,
            assignPublicIp: true, // ECR のPullが面倒なので、パブリックIPを割り当てる
        });
    }
}