package main

import (
	"fmt"
	"path/filepath"
	"io/fs"
	"mime"
	"path"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"

	"github.com/pulumi/pulumi-cloudflare/sdk/v5/go/cloudflare"
	
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		bucketName := cfg.Require("bucketName")
		domainName := cfg.Require("domainName")
		cloudflareZoneId := cfg.Require("cloudflareZoneId")

		bucket, err := createBucket(ctx, bucketName)
		if err != nil {
			return err
		}

		err = uploadDirectoryContent(ctx, bucket, "../dist")
		if err != nil {
			return err
		}

		provider, err := aws.NewProvider(ctx, "aws-provider-us-east-1", &aws.ProviderArgs{Region: pulumi.String("us-east-1")})
		if err != nil {
			return err
		}

		certResource, err := createValidatedCertificate(ctx, provider, "www." + domainName, cloudflareZoneId)
		if err != nil {
			return err
		}

		cdn, err := createDistribution(ctx, bucketName, certResource, bucket, domainName)
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "domain", &cloudflare.RecordArgs{
			Name:    pulumi.String(domainName),
			Proxied: pulumi.Bool(true),
			Type:    pulumi.String("A"),
			Value:   pulumi.String("192.0.2.1"),
			ZoneId:  pulumi.String(cloudflareZoneId),
		})
		if err != nil {
			return err
		}

		_, err = cloudflare.NewRecord(ctx, "subdomain", &cloudflare.RecordArgs{
			Name:   pulumi.String("www." + domainName),
			Type:   pulumi.String("CNAME"),
			Value:  cdn.DomainName,
			ZoneId: pulumi.String(cloudflareZoneId),
		})

		_, err = cloudflare.NewPageRule(ctx, "redirectRule", &cloudflare.PageRuleArgs{
			ZoneId:   pulumi.String(cloudflareZoneId),
			Priority: pulumi.Int(1),
			Target:   pulumi.String(domainName) + "/*",
			Actions:  &cloudflare.PageRuleActionsArgs{
				ForwardingUrl: &cloudflare.PageRuleActionsForwardingUrlArgs{
					StatusCode: pulumi.Int(301),
					Url:        pulumi.String("https://www." + domainName + "/$1"),
				},
			},
		})
		if err != nil {
			return err
		}

		err = createBucketPolicy(ctx, bucket, cdn)
		if err != nil {
			return err
		}

		ctx.Export("awsCloudfront", cdn)
		ctx.Export("awsS3", bucket)
		return nil
	})
}

func createBucket (ctx *pulumi.Context, bucketId string) (*s3.BucketV2, error) {
	bucket, err := s3.NewBucketV2(ctx, bucketId, nil)
	if err != nil {
		return nil, err
	}
	
	return bucket, nil
}

