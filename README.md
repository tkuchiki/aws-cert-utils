# aws-cert-utils

Certificate Utility for AWS(ACM, IAM, ALB, ELB, CloudFront)

## Installation

Download from https://github.com/tkuchiki/aws-cert-utils/releases

## Usage

```console
usage: aws-cert-utils [<flags>] <command> [<args> ...]

Certificate Utility for AWS(ACM, IAM, ALB, ELB, CloudFront)

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.

  acm list [<flags>]
    Retrieves a list of ACM Certificates and the domain name for each

  acm import [<flags>]
    Imports an SSL/TLS certificate into AWS Certificate Manager (ACM) to use
    with ACM's integrated AWS services

  acm delete [<flags>]
    Deletes an ACM Certificate and its associated private key

  iam list
    Lists the server certificates stored in IAM that have the specified path
    prefix

  iam upload [<flags>]
    Uploads a server certificate entity for the AWS account

  iam update [<flags>]
    Updates the name and/or the path of the specified server certificate stored
    in IAM

  iam delete [<flags>]
    Deletes the specified server certificate

  cloudfront list [<flags>]
    Lists the distributions

  cloudfront update [<flags>]
    Updates the configuration for a distribution

  cloudfront bulk-update [<flags>]
    Updates the configuration for distributions

  elb list [<flags>]
    Describes the specified the load balancers

  elb update [<flags>]
    Updates the specified a listener from the specified load balancer

  elb bulk-update [<flags>]
    Updates the specified listeners from the specified load balancer

  alb list [<flags>]
    Describes the specified load balancers

  alb update [<flags>]
    Updates the specified a listener from the specified load balancer

  alb bulk-update [<flags>]
    Updates the specified listeners from the specified load balancer

```

### ACM

```console
$ ./aws-cert-utils acm --help
usage: aws-cert-utils acm <command> [<args> ...]

AWS Certificate Manager (ACM)

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Subcommands:
  acm list [<flags>]
    Retrieves a list of ACM Certificates and the domain name for each

  acm import [<flags>]
    Imports an SSL/TLS certificate into AWS Certificate Manager (ACM) to use with ACM's integrated AWS services

  acm delete [<flags>]
    Deletes an ACM Certificate and its associated private key

```

#### List

```console
$ ./aws-cert-utils acm list
+------------------------+-----------------+-----------------+---------+-------------------------------+-------------------------------------------------------------------------------------+
|        NAME TAG        |   DOMAIN NAME   | ADDITIONAL NAME | IN USE? |           NOT AFTER           |                                   CERTIFICATE ARN                                   |
+------------------------+-----------------+-----------------+---------+-------------------------------+-------------------------------------------------------------------------------------+
|                        | *.example.com   | example.com     | Yes     | 2019-11-14 02:44:43 +0000 UTC | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+------------------------+                 +                 +         +                               +-------------------------------------------------------------------------------------+
| example.com            |                 |                 |         |                               | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy |
+------------------------+-----------------+-----------------+---------+-------------------------------+-------------------------------------------------------------------------------------+

```

#### Import

```console
$ openssl rsa -in 4096key.pem -text -noout | head -n 1
Private-Key: (4096 bit)

$ ./aws-cert-utils acm import --cert-path 4096cert.pem --pkey-path 4096key.pem
2017/11/30 17:58:03 Invalid private key length (4096 bit). AWS supports 1024 and 2048 bit RSA private key

$ ./aws-cert-utils acm import --cert-path cert.pem --pkey-path key.pem --chain-path ca.pem
Imported arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz
```

#### Delete

```console
$ ./aws-cert-utils acm delete
? Choose the server certificate you want to delete :  arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz
Deleted arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz
```

### IAM

```console
$ ./aws-cert-utils iam --help
usage: aws-cert-utils iam <command> [<args> ...]

AWS Identity and Access Management (IAM)

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Subcommands:
  iam list
    Lists the server certificates stored in IAM that have the specified path prefix

  iam upload [<flags>]
    Uploads a server certificate entity for the AWS account

  iam update [<flags>]
    Updates the name and/or the path of the specified server certificate stored in IAM

  iam delete [<flags>]
    Deletes the specified server certificate

```

#### List

