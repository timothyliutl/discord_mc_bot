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
"errors"
"time"
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
		message := fmt.Sprintf("Instance status: %s", getServerStatus())
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err!=nil{
			fmt.Println("Error sending message: ", err)
			return
		}
	}

	if m.Content == "!start"{
		if getServerStatus() == "stopping"{
			s.ChannelMessageSend(m.ChannelID, "Stopping server, try again later")
			return
		}

		if getServerStatus() == "running"{
			s.ChannelMessageSend(m.ChannelID, "Server is already running")
			return
		}

		err1 := startServer()
		if err1 != nil{
			message := fmt.Sprintf("Error starting instance: %v", err1)
			s.ChannelMessageSend(m.ChannelID, message)
			return
		}

		
		s.ChannelMessageSend(m.ChannelID, "Server is starting. Waiting for public IP...")
		err2 := waitForInstanceRunning()
		if err2 != nil {
			message := fmt.Sprintf("Error waiting for instance: %v", err2)
			s.ChannelMessageSend(m.ChannelID, message)
			return
		}

		
		err3 := updateInstanceIP()
		if err3 != nil {
			message := fmt.Sprintf("Error updating IP: %v", err3)
			s.ChannelMessageSend(m.ChannelID, message)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Server started and DNS updated.")
		
	}

	if m.Content == "!stop"{
		if getServerStatus() == "stopped"{
			s.ChannelMessageSend(m.ChannelID, "Already stopped")
			return
		}
		if getServerStatus() == "stopping"{
			s.ChannelMessageSend(m.ChannelID, "Already stopping...")
			return
		}
		err1 := stopServer()
		if err1!=nil{
			fmt.Println("Error sending message: ", err1)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Stopping server...")
		err2 := waitForInstanceStopped()
		if err2 != nil {
			message := fmt.Sprintf("Error waiting for shutdown: %v", err2)
			s.ChannelMessageSend(m.ChannelID, message)
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Server has fully stopped.")

	}
}

func startServer() error{
	instanceID := os.Getenv("EC2_INSTANCE_ID")

	_, err := aws.EC2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
        InstanceIds: []string{instanceID},
    })

	if err != nil{
		fmt.Println("Error has occurred:", err)
		return err
	}
	fmt.Println("Started instance")
	return nil
}


func stopServer() error{
	instanceID := os.Getenv("EC2_INSTANCE_ID")

	_, err := aws.EC2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
        InstanceIds: []string{instanceID},
    })

	if err != nil{
		fmt.Println("Error has occurred:", err)
		return err
	}
	fmt.Println("Stopped instance")
	return nil
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

	return string(status)
}

func updateInstanceIP() error{
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
	if instance.PublicIpAddress == nil {
		fmt.Println("Error: Public IP Address empty")
		return errors.New("public IP address not assigned yet")
	}
	publicIP := *instance.PublicIpAddress
	if publicIP == ""{
		fmt.Println("Error: Public IP Address empty")
		return errors.New("public IP address empty")
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
			return err
		}
		fmt.Println("Updated Route 53 record to:", publicIP)
		return nil

}



func waitForInstanceRunning() error {
	instanceID := os.Getenv("EC2_INSTANCE_ID")
	waiter := ec2.NewInstanceRunningWaiter(aws.EC2Client)
	err := waiter.Wait(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return err
	}
	return nil
}

func waitForInstanceStopped() error {
	instanceID := os.Getenv("EC2_INSTANCE_ID")
	waiter := ec2.NewInstanceStoppedWaiter(aws.EC2Client)
	err := waiter.Wait(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return err
	}
	return nil
}