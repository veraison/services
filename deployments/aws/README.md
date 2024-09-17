This directory contains scripts and other resources for instantiating a
Veraison deployment in AWS. The deployment is a CloudFormation stack with a
node running Veraison services, another node running Keycloak authentication,
and an RDS Postgres instance serving as the key-value store.

## Dependencies

This deployment depends on the `debian` deployment (which, in turn, depends on
the `native` deployment). Please see [its
README](../debian/README.md#dependencies) for the dependencies list.

Additionally, the following dependencies are required specifically for AWS deployment:

- `curl`: used to transfer the Debian package to the EC2 node.
- `openssl`: used to generate TLS certs (note: unlike with `native` deployment,
  where pre-generated certs may optionally be used, cert generation is mandatory
  for this deployment, as the certs must be specific to the created EC2 instance).
- `packer`: used to build AMI images using temporary EC2 instances.
- `psql`: Postgres client used to initialise the stores (may be packaged on its
  own or as part of `postgres`, depending on the platform).
- A number of Python packages used by the deployment script. Please see
  [requirements.txt](misc/requirements.txt) for details.

`curl` and `openssl` should be available from your OS's package manager. Python
dependencies are installable via `pip`/`PyPI`. For `packer`, please see [its
documentation](https://developer.hashicorp.com/packer/tutorials/aws-get-started/get-started-install-cli).
             
### Bootstrap

To simplify dependency installation, the deployment script implements bootstrap
for Arch, Ubuntu, and MacOSX (using [homebrew](https://brew.sh)). 

```bash
git clone https://github.com/veraison/services.git
cd services/deployments/aws

make bootstrap
```

(this will only work on the above-mentioned platforms).

### AWS account

Finally, you need an existing AWS account, that has at least one VPC with at
least two subnets (at least one of which is public) configured.

Please see [boto3
documentation](https://boto3.amazonaws.com/v1/documentation/api/latest/guide/credentials.html)
for how to configure `aws` CLI to access this account.


## Working with the deployment

Before creating a deployment, you need to provide account-specific
configuration that specifies the IDs of the VPC and subnets that will be used
for the deployment as well as the CIDR that will be granted access to the
deployment. Please use [misc/arm.cfg](misc/arm.cfg) for an example.

Once the account-specific config file is created, define `AWS_ACCOUNT_CFG`
environment variable to point to it and execute `make deploy` to create the
deployment.

```bash
export AWS_ACCOUNT_CFG=misc/arm.cfg  # replace with path to your config
make deploy
```

Deployment can be accessed via CLI front end:

```bash
source env/env.bash  # for bash, or alternatively, env/env.zsh for Zsh users
veraison status
```

This should display the DNS name and IP address of the instance and show
Veraison services as active and running.

To make sure the deployment works, you can run through
[end-to-end](../../end-to-end/README.md) flow.

For example 

```bash
# env/env.bash must be sourced
../../end-to-end/end-to-end-aws provision
# followed by
../../end-to-end/end-to-end-aws verify rp
# followed by
```

Finally, to remove the deployment, you can run

```bash
make really-clean
```
