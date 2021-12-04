---
title: "【SQS/Lambda】SQS + Lambda での部分バッチ応答が嬉しすぎた"
emoji: "💿"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["AWS","Lambda","SQS"]
published: false
---

先日、SQSをイベントソースとしたLambdaで部分的なバッチ応答が可能になったことが発表されました。
https://aws.amazon.com/about-aws/whats-new/2021/11/aws-lambda-partial-batch-response-sqs-event-source/?nc1=h_ls
これが出来ずに悩んだ経験があった自分としては、とても嬉しいアップデートです！

実は正直なところ、

> AWS Lambda が、イベントソースとしての SQS への部分バッチ応答のサポートを開始

というタイトルを見ただけでは何ができるようになったのかは理解できなかったのですが、本文とドキュメントを読んで意味が分かってテンション爆上がりしました。（2021/12/22時点では、Lambdaの日本語ドキュメントに部分バッチ応答に関する記載はありません。）

**部分バッチ応答**可能になって何ができるのかを簡単にまとてめみると、

- LambdaからSQSへの結果返却時に、成功したメッセージと失敗したメッセージをまとめて返却できるようになった

これは試さずにはいられないということで、Goでやってみました。

## 参考

https://aws.amazon.com/about-aws/whats-new/2021/11/aws-lambda-partial-batch-response-sqs-event-source/?nc1=h_ls
https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html