```console
$ ./aws-cert-utils iam list
+------------------------------+-----------------------+--------------------------------+-------------------------------------------------------------------------------------+
|             NAME             |          ID           |              PATH              |                                         ARN                                         |
+------------------------------+-----------------------+--------------------------------+-------------------------------------------------------------------------------------+
| test-certificate             | XXXXXXXXXXXXXXXXXXXXX | /                              | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
| test-cloudfront-certificate  | YYYYYYYYYYYYYYYYYYYYY | /cloudfront/                   | arn:aws:iam::xxxxxxxxxxxx:server-certificate/cloudfront/yyyyyyyyyyyyyyyyyyyyyyyyyyy |
+------------------------------+-----------------------+--------------------------------+-------------------------------------------------------------------------------------+
```

#### Upload

```console
$ ./aws-cert-utils iam upload --cert-path cert.pem --chain-path ca.pem --pkey-path key.pem --path /cloudfront/ --name test-cert
Uploaded test-cert arn:aws:iam::xxxxxxxxxxxx:server-certificate/cloudfront/yyyyyyyyyyyyyyyyyyyyyyyyyyy
```

#### Update

```console
$ ./aws-cert-utils iam update --new-path / --new-name test-cert2 --name test-cert
Updated test-cert -> test-cert2
```

#### Delete

```console
$ ./aws-cert-utils iam delete
? Choose the server certificate you want to delete :  test-cert2
Deleted test-cert2
```

### ALB

```console
$ ./aws-cert-utils alb --help
usage: aws-cert-utils alb <command> [<args> ...]

Application Load Balancing

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Subcommands:
  alb list [<flags>]
    Describes the specified load balancers

  alb update [<flags>]
    Updates the specified a listener from the specified load balancer

  alb bulk-update [<flags>]
    Updates the specified listeners from the specified load balancer

```

#### List

```console
$ ./aws-cert-utils alb list
+-----------+------+-------------------------------------------------------------------------------------+
|   NAME    | PORT |                              LISTENER SSL CERTIFICATE                               |
+-----------+------+-------------------------------------------------------------------------------------+
| test-alb  |  443 | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
| test2-alb |  443 | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
```

#### Update

```console
$ ./aws-cert-utils alb update --name test-alb --cert-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Updated test-alb:443 arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

#### Bulk update

```console
$ ./aws-cert-utils alb list
+-----------+------+-------------------------------------------------------------------------------------+
|   NAME    | PORT |                              LISTENER SSL CERTIFICATE                               |
+-----------+------+-------------------------------------------------------------------------------------+
| test-alb  |  443 | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
| test2-alb |  443 | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
  
$ ./aws-cert-utils alb bulk-update --source-cert-arn arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx --dest-cert-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
# Dry run mode

Updated test-alb:443 arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Updated test2-alb:443 arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

$ ./aws-cert-utils alb bulk-update --source-cert-arn arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx --dest-cert-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --no-dry-run
Updated test-alb:443 arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Updated test2-alb:443 arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

$ ./aws-cert-utils alb list
+-----------+------+-------------------------------------------------------------------------------------+
|   NAME    | PORT |                              LISTENER SSL CERTIFICATE                               |
+-----------+------+-------------------------------------------------------------------------------------+
| test-alb  |  443 | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
| test2-alb |  443 | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
```

### ELB

```console
$ ./aws-cert-utils elb --help
usage: aws-cert-utils elb <command> [<args> ...]

Elastic Load Balancing

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Subcommands:
  elb list [<flags>]
    Describes the specified the load balancers

  elb update [<flags>]
    Updates the specified a listener from the specified load balancer

  elb bulk-update [<flags>]
    Updates the specified listeners from the specified load balancer

```

#### List

```console
$ ./aws-cert-utils elb list
+-----------+------+-------------------------------------------------------------------------------------+
|   NAME    | PORT |                              LISTENER SSL CERTIFICATE                               |
+-----------+------+-------------------------------------------------------------------------------------+
| test-elb  |  443 | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
| test2-elb |  443 | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
```

#### Update

```console
$ ./aws-cert-utils elb update --name test-elb --port 443 --cert-arn arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Updated test-elb:443 arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx -> arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

