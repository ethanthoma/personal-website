package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"fmt"
	"path/filepath"
	"io/fs"
	"mime"
	"path"
)

const SOURCE_DIR = "../dist"

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)
		bucket, err := s3.NewBucket(ctx, "website-bucket", &s3.BucketArgs{
		    Website: s3.BucketWebsiteArgs{
			IndexDocument: pulumi.String("index.html"),
		    },
		})
		if err != nil {
			return err
		}

		// Upload all files within the dir
		err = filepath.Walk(SOURCE_DIR, func(name string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if !info.IsDir() {
				rel, err := filepath.Rel(SOURCE_DIR, name)
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

		// Set the access policy for the bucket so all objects are readable.
		if _, err := s3.NewBucketPolicy(ctx, "bucket-policy", &s3.BucketPolicyArgs{
			Bucket: bucket.ID(), // refer to the bucket created earlier
			Policy: pulumi.Any(map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect":    "Allow",
						"Principal": "*",
						"Action": []interface{}{
							"s3:GetObject",
						},
						"Resource": []interface{}{
							pulumi.Sprintf("arn:aws:s3:::%s/*", bucket.ID()), // policy refers to bucket name explicitly
						},
					},
				},
			}),
		}); err != nil {
			return err
		}

		// Export website endpoint
		ctx.Export("bucketEndpoint", bucket.WebsiteEndpoint.ApplyT(func(websiteEndpoint string) (string, error) {
		    return fmt.Sprintf("http://%v", websiteEndpoint), nil
		}).(pulumi.StringOutput))

		return nil
	})
}
