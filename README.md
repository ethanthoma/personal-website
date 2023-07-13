# My Personal Portfolio Website

## 0: Requirements

In order to run the bash script, `./build.sh`, you will need both `jq`, `rsync`, and `sass`. 
To deploy the code to pulumi and AWS, you will need to install both the AWS CLI and pulumi.
You will need to setup your creditials in AWS IAM to get an access key and other info.
Set-up the AWS CLI via `aws configure` or follow best practices for your use case.

## 1: Deploying

Once everything has been installed and configured, run the build script in the project directory
with `./build.sh`. If the file doesn't have perms but you do, run `chmod +x build.sh` and try
again. Then, navigate to the `./infra/` directory and run `pulumi up`.
