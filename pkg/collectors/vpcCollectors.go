package collectors

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

const (
	QUOTA_CODE_IPV4_BLOCKS_PER_VPC string = "L-83CA0A9D"
	SERVICE_CODE_VPC               string = "vpc"
)

type Ipv4BlocksPerVPCCollector struct {
	ServiceQuotaClient *servicequotas.Client
	Ec2Client          *ec2.Client
	VpcID              string
}

func (c Ipv4BlocksPerVPCCollector) Quota() (float64, error) {
	return GetQuotaValue(c.ServiceQuotaClient, SERVICE_CODE_VPC, QUOTA_CODE_IPV4_BLOCKS_PER_VPC)
}

func (c Ipv4BlocksPerVPCCollector) Usage() (float64, error) {
	descVpcOut, err := c.Ec2Client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
		DryRun: aws.Bool(false),
		VpcIds: []string{c.VpcID},
	})

	if err != nil {
		return 0, err
	}
	if len(descVpcOut.Vpcs) != 1 {
		return 0, errors.New("Unexcpected number of VPCs returned")
	}

	return float64(len(descVpcOut.Vpcs[0].CidrBlockAssociationSet)), nil
}

func (c Ipv4BlocksPerVPCCollector) Name() string {
	return "ipv4_blocks_per_vpc_" + c.VpcID
}
