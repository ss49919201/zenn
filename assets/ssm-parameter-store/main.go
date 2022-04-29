package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func main() {
	fetchAndPrint()
}

func put() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := ssm.NewFromConfig(cfg)

	for i := 11; i <= 20; i++ {
		_, err := svc.PutParameter(ctx, &ssm.PutParameterInput{
			Name:  aws.String("/beats/" + strconv.Itoa(i)),
			Value: aws.String(strconv.Itoa(i)),
			Type:  types.ParameterTypeString,
		})
		if err != nil {
			log.Fatalf("failed to put parameter, %v", err)
		}
	}
}

func fetchAndPrint() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := ssm.NewFromConfig(cfg)

	var params []types.Parameter
	var nextToken *string
	for {
		input := &ssm.GetParametersByPathInput{
			Path:      aws.String("/beats/"),
			NextToken: nextToken,
		}

		op, err := svc.GetParametersByPath(ctx, input)
		if err != nil {
			log.Fatalf("failed to get parameter, %v", err)
		}

		params = append(params, op.Parameters...)

		nextToken = op.NextToken
		if op.NextToken == nil {
			break
		}
	}

	for _, v := range params {
		os.Setenv(*v.Name, *v.Value)
	}

	for _, v := range params {
		fmt.Println(os.Getenv(*v.Name))
	}
}
