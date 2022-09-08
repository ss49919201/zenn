import * as cdk from "aws-cdk-lib/core";
import { Construct } from 'constructs';
import * as ecr from "aws-cdk-lib/aws-ecr";

export class EcrStack extends cdk.Stack {
    public repository: ecr.Repository;
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        // ECRリポジトリを作成
        this.repository = new ecr.Repository(this, "Repository")
    }
}