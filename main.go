package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/openshift/privatelink-infrastructure-monitor/pkg/collectors"
)

func main() {

	var hostedZoneID = flag.String("hosted-zone-id", "", "ID of a checked hosted zone")
	var region = flag.String("region", "", "Aws Region to run this in")
	flag.Parse()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg.Region = *region

	if err != nil {
		panic("config error, " + err.Error())
	}
	ec2Client := ec2.NewFromConfig(cfg)
	serviceQuotaClient := servicequotas.NewFromConfig(cfg)
	r53Client := route53.NewFromConfig(cfg)
	allCollectors := []collectors.QuotaCollector{}

	allCollectors = append(allCollectors, &collectors.TransitGatewaysPerAcctCollector{
		ServiceQuotaClient: serviceQuotaClient,
		Ec2Client:          ec2Client,
	})
	allCollectors = append(allCollectors, &collectors.Route53RecordsPerHostedZoneCollector{
		ServiceQuotaClient: serviceQuotaClient,
		R53Client:          r53Client,
		HostedZoneID:       *hostedZoneID,
	})
	allCollectors = append(allCollectors, &collectors.Ipv4BlocksPerVPCCollector{
		ServiceQuotaClient: serviceQuotaClient,
		Ec2Client:          ec2Client,
		VpcID:              "vpc-5bc7103e",
	})

	for _, col := range allCollectors {
		quota, err := col.Quota()
		if err != nil {
			panic("Could not get quota: " + err.Error())
		}
		usage, err := col.Usage()
		if err != nil {
			panic("Could not get usage: " + err.Error())
		}

		fmt.Printf("%s\n\tQuota: %.2f\n\tUsage: %.2f\n", col.Name(), quota, usage)
	}

}
