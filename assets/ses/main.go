package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

func main() {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("ap-northeast-1")}))
	svc := ses.New(sess)

	input := new(ses.SendEmailInput)

	input.SetDestination(&ses.Destination{
		ToAddresses: []*string{
			aws.String("recipient@example.com"),
		},
	})

	input.SetMessage(&ses.Message{
		Body: &ses.Body{
			Text: &ses.Content{
				Data: aws.String("ぼでぃ"),
			},
		},
		Subject: &ses.Content{
			Data: aws.String("さぶじぇくと"),
		},
	})

	input.SetSource("そうしんしゃ<sender@example.com>")

	_, err := svc.SendEmail(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Success")
}
