---
title: "【AWS】AWS CLI を使って App Runner にデプロイ"
emoji: "📚"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["AWS","apprunner"]
published: true
---

先日、AWS App Runner がローンチされました。
https://aws.amazon.com/jp/blogs/news/introducing-aws-app-runner/

コンテナ上で動作するアプリケーションを簡単にデプロイできるフルマネージドサービスとのことです。
ドキュメントや関連記事を読んでいると自分も使ってみたくてウズウズしてきたので、実際に手を動かしてみました。
なお今回はタイトルの通り AWS CLI を使ってリソースの作成、デプロイを行いました。

:::message
AWS CLI は現時点では v1 のみが App Runner をサポートしています。
v2 のサポートはもう時期にリリースされるとのことです。(Twitterで教えてもらいました)
https://twitter.com/toricls/status/1395982897320890369?s=20

2021/05/30 追記
v2 でのサポートがリリースされました！(またまたTwitterで教えてもらいました)
https://twitter.com/toricls/status/1397359793988194305
:::

## ソースコード

今回デプロイするWebアプリケーションのソースコードです。
Goならたったこれだけの記述でWebサーバが立てられちゃうのです。

```go
package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## イメージの作成

App Runner で利用するソースとして GitHub リポジトリか、ECR のレジストリのどちらかを一つを選択できるのですが、今回は AWS サービスで全て完結させたかったので ECR を選択しました。
よってリポジトリに配置するためのイメージを作成する必要があります。
こちらがイメージを作成するためのDockerfileです。

```docker
FROM golang:1.16-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go build -o /go/bin/sample

FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/sample /
CMD ["/sample"]
```

ファイル構成はこちらを参考させていただきました。
https://zenn.dev/komisan19/articles/45b00df6bfe7ad

## リポジトリの作成

ECR にリポジトリを作成していきます。
今回作成するリポジトリは`sample-repo`と名付けます。
リージョン名は後のコマンドでも使うので、シェル変数として設定しておきます。

```sh
$ AWS_REGION=ap-northeast-1 
$ aws ecr create-repository --repository-name sample-repo --region ${AWS_REGION}
```

無事作成されるとリポジトリの情報が出力されます。
出力の`registryId`と`repositoryUri`は、リージョン名と同様に後で使うので変数として設定します。

```sh
$ REGISTORY_ID=xxxxxxxxxxx
$ REPOSITORY_URI=xxxxxxxxxxx.dkr.ecr.ap-northeast-1.amazonaws.com/sample-repo
```

## イメージのビルド

用意したGoのソースコードとDockerfileからイメージをビルドします。

```sh
$ docker build -t sample .
```

一応ローカルで動作確認をしたい場合は、`docker run`して`localhost:8080/hello`にアクセスします。

```sh
$ docker run -t -i -p 8080:8080 --name sample sample

# 別タブ
$ curl "localhost:8080/hello"
Hello, world!
```

## イメージのPush

ECR のリポジトリにPushする為に、先ほどビルドしたイメージにタグを付けます。

```sh
$ docker tag sample ${REPOSITORY_URI}
```

ECRにログインし、イメージをプッシュします。

```sh
$ aws ecr get-login-password | docker login --username AWS --password-stdin ${REGISTRY_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com
$ docker push ${REPOSITORY_URI}
```
## IAM ロールの作成

App Runner の Service の為のロールを作成します。
実はCLIでロールを作るのは初めてでした。

まずは信頼ポリシーをJSONで作成します。

```json:role-trust-policy.json
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Principal": {
                "Service": "build.apprunner.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
```

ロールを作成します。
ロールは`AppRunnerECRAccessRole`と名付けます。
`--assume-role-policy-document`に先ほど作成したJSONファイルを指定します。

```sh
$ aws iam create-role --role-name AppRunnerECRAccessRole --assume-role-policy-document file://role-trust-policy.json
```

ここで出力されたロールのARNは後で必要なので、控えておきましょう。

作成したロールに ECR へのアクセス権限をアタッチします。

```sh
$ aws iam attach-role-policy --role-name AppRunnerECRAccessRole --policy-arn 'arn:aws:iam::aws:policy/service-role/AWSAppRunnerServicePolicyForECRAccess'
```

## App Runner Service の作成

いよいよ Service の作成です。
`create-service`のSynopsisがこちらです。

```
  create-service
--service-name <value>
--source-configuration <value>
[--instance-configuration <value>]
[--tags <value>]
[--encryption-configuration <value>]
[--health-check-configuration <value>]
[--auto-scaling-configuration-arn <value>]
[--cli-input-json <value>]
[--generate-cli-skeleton <value>]
```
https://docs.aws.amazon.com/cli/latest/reference/apprunner/create-service.html

ここで`--source-configuration` に構造化した設定値を渡してあげる必要があるのですが、途中でtypoしまくって断念しました。
ということでスケルトンから、設定値を定義したJSONを作ります。

```sh
$ aws apprunner create-service --generate-cli-skeleton input > app-runner.json
```
完成したJSONがこちらです。

```json:app-runner.json
{
    "ServiceName": "sample-app",
    "SourceConfiguration": {
        "ImageRepository": {
            "ImageIdentifier": "xxxxxxxxxxx.dkr.ecr.ap-northeast-1.amazonaws.com/sample-repo:latest",
            "ImageRepositoryType": "ECR",
            "ImageConfiguration": {
                "Port": "8080"
            }
        },
        "AutoDeploymentsEnabled": true,
        "AuthenticationConfiguration": {
            "AccessRoleArn": "arn:aws:iam::xxxxxxxxxxx:role/AppRunnerECRAccessRole"
        }
    }
}
```

設定値はほとんどデフォルトにしています。
スペックやタグ、環境変数(一番嬉しい)なども設定可能です。
Dockerイメージ作成時にもさらっと書きましたが、アプリケーションのソースとして GitHub か ECR を選択できます。
GitHubを選択した場合には`ImageRepository`ではなく`CodeRepository`を指定する必要があります。
その他の設定値の意味はざっと箇条書きしておきます。

- `ImageRepositoryType`->Privateリポジトリの場合は`ECR`を、Publicリポジトリの場合は`ECR_PUBLIC`を指定
- `ImageIdentifier`->イメージのURIを指定 
- `Port`->アプリケーションのポート番号を指定
- `AutoDeploymentsEnabled`->自動デプロイのON,OFFを指定
- `AccessRoleArn`->IAM のARNを指定(先ほど作成したロールのARN)

作成したJSONを基にサービスを作成します。

```sh
$ aws apprunner create-service --cli-input-json file://app-runner.json
```
Service の情報が出力されていれば作成が始まってい5るはずなので5分ほど待ちましょう☕️

5分経ったので出力の`ServiceUrl`にアクセスしてみます。ちなみにSSL通信です。(優秀)

```sh
$ curl https://...
404 page not found
```

!?
`/hello`でした。忘れてました。

```sh
$ curl https://.../hello
Hello, world!
```

嬉しい！

## App Runner Service の削除

お掃除しておきます。
Service 作成時の出力の`ServiceArn`を指定して削除します。

```sh
$ aws apprunner delete-service --service-arn arn:aws:apprunner:ap-northeast-1:xxxxxxxxxxxx:service/sample-app/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## 所感

AWS App Runner は始まったばかりのサービスでありまだまだ不十分な機能もあるようですが、GitHubのリポジトリで積極的に要望を集めたりしていて今後が非常に楽しみなサービスです。
https://twitter.com/toricls/status/1395993104964997125?s=20
新しい技術を試してみるのは興奮しますね。
