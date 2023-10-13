package main

import (
	"io/fs"
	"mime"
	"path"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
)

func CreateBucket (ctx *pulumi.Context, prefix string) (*s3.BucketV2, error) {
	conf := config.New(ctx, "bucket")
	
	name := conf.Require("name")
	bucket, err := s3.NewBucketV2(ctx, prefix + "-" + name, nil)
	if err != nil {
		return nil, err
	}
	
	return bucket, nil
}

type UploadDirectoryContentArgs struct {
	Bucket	      *s3.BucketV2
	DirectoryPath string
}

func UploadDirectoryContent (ctx *pulumi.Context, args *UploadDirectoryContentArgs) error {
	err := filepath.Walk(args.DirectoryPath, func(name string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			rel, err := filepath.Rel(args.DirectoryPath, name)
			if err != nil {
				return err
			}

			if _, err := s3.NewBucketObject(ctx, rel, &s3.BucketObjectArgs{
				Bucket:      args.Bucket.ID(),                                    // reference to the s3.Bucket object
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

type CreateBucketPolicyArgs struct {
	Bucket   	       *s3.BucketV2
	CloudfrontDistribution *cloudfront.Distribution
}

func CreateBucketPolicy(ctx *pulumi.Context, prefix string, args *CreateBucketPolicyArgs) error {
	conf := config.New(ctx, "bucket")
	
	_, err := s3.NewBucketPublicAccessBlock(ctx, "website-bucket-allow-public-access", &s3.BucketPublicAccessBlockArgs{
		Bucket:                args.Bucket.ID(),
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
					pulumi.Sprintf("arn:aws:s3:::%s/*", args.Bucket.ID()),		
				},
				"Condition": map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"AWS:SourceArn": args.CloudfrontDistribution.Arn,
					},
				},
			},
		},
	}

	// Attach the policy to the bucket
	bucketPolicyName := conf.Require("bucketPolicyName")
	_, err = s3.NewBucketPolicy(ctx, prefix + "-" + bucketPolicyName, &s3.BucketPolicyArgs{
		Bucket: args.Bucket.ID(),             
		Policy: pulumi.Any(policy),
	})
	if err != nil {
		return err
	}

	return nil
}

