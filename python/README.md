# APM Python Lambda Layer

Scripts and files used to build AWS Lambda Layer for running APM Python on AWS Lambda.

### Sample App 

1. Install
   * [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
   * [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
   * [Go](https://go.dev/doc/install)
   * [Docker](https://docs.docker.com/get-docker)
2. Run aws configure to [set aws credential(with administrator permissions)](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install-mac.html#serverless-sam-cli-install-mac-iam-permissions) and default region.
3. Download a local copy of this repository from Github.
4. Navigate to the path `cd python/src`
5. To create a zip file with the APM Python AWS Lambda layer compatible only with Python 3.8: `bash build.sh`
