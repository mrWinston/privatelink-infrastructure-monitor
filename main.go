package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/openshift/privatelink-infrastructure-monitor/pkg/collectors"
)

func main() {

	var hostedZoneId = flag.String("hosted-zone-id", "", "ID of a checked hosted zone")
	var region = flag.String("region", "", "Aws Region to run this in")
	var vpcId = flag.String("vpc-id", "", "ID of a checked vpc")

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

	routeTables, err := ec2Client.DescribeRouteTables(context.TODO(), &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{{
			Name:   aws.String("vpc-id"),
			Values: []string{*vpcId},
		}},
	})
	if err != nil {
		fmt.Println("Could not retrieve route tables for VPC, can not monitor the quotas.")
	} else {
		for i, _ := range routeTables.RouteTables {
			routeTable := routeTables.RouteTables[i]
			allCollectors = append(allCollectors, &collectors.RoutesPerRouteTableCollector{
				ServiceQuotaClient: serviceQuotaClient,
				Ec2Client:          ec2Client,
				RouteTableID:       *routeTable.RouteTableId,
			})
		}
	}

	allCollectors = append(allCollectors, &collectors.TransitGatewaysPerAcctCollector{
		ServiceQuotaClient: serviceQuotaClient,
		Ec2Client:          ec2Client,
	})
	allCollectors = append(allCollectors, &collectors.Route53RecordsPerHostedZoneCollector{
		ServiceQuotaClient: serviceQuotaClient,
		R53Client:          r53Client,
		HostedZoneID:       *hostedZoneId,
	})
	allCollectors = append(allCollectors, &collectors.Ipv4BlocksPerVPCCollector{
		ServiceQuotaClient: serviceQuotaClient,
		Ec2Client:          ec2Client,
		VpcID:              *vpcId,
	})
	allCollectors = append(allCollectors, &collectors.RouteTablesPerVPCCollector{
		ServiceQuotaClient: serviceQuotaClient,
		Ec2Client:          ec2Client,
		VpcID:              *vpcId,
	})
	allCollectors = append(allCollectors, &collectors.InterfaceVpcEndpointsPerVpc{
		ServiceQuotaClient: serviceQuotaClient,
		Ec2Client:          ec2Client,
		VpcID:              *vpcId,
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
