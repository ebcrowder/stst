# stst
Command line tool for managing temporary Amazon Web Services (AWS) credentials when utilizing AWS services that are protected via multi-factor authentication (MFA) requirements.

AWS [recommends](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_mfa.html) that MFA be configured on IAM and root accounts to protect resources. This tool helps manage temporary credentials when interacting with AWS services via the AWS SDK, AWS CLI or other means.

This tool takes the 6-digit MFA code as input and uses it, along with the `region` and `mfa_serial` specified in the AWS `config` file, to authenticate with AWS Security Token Service (STS). If the tool is able to successfully authenticate with STS, it will store the temporary credentials in the format that the AWS CLI expects within the AWS `credentials` file:
```bash
[temp]
aws_access_key_id = accessKeyId
aws_secret_access_key = secretAccessKey
aws_security_token = sessionToken
aws_token_expiration = 2020-01-01T00:00:00Z
```
If the `profile` entry does not already exist in the `credentials` file, the tool will create it. On subsequent runs, the tool will update the entry in place.

At which point, you can invoke the AWS CLI by specifying the appropriate profile (`temp` in this example) by doing the following:
```bash
$ aws iam get-user --profile temp

{
    "User": {
        "Path": "/",
        "UserName": "user",
        "UserId": "iamUser123",
        "Arn": "arn:aws:iam::123456789012:user/iamUser123",
        "CreateDate": "2020-01-01T00:00:00Z",
        "PasswordLastUsed": "2021-01-01T00:00:00Z"
    }
}
```
If you are using the AWS SDK, the profile name will need to be specified programmatically using the methods for the particular SDK being utilized.
## Prerequisites
- An AWS account (root or IAM) that possesses an IAM policy requiring MFA on all or specific AWS services.
- An AWS `config` and `credentials` file. See [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) for further information.
- The AWS `config` file should specify the `region` and `mfa_serial` values.
- Go 1.15+
## Installation
Clone this repo and then run:
```bash
$ go build
```
For Linux and macOS environments, move the binary to `/usr/bin` or wherever executable binaries for applications are managed in your environment. For Windows, move the executable to `C:\Program Files`.
## Usage
### Help
```bash
$ stst --help

Usage of stst:
-config string
Path to aws config file (default "/home/$USER/.aws/config")
-credentials string
Path to aws credentials file (default "/home/$USER/.aws/credentials")
-duration int
Duration in seconds that temporary credentials should remain valid (default 900)
-profile string
Profile name used to associate temporary credentials. (default "temp")
```
### Authenticate with AWS
If the temporary credentials have expired or do not exist:
```bash
$ stst
Please enter your MFA Code: 123456
Successfully authenticated with AWS STS and updated the AWS credentials file at: /home/$USER/.aws/credentials
```
If the current temporary credentials are valid:
```bash
$ stst
Temporary credentials have not expired and remain valid.
```
## Tests
```bash
$ go test
PASS
ok  	github.com/ebcrowder/stst/v2	0.002s
```