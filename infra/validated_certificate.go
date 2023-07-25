package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type CreateValidatedCertificateArgs struct {
	CloudflareZoneId string
	DomainName       string
}

func CreateValidatedCertificate(ctx *pulumi.Context, prefix string, args *CreateValidatedCertificateArgs) (*acm.CertificateValidation, error) {
	conf := config.New(ctx, "validatedCertificate")

	providerName := conf.Require("providerName")
	provider, err := aws.NewProvider(ctx, prefix + "-" + providerName, &aws.ProviderArgs{Region: pulumi.String("us-east-1")})
	if err != nil {
		return nil, err
	}

	certificateName := conf.Require("certificateName")
	certificate, err := acm.NewCertificate(ctx, prefix + "-" + certificateName + "-" + args.DomainName, &acm.CertificateArgs{
		DomainName:       pulumi.String(args.DomainName),
		ValidationMethod: pulumi.String("DNS"),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	err = CreateCertificateRecord(ctx, prefix, &CreateCertificateRecordArgs{
		args.CloudflareZoneId,
		certificate,
		args.DomainName,
	})
	if err != nil {
		return nil, err
	}

	certificateValidationName := conf.Require("certificateValidationName")
	certResource, err := acm.NewCertificateValidation(ctx, prefix + "-" + certificateValidationName + "-" + args.DomainName, &acm.CertificateValidationArgs{
		CertificateArn: certificate.Arn,
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	return certResource, nil
}

