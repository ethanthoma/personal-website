package main

import (
	"path/filepath"
	"io/ioutil"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type CreateDistributionArgs struct {
	Bucket 	    	      *s3.BucketV2
	BucketName 	      string
	CertificateValidation *acm.CertificateValidation
	DomainName 	      string
}

func CreateDistribution(ctx *pulumi.Context, prefix string, args *CreateDistributionArgs) (*cloudfront.Distribution, error) {
	conf := config.New(ctx, "distribution")
	
	originAccessControlName := conf.Require("originAccessControlName")
	// Creates a new Origin Access Control
	oac, err := cloudfront.NewOriginAccessControl(ctx, prefix + "-" + originAccessControlName, &cloudfront.OriginAccessControlArgs{
		OriginAccessControlOriginType: pulumi.String("s3"),
		SigningBehavior:               pulumi.String("always"),
		SigningProtocol:               pulumi.String("sigv4"),
	})
	if err != nil {
		return nil, err
	}

	functionCodeFilePath := conf.Require("functionCodeFilePath")
	functionName := conf.Require("functionName")
	function, err := CreateCloudfrontFunction(ctx, prefix, &CreateCloudfrontFunctionArgs{
		functionCodeFilePath,
		functionName,
	})
	if err != nil {
		return nil, err
	}

	distributionName := conf.Require("distributionName")
	cdn, err := cloudfront.NewDistribution(ctx, prefix + "-" + distributionName, &cloudfront.DistributionArgs{
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				DomainName: 	       args.Bucket.BucketRegionalDomainName,
				OriginId:              pulumi.String(args.BucketName + "-origin"),
				OriginAccessControlId: oac.ID(),
			},
		},
		Enabled:           pulumi.Bool(true),
		IsIpv6Enabled:     pulumi.Bool(true),
		DefaultRootObject: pulumi.String("index.html"),
		Aliases: pulumi.StringArray{
			//pulumi.String(args.DomainName),
			pulumi.String("www." + args.DomainName),
		},
		Restrictions: &cloudfront.DistributionRestrictionsArgs{
			GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
				RestrictionType: pulumi.String("none"),
			},
		},
		DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
			AllowedMethods: pulumi.ToStringArray([]string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"}),
			CachedMethods: pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
			TargetOriginId: pulumi.String(args.BucketName + "-origin"),
			ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
				QueryString: pulumi.Bool(false),
				Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
					Forward: pulumi.String("none"),
				},
			},
			ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
			MinTtl:               pulumi.Int(0),
			DefaultTtl:           pulumi.Int(0), //3600
			MaxTtl:               pulumi.Int(0), //86400
			FunctionAssociations: cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArray{
				cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArgs{
					FunctionArn: function.Arn,
					EventType:   pulumi.String("viewer-request"),
				},
			},
		},
		ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn: args.CertificateValidation.CertificateArn,
			SslSupportMethod:  pulumi.String("sni-only"),
		},
	})
	if err != nil {
		return nil, err
	}

	return cdn, nil
}

type CreateCloudfrontFunctionArgs struct {
	Filename    string
	FunctionName string
}

func CreateCloudfrontFunction(ctx *pulumi.Context, prefix string, args *CreateCloudfrontFunctionArgs) (*cloudfront.Function, error) {
	functionCode, err := readCloudFrontFunctionCode(args.Filename)
	if err != nil {
		return nil, err
	}

	cloudfrontFunction, err := cloudfront.NewFunction(ctx, prefix + "-" + args.FunctionName, &cloudfront.FunctionArgs{
		Code:    pulumi.String(functionCode),
		Publish: pulumi.Bool(true),
		Runtime: pulumi.String("cloudfront-js-1.0"),
	})
	if err != nil {
		return nil, err
	}

	return cloudfrontFunction, nil
}

func readCloudFrontFunctionCode(filename string) (string, error) {
	// Get the absolute file path
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}

	// Read the contents of the JavaScript file
	bytes, err := ioutil.ReadFile(absFilePath)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

