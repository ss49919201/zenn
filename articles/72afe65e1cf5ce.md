---
title: "【SQS/Lambda】SQS + Lambda での部分バッチ応答を試してみる"
emoji: "💿"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["AWS","Lambda","SQS"]
published: false
---

先日、SQSをイベントソースとしたLambdaで部分的なバッチ応答が可能になったことが発表されました。
https://aws.amazon.com/about-aws/whats-new/2021/11/aws-lambda-partial-batch-response-sqs-event-source/?nc1=h_ls
これが出来ずに悩んだ経験があった自分としては、とても嬉しいアップデートです！
**部分バッチ応答**可能になったことで、複数メッセージを処理時に失敗したメッセージのみをキューに再表示できるようになりました。(元々はLambda側で自前実装が必要でした)

これは試さずにはいられないということで、GoでLambda関数を書いて試してみました。

## リソース作成

今回使用するリソースは、

- SQS キュー(イベントソース用キュー、デッドレターキュー)
- Lambda 関数

となります。
CDK で作成していきます。
CDK のバージョンは2.2.0です。

ディレクトリ構成は下記の`tree`コマンドの出力の通りで、今回は一つのスタックに全てのリソースを詰め込んでいます。

```sh
% tree -I "node_modules|cdk.out"
.
├── README.md
├── bin
│   └── cdk-batch-failure.ts
├── cdk.json
├── cli-input
│   ├── README.md
│   └── send-message-batch.json
├── jest.config.js
├── lambda
│   └── go
│       ├── go.mod
│       ├── go.sum
│       └── main.go
├── lib
│   └── cdk-batch-failure-stack.ts
├── package.json
├── test
│   └── cdk-batch-failure.test.ts
├── tsconfig.json
└── yarn.lock

6 directories, 14 files
```

下記が今回作成するスタックです。
イベントソース作成時のパラメータ`reportBatchItemFailures`をtrueにすることで、部分バッチ応答を有効にしています。
また、一度でもLambda関数内で処理に失敗したメッセージはデッドレターキューに移動するようにしています。

```typescript:batch-stack.ts
import { Stack, StackProps, Duration } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as lambda from '@aws-cdk/aws-lambda-go-alpha';
import { SqsEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';

export class CdkBatchFailureStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const deadLetterQueue = new sqs.Queue(this, "deadLetterQueue", {
      queueName: "deadLetterQueue",
    })

    // 一度でもLambda関数内で処理に失敗したメッセージはデッドレターキューに移動
    const queue = new sqs.Queue(this, "queue", {
      queueName: "queue",
      deadLetterQueue: {
        queue: deadLetterQueue,
        maxReceiveCount: 1,
      }
    })

    const source = new SqsEventSource(queue, {
      reportBatchItemFailures: true, // 部分バッチ応答
    })
    new lambda.GoFunction(this, 'RandomResult', {
      entry: 'lambda/go',
      events: [source],
    })
  }
}
```

## 関数作成

Goのバージョンは1.17.3です。

複数件のメッセージがバッチ処理された場合、ランダムで一件のみ失敗するようにします。
handler関数のレスポンスの型は、[ドキュメント](https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html)を参考に独自で宣言しています。

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type SqsBatchResponse struct {
	BatchItemFailures []BatchItemFailure `json:"batchItemFailures"`
}

type BatchItemFailure struct {
	ItemIdentifier string `json:"itemIdentifier"`
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) (res SqsBatchResponse, err error) {
	fmt.Printf("Received %d records\n", len(sqsEvent.Records))

	rand.Seed(time.Now().Unix())
	var randInt int
	if len(sqsEvent.Records) > 1 {
		randInt = rand.Intn(len(sqsEvent.Records))
	}

	for i, message := range sqsEvent.Records {
		fmt.Printf("The message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)

		if len(sqsEvent.Records) > 1 && i == randInt {
			fmt.Printf("Failuer message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)
			res.BatchItemFailures = append(res.BatchItemFailures, BatchItemFailure{message.MessageId})
		}
	}

	// デバッグ用に出力
	b, err := json.Marshal(res)
	if err != nil {
		return
	}
	fmt.Println(string(b))

	return
}

func main() {
	lambda.Start(handler)
}
```

## 動作確認

実際にキューにメッセージを送信してみて、関数を実行させます。
メッセージ送信はCLIでさくっとやります。

```sh
aws sqs send-message-batch --queue-url https://sqs.ap-northeast-1.amazonaws.com/${ACCOUNT_ID}/queue --entries file://cli-input/send-message-batch.json
```

```json:send-message-batch.json
[
    {
        "Id": "1",
        "MessageBody": "1"
    },
    {
        "Id": "2",
        "MessageBody": "2"
    }
]
```

Lambdaのログを確認してみます。


送信した二件のメッセージがバッチ処理され、そのうち一件が失敗メッセージとして返却されています。
続いて同じIDのメッセージがデッドレターキューの方にエンキューされていることを確認します。

想定通り、失敗メッセージがデッドレターキューに入っています。

## さいごに

SQS + Lambda は個人的に大好きな組み合わせなので、今回のアップデートはとても嬉しいです。(二回目)
AWS はかなりハイペースで各サービスの機能の追加や拡張が成されますので、キャッチアップを怠らず日々のエンジニア活動に勤しみたい所存です。

## 参考

https://aws.amazon.com/about-aws/whats-new/2021/11/aws-lambda-partial-batch-response-sqs-event-source/?nc1=h_ls
https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html
https://intro-to-cdk.workshop.aws/the-workshop.html
https://dev.classmethod.jp/articles/sqs-lambda/