func uploadDirectoryContent (ctx *pulumi.Context, bucket *s3.BucketV2, directoryPath string) error {
	err := filepath.Walk(directoryPath, func(name string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			rel, err := filepath.Rel(directoryPath, name)
			if err != nil {
				return err
			}

			if _, err := s3.NewBucketObject(ctx, rel, &s3.BucketObjectArgs{
				Bucket:      bucket.ID(),                                         // reference to the s3.Bucket object
				Source:      pulumi.NewFileAsset(name),                           // use FileAsset to point to a file
				ContentType: pulumi.String(mime.TypeByExtension(path.Ext(name))), // set the MIME type of the file
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func createValidatedCertificate(ctx *pulumi.Context, provider *aws.Provider, domainName string, cloudflareZoneId string) (*acm.CertificateValidation, error) {
	cert, err := acm.NewCertificate(ctx, "cert-" + domainName, &acm.CertificateArgs{
		DomainName:       pulumi.String(domainName),
		ValidationMethod: pulumi.String("DNS"),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	_, err = cloudflare.NewRecord(ctx, "acm-" + domainName, &cloudflare.RecordArgs{
		ZoneId: pulumi.String(cloudflareZoneId),
		Name: cert.DomainValidationOptions.ApplyT(func(domainValidationOptions []acm.CertificateDomainValidationOption) string {
			return *domainValidationOptions[0].ResourceRecordName
		}).(pulumi.StringInput),
		Value: cert.DomainValidationOptions.ApplyT(func(domainValidationOptions []acm.CertificateDomainValidationOption) string {
			return *domainValidationOptions[0].ResourceRecordValue
		}).(pulumi.StringInput),
		Type: cert.DomainValidationOptions.ApplyT(func(domainValidationOptions []acm.CertificateDomainValidationOption) string {
			return *domainValidationOptions[0].ResourceRecordType
		}).(pulumi.StringInput),
	})
	if err != nil {
		return nil, err
	}

	certResource, err := acm.NewCertificateValidation(ctx, "certResource-" + domainName, &acm.CertificateValidationArgs{
		CertificateArn: cert.Arn,
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	ctx.Export("awsAcm-" + domainName, cert)

	return certResource, nil
}

func createDistribution(ctx *pulumi.Context, bucketName string, certResource *acm.CertificateValidation, bucket *s3.BucketV2, domainName string) (*cloudfront.Distribution, error) {
	//Creates a new Origin Access Control
	oac, err := cloudfront.NewOriginAccessControl(ctx, "website-OAC", &cloudfront.OriginAccessControlArgs{
		OriginAccessControlOriginType: pulumi.String("s3"),
		SigningBehavior:               pulumi.String("always"),
		SigningProtocol:               pulumi.String("sigv4"),
	})
	if err != nil {
		return nil, err
	}

	function, err := createCloudfrontFunction(ctx)
	if err != nil {
		return nil, err
	}

	cdn, err := cloudfront.NewDistribution(ctx, "dist", &cloudfront.DistributionArgs{
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				DomainName: bucket.BucketRegionalDomainName,
				OriginId:   pulumi.String(fmt.Sprintf("S3-%v", bucketName)),
				OriginAccessControlId: oac.ID(),
			},
		},
		Enabled:           pulumi.Bool(true),
		IsIpv6Enabled:     pulumi.Bool(true),
		DefaultRootObject: pulumi.String("index.html"),
		Aliases: pulumi.StringArray{
			pulumi.String(domainName),
			pulumi.String("www." + domainName),
		},
		Restrictions: &cloudfront.DistributionRestrictionsArgs{
			GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
				RestrictionType: pulumi.String("none"),
			},
		},
		DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
			AllowedMethods: pulumi.ToStringArray([]string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"}),
			CachedMethods: pulumi.ToStringArray([]string{"GET", "HEAD", "OPTIONS"}),
			TargetOriginId: pulumi.String(fmt.Sprintf("S3-%v", bucketName)),
			ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
				QueryString: pulumi.Bool(false),
				Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
					Forward: pulumi.String("none"),
				},
			},
			ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
			MinTtl:               pulumi.Int(0),
			DefaultTtl:           pulumi.Int(0), //3600
			MaxTtl:               pulumi.Int(0),//86400
			FunctionAssociations: cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArray{
				cloudfront.DistributionDefaultCacheBehaviorFunctionAssociationArgs{
					FunctionArn: function.Arn,
					EventType:   pulumi.String("viewer-request"),
				},
			},
		},
		ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
			AcmCertificateArn: certResource.CertificateArn,
			SslSupportMethod:  pulumi.String("sni-only"),
		},
	})
	if err != nil {
		return nil, err
	}

	return cdn, nil
}

func createCloudfrontFunction(ctx *pulumi.Context) (*cloudfront.Function, error) {
	functionCode := `
	function handler(event) { 
		var request = event.request; 
		var uri = request.uri; 

		if (uri == "/" || uri == "/index.html") { 
			request.uri = "/html/index.html"; 
		} else if (uri.startsWith("/pages/")) {
			request.uri = "/html" + uri;
		}

		return request; 
	}
	`
	cloudfrontFunction, err := cloudfront.NewFunction(ctx, "redirect-function", &cloudfront.FunctionArgs{
		Runtime: pulumi.String("cloudfront-js-1.0"),
		Publish: pulumi.Bool(true),
		Code: pulumi.String(functionCode),
	})
	if err != nil {
		return nil, err
	}

	return cloudfrontFunction, nil
}


func createBucketPolicy(ctx *pulumi.Context, bucket *s3.BucketV2, cdn *cloudfront.Distribution) error {
	_, err := s3.NewBucketPublicAccessBlock(ctx, "website-bucket-allow-public-access", &s3.BucketPublicAccessBlockArgs{
		Bucket:                bucket.ID(),
		BlockPublicAcls:       pulumi.Bool(true),
		BlockPublicPolicy:     pulumi.Bool(false),
		IgnorePublicAcls:      pulumi.Bool(true),
		RestrictPublicBuckets: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":      "PublicReadGetObject",
				"Effect":   "Allow",
				"Principal": map[string]interface{}{
					"Service": "cloudfront.amazonaws.com",
				},
				"Action":   []string{"s3:GetObject"},
				"Resource": []interface{}{
					pulumi.Sprintf("arn:aws:s3:::%s/*", bucket.ID()),		
				},
				"Condition": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"AWS:SourceArn": cdn.Arn,
					},
				},
			},
		},
	}

	// Attach the policy to the bucket
	_, err = s3.NewBucketPolicy(ctx, "website-bucket-policy", &s3.BucketPolicyArgs{
		Bucket: bucket.ID(),              
		Policy: pulumi.Any(policy),
	})
	if err != nil {
		return err
	}

	return nil
}

