package aws

import(
	// "fmt"
	"log"
	"context"
	// "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var (
    EC2Client *ec2.Client
    R53Client *route53.Client
)

func InitAWS(){
	cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    // Create an Amazon S3 service client
	EC2Client = ec2.NewFromConfig(cfg)
    R53Client = route53.NewFromConfig(cfg)

}