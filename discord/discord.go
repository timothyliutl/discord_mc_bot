package discord

import(
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"mc_bot/aws"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
"github.com/aws/aws-sdk-go-v2/service/route53"
"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func Server(token string) (*discordgo.Session, error){
	dg, err := discordgo.New("Bot " + token)
	dg.AddHandler(messageListener)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return nil, err
    }
	return dg, err

}

func messageListener(s *discordgo.Session, m *discordgo.MessageCreate){
	if m.Author.ID == s.State.User.ID {
        return
    }
	fmt.Println(m.Content)

	if m.Content == "!status"{
		_, err := s.ChannelMessageSend(m.ChannelID, getServerStatus())
		if err!=nil{
			fmt.Println("Error sending message: ", err)
		}
	}

	if m.Content == "!start"{
		updateInstanceIP()
	}
}




func getServerStatus() string{
	// err := godotenv.Load()
	instanceID := os.Getenv("EC2_INSTANCE_ID")
	result, err := aws.EC2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
        InstanceIds: []string{instanceID},
    })
	if err != nil{
    fmt.Println("Error has occurred:", err)
	}

	instance := result.Reservations[0].Instances[0]
	status := instance.State.Name
	message := fmt.Sprintf("Instance status: %s", status)

	return message
}

func updateInstanceIP() {
	instanceID := os.Getenv("EC2_INSTANCE_ID")
	hostedZoneID := os.Getenv("ROUTE53_HOSTED_ZONE_ID")
	recordName := os.Getenv("ROUTE53_RECORD_NAME")

	result, err := aws.EC2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
        InstanceIds: []string{instanceID},
    })

	if err != nil{
		fmt.Println("Error has occurred:", err)
	}
	instance := result.Reservations[0].Instances[0]
	publicIP := *instance.PublicIpAddress
	if publicIP == ""{
		fmt.Println("Error: Public IP Address empty")
		return
	}
		// update route 53
		_, err = aws.R53Client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: sdkaws.String(hostedZoneID),
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action: types.ChangeActionUpsert,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: sdkaws.String(recordName),
							Type: types.RRTypeA,
							TTL:  sdkaws.Int64(60),
							ResourceRecords: []types.ResourceRecord{
								{
									Value: sdkaws.String(publicIP),
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			fmt.Println("Error updating Route 53:", err)
			return
		}
		fmt.Println("Updated Route 53 record to:", publicIP)

}