#### Bulk update

```console
$ ./aws-cert-utils elb list
+-----------+------+-------------------------------------------------------------------------------------+
|   NAME    | PORT |                              LISTENER SSL CERTIFICATE                               |
+-----------+------+-------------------------------------------------------------------------------------+
| test-elb  |  443 | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
| test2-elb |  443 | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+

$ ./aws-cert-utils elb bulk-update --source-cert-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --dest-cert-arn arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
# Dry run mode

Updated test-elb:443 arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx -> arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Updated test2-elb:443 arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx -> arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

$ ./aws-cert-utils elb bulk-update --source-cert-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --dest-cert-arn arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx --no-dry-run
Updated test-elb:443 arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx -> arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Updated test2-elb:443 arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx -> arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

$ ./aws-cert-utils elb list
+-----------+------+-------------------------------------------------------------------------------------+
|   NAME    | PORT |                              LISTENER SSL CERTIFICATE                               |
+-----------+------+-------------------------------------------------------------------------------------+
| test-elb  |  443 | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
| test2-elb |  443 | arn:aws:iam::xxxxxxxxxxxx:server-certificate/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx |
+-----------+------+-------------------------------------------------------------------------------------+
```

### CloudFront

```console
$ ./aws-cert-utils cloudfront --help
usage: aws-cert-utils cloudfront [<flags>] <command> [<args> ...]

Amazon CloudFront

Flags:
  --help           Show context-sensitive help (also try --help-long and --help-man).
  --version        Show application version.
  --max-items=100  The total number of items to return in the command's output

Subcommands:
  cloudfront list [<flags>]
    Lists the distributions

  cloudfront update [<flags>]
    Updates the configuration for a distribution

  cloudfront bulk-update [<flags>]
    Updates the configuration for distributions

```

#### List

```console
$ ./aws-cert-utils
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
| DISTRIBUTION ID |           ALIASES            |                                   SSL CERTIFICATE                                   |
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
| 11111111111111  | iam.example.com              | XXXXXXXXXXXXXXXXXXXXX                                                               |
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
| 22222222222222  | acm.example.com              | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
```

#### Update

```console
$ ./aws-cert-utils cloudfront update --dist-id 11111111111111 --iam-id XXXXXXXXXXXXXXXXXXXXX
Updated 11111111111111 iam.example.com XXXXXXXXXXXXXXXXXXXXX -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

#### Bulk update

```console
$ ./aws-cert-utils cloudfront list
+-----------------+------------------------------+-----------------------------------------------------------------+
| DISTRIBUTION ID |           ALIASES            |                         SSL CERTIFICATE                         |
+-----------------+------------------------------+-----------------------------------------------------------------+
| 11111111111111  | iam.example.com              | XXXXXXXXXXXXXXXXXXXXX                                           |
+-----------------+------------------------------+-----------------------------------------------------------------+
| 22222222222222  | iam2.example.com              | XXXXXXXXXXXXXXXXXXXXX                                           |
+-----------------+------------------------------+-----------------------------------------------------------------+

$ ./aws-cert-utils cloudfront bulk-update --source-iam-id XXXXXXXXXXXXXXXXXXXXX --dest-acm-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
# Dry run mode

Updated 11111111111111 iam.example.com XXXXXXXXXXXXXXXXXXXXX -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Updated 22222222222222 iam2.example.com XXXXXXXXXXXXXXXXXXXXX -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

$ ./aws-cert-utils cloudfront bulk-update --source-iam-id XXXXXXXXXXXXXXXXXXXXX --dest-acm-arn arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --no-dry-run
Updated 11111111111111 iam.example.com XXXXXXXXXXXXXXXXXXXXX -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
Updated 22222222222222 iam2.example.com XXXXXXXXXXXXXXXXXXXXX -> arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

$ ./aws-cert-utils cloudfront list
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
| DISTRIBUTION ID |           ALIASES            |                                   SSL CERTIFICATE                                   |
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
| 11111111111111  | iam.example.com              | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
| 22222222222222  | iam2.example.com             | arn:aws:acm:us-east-1:xxxxxxxxxxxx:certificate/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx |
+-----------------+------------------------------+-------------------------------------------------------------------------------------+
```