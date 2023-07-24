# My Personal Portfolio Website

## 0: Requirements

In order to run the bash script, `./build.sh`, you will need both `jq`, `rsync`, and `sass`. 
To deploy the code to AWS and cloudflare, you will need to install both the AWS CLI and pulumi.
You will need to setup your creditials in AWS IAM to get an access key and other info.
Set-up the AWS CLI via `aws configure` or follow best practices for your use case.
You will need to configure the all the parameters stored in `Pulumi.dev.yaml`.

## 1: Deploying

Once everything has been installed and configured, run the build script in the bin directory
with `./build.sh`. If the file doesn't have perms but you do, run `chmod +x build.sh` and try
again. Then, navigate to the `./infra/` directory and run `pulumi up`. Assuming everything has
been configured correctly, it should deploy to the cloud and connect to your domain name on 
cloudflare.
