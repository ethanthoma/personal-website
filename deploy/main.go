package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		resourcePrefix := ctx.Project() + "-" + ctx.Stack()

		cfg := config.New(ctx, "")

		bucketName := "website-bucket"
		domainName := cfg.Require("domainName")
		cloudflareZoneId := cfg.Require("cloudflareZoneId")

		// Create and upload content to s3 bucket
		bucket, err := CreateBucket(ctx, resourcePrefix)
		if err != nil {
			return err
		}

		err = UploadDirectoryContent(ctx, &UploadDirectoryContentArgs{
			bucket, 
			"../dist",
		})
		if err != nil {
			return err
		}

		// Create valid certificate for domain
		certResource, err := CreateValidatedCertificate(ctx, resourcePrefix, &CreateValidatedCertificateArgs{
			cloudflareZoneId,
			"www." + domainName,
		})
		if err != nil {
			return err
		}

		// Create cloudfront distribution for caching 
		cdn, err := CreateDistribution(ctx, resourcePrefix, &CreateDistributionArgs{
			bucket,
			bucketName,
			certResource,
			domainName,
		})
		if err != nil {
			return err
		}

		// Setup Cloudflare DNS for routing domain name to cloudfront 
		err = SetupCloudflareDNS(ctx, resourcePrefix, &SetupCloudflareDNSArgs{
			cloudflareZoneId,
			cdn, 
			domainName, 
		})
		if err != nil {
			return err
		}

		// Set bucket policy to allow cloudfront access
		err = CreateBucketPolicy(ctx, resourcePrefix, &CreateBucketPolicyArgs{
			bucket, 
			cdn,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

