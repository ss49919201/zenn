package main

import (
	"context"
	"fmt"
	"log"
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

	var nextToken *string
	for {
		op, err := svc.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:      aws.String("/beats/"),
			NextToken: nextToken,
		})
		if err != nil {
			log.Fatalf("failed to get parameter, %v", err)
		}

		for _, v := range op.Parameters {
			fmt.Printf("Name: %#v\n", *v.Name)
			fmt.Printf("Value: %#v\n", *v.Value)
		}

		if op.NextToken == nil {
			break
		} else {
			nextToken = op.NextToken
		}
	}
}
