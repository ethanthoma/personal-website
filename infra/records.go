package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudfront"

	"github.com/pulumi/pulumi-cloudflare/sdk/v5/go/cloudflare"
)

type CreateCertificateRecordArgs struct {
	CloudflareZoneId string
	Certificate 	 *acm.Certificate
	DomainName       string
}

func CreateCertificateRecord(ctx *pulumi.Context, prefix string, args *CreateCertificateRecordArgs) error {
	conf := config.New(ctx, "records")

	certificateRecordName := conf.Require("certificateRecordName")

	_, err := cloudflare.NewRecord(ctx, prefix + "-" + certificateRecordName + "-" + args.DomainName, &cloudflare.RecordArgs{
		ZoneId: pulumi.String(args.CloudflareZoneId),
		Name: args.Certificate.DomainValidationOptions.ApplyT(func(domainValidationOptions []acm.CertificateDomainValidationOption) string {
			return *domainValidationOptions[0].ResourceRecordName
		}).(pulumi.StringInput),
		Value: args.Certificate.DomainValidationOptions.ApplyT(func(domainValidationOptions []acm.CertificateDomainValidationOption) string {
			return *domainValidationOptions[0].ResourceRecordValue
		}).(pulumi.StringInput),
		Type: args.Certificate.DomainValidationOptions.ApplyT(func(domainValidationOptions []acm.CertificateDomainValidationOption) string {
			return *domainValidationOptions[0].ResourceRecordType
		}).(pulumi.StringInput),
	})
	if err != nil {
		return err
	}

	return nil
}

type SetupCloudflareDNSArgs struct {
	CloudflareZoneId       string
	CloudfrontDistribution *cloudfront.Distribution
	DomainName             string
}

func SetupCloudflareDNS(ctx *pulumi.Context, prefix string, args *SetupCloudflareDNSArgs) error {
	conf := config.New(ctx, "records")
	
	domainCertificateName := conf.Require("domainCertificateName")
	_, err := cloudflare.NewRecord(ctx, prefix + "-" + domainCertificateName, &cloudflare.RecordArgs{
		Name:    pulumi.String(args.DomainName),
		Proxied: pulumi.Bool(true),
		Type:    pulumi.String("A"),
		Value:   pulumi.String("192.0.2.1"),
		ZoneId:  pulumi.String(args.CloudflareZoneId),
	})
	if err != nil {
		return err
	}

	subdomainCertificateName := conf.Require("subdomainCertificateName")
	_, err = cloudflare.NewRecord(ctx, prefix + "-" + subdomainCertificateName, &cloudflare.RecordArgs{
		Name:   pulumi.String("www." + args.DomainName),
		Type:   pulumi.String("CNAME"),
		Value:  args.CloudfrontDistribution.DomainName,
		ZoneId: pulumi.String(args.CloudflareZoneId),
	})

	redirectRuleName := conf.Require("redirectRuleName")
	_, err = cloudflare.NewPageRule(ctx, prefix + "-" + redirectRuleName, &cloudflare.PageRuleArgs{
		ZoneId:   pulumi.String(args.CloudflareZoneId),
		Priority: pulumi.Int(1),
		Target:   pulumi.String(args.DomainName) + "/*",
		Actions:  &cloudflare.PageRuleActionsArgs{
			ForwardingUrl: &cloudflare.PageRuleActionsForwardingUrlArgs{
				StatusCode: pulumi.Int(301),
				Url:        pulumi.String("https://www." + args.DomainName + "/$1"),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

