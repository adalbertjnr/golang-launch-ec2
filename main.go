package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	instanceId     string
	instanceIp     string
	inputDelete    string
	getKeyName     string
	err            error
	selectedOs     int
	selectedRegion int
	regionToLaunch string
	osType         string
	osOwner        string
)

func exibeMenu() int {
	var selecaoMenu int
	for {
		fmt.Println("Escolha a imagem para a instância")
		fmt.Println("[1] -> ubuntu")
		fmt.Println("[2] -> awslinux")
		fmt.Println("[3] -> redhat")
		var input string
		fmt.Scan(&input)
		selecaoMenu, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("Digite apenas o número do índice!")
			continue
		}
		if selecaoMenu < 1 || selecaoMenu > 3 {
			fmt.Println("Índice não existe")
			continue
		}
		break
	}
	if selecaoMenu == 1 {
		fmt.Print("Imagem escolhida: ubuntu\n\n")
	} else if selecaoMenu == 2 {
		fmt.Print("Imagem escolhida: awslinux\n\n")
	} else if selecaoMenu == 3 {
		fmt.Print("Imagem escolhida: redhat\n\n")
	} else {
		fmt.Print("Imagem não existe!\n\n")
	}
	return selecaoMenu
}

func exibeRegiao() int {
	var regiaoMenu int
	for {
		fmt.Println("Escolha a região para a instância")
		fmt.Println("[1] -> us-east-1")
		fmt.Println("[2] -> us-west-1")
		fmt.Println("[3] -> sa-east-1")
		var input string
		fmt.Scan(&input)
		regiaoMenu, err = strconv.Atoi(input)
		if err != nil {
			println("Digite apenas o número do índice")
			continue
		}
		if regiaoMenu < 1 || regiaoMenu > 3 {
			fmt.Println("Índice não existe")
			continue
		}
		break
	}
	if regiaoMenu == 1 {
		fmt.Print("Região escolhida us-east-1\n\n")
	} else if regiaoMenu == 2 {
		fmt.Print("Região escolhida us-west-1\n\n")
	} else if regiaoMenu == 3 {
		fmt.Print("Região escolhida sa-east-1\n\n")
	} else {
		fmt.Print("Região não existe\n\n")
	}
	return regiaoMenu
}

func selectOs(os int) {
	switch os {
	case 1:
		osType = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
		osOwner = "099720109477"
	case 2:
		osType = "al2023-ami-2023.0.20230503.0-kernel-6.1-x86_64"
		osOwner = "137112412989"
	case 3:
		osType = "RHEL-9.0.0_HVM-20230313-x86_64-43-Hourly2-GP2"
		osOwner = "309956199498"
	default:
		fmt.Println("Digite o índice ou o nome da imagem da lista!")
	}
}

func selectRegion(reg int) {
	switch reg {
	case 1:
		regionToLaunch = "us-east-1"
	case 2:
		regionToLaunch = "us-west-1"
	case 3:
		regionToLaunch = "sa-east-1"
	default:
		fmt.Println("Digite a região da lista")
	}
}

func getKey() string {
	fmt.Printf("Digite o nome da sua chave para SSH:\n")
	fmt.Scan(&getKeyName)
	fmt.Printf("Chave SSH escolhida: %s\n\n", getKeyName)
	return getKeyName
}

func validateDelete(ctx context.Context, instanceId string) {
	fmt.Printf("Deseja deletar a instância %s? y/n\n", instanceId)
	fmt.Scan(&inputDelete)

	if inputDelete == "y" || inputDelete == "Y" {
		fmt.Printf("Deletando a instância %s\n\n", instanceId)
		deleteInstance(ctx, instanceId)
	} else {
		fmt.Println("A instância não será deletada!")
	}
}

func main() {
	for {
		selectedOs = exibeMenu()
		selectOs(selectedOs)

		selectedRegion = exibeRegiao()
		selectRegion(selectedRegion)

		getKey()

		ctx := context.Background()
		if instanceId, err = createEC2(ctx, regionToLaunch, osType, osOwner, getKeyName); err != nil {
			fmt.Printf("Create EC2 error: %s", err)
			os.Exit(1)
		}
		fmt.Printf("Instance ID: %s\n", instanceId)

		if instanceIp, err = getInstanceIp(ctx, instanceId); err != nil {
			fmt.Printf("EC2 show ip error: %s", err)
			os.Exit(1)
		}

		fmt.Printf("Instance IP: %s\n\n", instanceIp)

		validateDelete(ctx, instanceId)

	}

}

func createEC2(ctx context.Context, region, osType, osOwner, getKeyName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("Unable to load SDK %s", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	keyPair, err := ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{getKeyName},
	})

	if err != nil && strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
		return "", fmt.Errorf("DescribeKeyPair error: %s", err)
	}

	if keyPair == nil || len(keyPair.KeyPairs) == 0 {
		keyPairCreate, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
			KeyName: aws.String(getKeyName),
		})
		if err != nil {
			return "", fmt.Errorf("CreateKeyPair error: %s", err)
		}
		err = os.WriteFile(getKeyName+".pem", []byte(*keyPairCreate.KeyMaterial), 0600)
		if err != nil {
			return "", fmt.Errorf("WriteFile error: %s", err)
		}
	}

	imageOut, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{osType},
			},
		},
		Owners: []string{osOwner},
	})
	if err != nil {
		return "", fmt.Errorf("DescribeImage error: %s", err)
	}

	if len(imageOut.Images) == 0 {
		return "", fmt.Errorf("imageOut.Images is empty")
	}

	ec2, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      imageOut.Images[0].ImageId,
		KeyName:      aws.String(getKeyName),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		return "", fmt.Errorf("EC2 run error: %s", err)
	}

	if len(ec2.Instances) == 0 {
		return "", fmt.Errorf("ec2.instances == 0")
	}

	return *ec2.Instances[0].InstanceId, nil
}

func getInstanceIp(ctx context.Context, instanceId string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("Erro ao carregar SDK para receber o IP. erro: %v", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)

	resp, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return "", fmt.Errorf("Erro ao descrever instâncias. erro: %v", err)
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("Não há instância com o ID: %v", instanceId)
	}
	return *resp.Reservations[0].Instances[0].PublicIpAddress, nil
}

func deleteInstance(ctx context.Context, instanceId string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("Erro ao carregar SDK para deletar Instância. erro: %v", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)
	_, err = ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return fmt.Errorf("Erro ao deletar a instância: %s erro: %s", instanceId, err)
	}
	return nil
}
