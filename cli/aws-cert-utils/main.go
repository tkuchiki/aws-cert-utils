package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tkuchiki/aws-cert-utils"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	crtUtils           = kingpin.New("aws-cert-utils", "Certificate Utility for AWS(ACM, IAM, ALB, ELB, CloudFront)")
	awsAccessKeyID     = crtUtils.Flag("access-key", "The AWS access key ID").String()
	awsSecretAccessKey = crtUtils.Flag("secret-key", "The AWS secret access key").String()
	awsArn             = crtUtils.Flag("assume-role-arn", "The AWS assume role ARN").String()
	awsToken           = crtUtils.Flag("token", "The AWS access token").String()
	awsRegion          = crtUtils.Flag("region", "The AWS region").String()
	awsProfile         = crtUtils.Flag("profile", "The AWS CLI profile").String()
	awsConfig          = crtUtils.Flag("aws-config", "The AWS CLI Config file").String()
	awsCreds           = crtUtils.Flag("credentials", "The AWS CLI Credential file").String()

	// acm
	acmCmd = crtUtils.Command("acm", "AWS Certificate Manager (ACM)")
	// acm list
	acmListCmd      = acmCmd.Command("list", "Retrieves a list of ACM Certificates and the domain name for each")
	acmListStatuses = acmListCmd.Flag("cert-statuses", "The status or statuses on which to filter the list of ACM Certificates(comma separated)").Default("ALL").String()
	acmListMaxItems = acmListCmd.Flag("max-items", "The total number of items to return in the command's output").Int()

	// acm import
	acmImportCmd       = acmCmd.Command("import", "Imports an SSL/TLS certificate into AWS Certificate Manager (ACM) to use with ACM's integrated AWS services")
	acmImportCertPath  = acmImportCmd.Flag("cert-path", "Path to certificate").String()
	acmImportChainPath = acmImportCmd.Flag("chain-path", "Path to certificate chain").String()
	acmImportPkeyPath  = acmImportCmd.Flag("pkey-path", "Path to private key").String()
	acmImportCert      = acmImportCmd.Flag("cert", "The certificate to import").String()
	acmImportChain     = acmImportCmd.Flag("chain", "The certificate chain").String()
	acmImportPkey      = acmImportCmd.Flag("pkey", "The private key that matches the public key in the certificate").String()
	acmImportName      = acmImportCmd.Flag("name", "The name tag value").String()

	// acm delete
	acmDeleteCmd      = acmCmd.Command("delete", "Deletes an ACM Certificate and its associated private key")
	acmDeleteArn      = acmDeleteCmd.Flag("arn", "String that contains the ARN of the ACM Certificate to be deleted").String()
	acmDeleteStatuses = acmDeleteCmd.Flag("cert-statuses", "The status or statuses on which to filter the list of ACM Certificates(comma separated)").Default("ALL").String()
	acmDeleteMaxItems = acmDeleteCmd.Flag("max-items", "The total number of items to return in the command's output").Int()

	// iam
	iamCmd = crtUtils.Command("iam", "AWS  Identity and Access Management (IAM)")
	// iam list
	iamListCmd        = iamCmd.Command("list", "Lists the server certificates stored in IAM that have the specified path prefix")
	iamListMarker     = iamListCmd.Flag("marker", "Paginating results and only after you receive a response indicating that the results are truncated").String()
	iamListMaxItems   = iamListCmd.Flag("max-items", "The total number of items to return").Int()
	iamListPathPrefix = iamListCmd.Flag("path-prefix", "The path prefix for filtering the results").Default("/").String()

	// iam upload
	iamUploadCmd       = iamCmd.Command("upload", "Uploads a server certificate entity for the AWS account")
	iamUploadCertPath  = iamUploadCmd.Flag("cert-path", "Path to certificate").String()
	iamUploadChainPath = iamUploadCmd.Flag("chain-path", "Path tocertificate chain").String()
	iamUploadPkeyPath  = iamUploadCmd.Flag("pkey-path", "Path to private key").String()
	iamUploadCert      = iamUploadCmd.Flag("cert", "The contents of the public key certificate").String()
	iamUploadChain     = iamUploadCmd.Flag("chain", "The contents of the certificate chain").String()
	iamUploadPkey      = iamUploadCmd.Flag("pkey", "The contents of the private key").String()
	iamUploadPath      = iamUploadCmd.Flag("path", "The path for the server certificate").Default("/").String()
	iamUploadName      = iamUploadCmd.Flag("name", "The name for the server certificate").String()

	// iam update
	iamUpdateCmd     = iamCmd.Command("update", "Updates the name and/or the path of the specified server certificate stored in IAM")
	iamUpdateNewPath = iamUpdateCmd.Flag("new-path", "The new path for the server certificate").String()
	iamUpdateNewName = iamUpdateCmd.Flag("new-name", "The new name for the server certificate").String()
	iamUpdateName    = iamUpdateCmd.Flag("name", "The name for the server certificate").String()

	// iam delete
	iamDeleteCmd        = iamCmd.Command("delete", "Deletes the specified server certificate")
	iamDeleteName       = iamDeleteCmd.Flag("name", "The name of the server certificate you want to delete").String()
	iamDeleteMarker     = iamDeleteCmd.Flag("marker", "Paginating results and only after you receive a response indicating that the results are truncated").String()
	iamDeleteMaxItems   = iamDeleteCmd.Flag("max-items", "The total number of items to return").Int()
	iamDeletePathPrefix = iamDeleteCmd.Flag("path-prefix", "The path prefix for filtering the results").Default("/").String()

	// cloudfront
	cfCmd      = crtUtils.Command("cloudfront", "Amazon CloudFront")
	cfMarker   = cfCmd.Flag("marker", "Paginating results and only after you receive a response indicating that the results are truncated").String()
	cfMaxItems = cfCmd.Flag("max-items", "The total number of items to return in the command's output").Int()

	// cloudfront list
	cfListCmd           = cfCmd.Command("list", "Lists the distributions")
	cfListCertFilter    = cfListCmd.Flag("cert", "ACM Arn or IAM Certificate ID").String()
	cfListAliasesFilter = cfListCmd.Flag("aliases", "Domain name").String()

	// cloudfront update
	cfUpdateCmd    = cfCmd.Command("update", "Updates the configuration for a distribution")
	cfUpdateDistId = cfUpdateCmd.Flag("dist-id", "The distribution's id").String()
	cfUpdateACMArn = cfUpdateCmd.Flag("acm-arn", "String that contains the ARN of the ACM Certificate").String()
	cfUpdateIAMId  = cfUpdateCmd.Flag("iam-id", "String that contains the IAM Certificate ID").String()

	// cloudfront bulk-update
	cfBUpdateCmd        = cfCmd.Command("bulk-update", "Updates the configuration for distributions")
	cfBUpdateSrcACMArn  = cfBUpdateCmd.Flag("source-acm-arn", "String that contains the ARN of the source ACM Certificate").String()
	cfBUpdateSrcIAMId   = cfBUpdateCmd.Flag("source-iam-id", "String that contains the source IAM Certificate ID").String()
	cfBUpdateDestACMArn = cfBUpdateCmd.Flag("dest-acm-arn", "String that contains the ARN of the destination ACM Certificate").String()
	cfBUpdateDestIAMId  = cfBUpdateCmd.Flag("dest-iam-id", "String that contains the destination IAM Certificate ID").String()
	cfBUpdateNoDryRun   = cfBUpdateCmd.Flag("no-dry-run", "Disable dry-run mode").Bool()

	// elb
	elbCmd = crtUtils.Command("elb", "Elastic Load Balancing")
	// elb list
	elbListCmd        = elbCmd.Command("list", "Describes the specified  the load balancers")
	elbListCertFilter = elbListCmd.Flag("cert", "String that contains the ARN of the ACM/IAM Certificate").PlaceHolder("ARN").String()
	// elb update
	elbUpdateCmd  = elbCmd.Command("update", "Updates the specified a listener from the specified load balancer")
	elbUpdateName = elbUpdateCmd.Flag("name", "The name of the load balancer").String()
	elbUpdatePort = elbUpdateCmd.Flag("port", "The port that uses the specified SSL certificate").Default("443").Int()
	elbUpdateArn  = elbUpdateCmd.Flag("cert-arn", "The ARN of the ACM/IAM SSL Certificate").String()

	// elb bulk-update
	elbBUpdateCmd         = elbCmd.Command("bulk-update", "Updates the specified listeners from the specified load balancer")
	elbBUpdateSrcCertArn  = elbBUpdateCmd.Flag("source-cert-arn", "The ARN of the source ACM/IAM SSL Certificate").String()
	elbBUpdateDestCertArn = elbBUpdateCmd.Flag("dest-cert-arn", "The ARN of the destination ACM/IAM SSL Certificate").String()
	elbBUpdateNoDryRun    = elbBUpdateCmd.Flag("no-dry-run", "Disable dry-run mode").Bool()

	// alb
	albCmd = crtUtils.Command("alb", "Application Load Balancing")
	// alb list
	albListCmd        = albCmd.Command("list", "Describes the specified load balancers")
	albListCertFilter = albListCmd.Flag("cert", "The ARN of the ACM/IAM SSL Certificate").PlaceHolder("ARN").String()

	// alb update
	albUpdateCmd  = albCmd.Command("update", "Updates the specified a listener from the specified load balancer")
	albUpdateName = albUpdateCmd.Flag("name", "The name of the load balancer").String()
	albUpdateArn  = albUpdateCmd.Flag("cert-arn", "The ARN of the source ACM/IAM SSL Certificate").String()

	// alb bulk-update
	albBUpdateCmd         = albCmd.Command("bulk-update", "Updates the specified listeners from the specified load balancer")
	albBUpdateSrcCertArn  = albBUpdateCmd.Flag("source-cert-arn", "The ARN of the source ACM/IAM SSL Certificate").String()
	albBUpdateDestCertArn = albBUpdateCmd.Flag("dest-cert-arn", "The ARN of the destination ACM/IAM SSL Certificate").String()
	albBUpdateNoDryRun    = albBUpdateCmd.Flag("no-dry-run", "Disable dry-run mode").Bool()
)

func main() {
	crtUtils.Version("0.1.1")
	subCmd, err := crtUtils.Parse(os.Args[1:])

	if err != nil {
		log.Fatal(err)
	}

	cmds := strings.Split(subCmd, " ")

	var region string
	if cmds[0] == "iam" || cmds[0] == "cloudfront" {
		region = "us-east-1"
	} else {
		region = *awsRegion
	}

	sess, err := certutils.NewAWSSession(*awsAccessKeyID, *awsSecretAccessKey, *awsArn, *awsToken, region, *awsProfile, *awsConfig, *awsCreds)
	if err != nil {
		log.Fatal(err)
	}

	switch cmds[0] {
	case "acm":
		a := certutils.NewACM(sess)
		switch cmds[1] {
		case "list":
			out, err := a.List(*acmListStatuses, int64(*acmListMaxItems), "")
			if err != nil {
				log.Fatal(err)
			}

			a.ReadableList(out)
		case "import":
			err := certutils.CheckTagValuePattern(*acmImportName)
			if err != nil {
				log.Fatal(err)
			}

			cm := certutils.NewCertificateManager()

			err = cm.LoadCertificate(*acmImportCert, *acmImportCertPath)
			if err != nil {
				log.Fatal(err)
			}

			err = cm.LoadChain(*acmImportChain, *acmImportChainPath)
			if err != nil {
				log.Fatal(err)
			}

			err = cm.LoadPrivateKey(*acmImportPkey, *acmImportPkeyPath)
			if err != nil {
				log.Fatal(err)
			}

			err = cm.CheckPrivateKeyBitLen()
			if err != nil {
				log.Fatal(err)
			}

			arn, msg, err := a.Import(cm.Cert, cm.Chain, cm.Pkey)
			if err != nil {
				log.Fatal(err)
			}

			if *acmImportName != "" {
				err = a.AddTags(arn, []certutils.Tag{
					certutils.Tag{
						Key:   "Name",
						Value: *acmImportName,
					},
				})
				if err != nil {
					log.Fatal(err)
				}
			}

			fmt.Println(msg)
		case "delete":
			arn := *acmDeleteArn
			if arn == "" {
				arns, targets, err := a.ListDeleteTargets(*acmDeleteStatuses, int64(*acmDeleteMaxItems), "")
				if err != nil {
					log.Fatal(err)
				}
				arn = certutils.Choice(arns, "Choose the server certificate you want to delete : ", 20)

				arn = targets[arn]

				if arn == "" {
					os.Exit(0)
				}
			}

			msg, err := a.Delete(arn)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(msg)
		}
	case "iam":
		i := certutils.NewIAM(sess)
		switch cmds[1] {
		case "list":
			descs, err := i.List(*iamListMarker, int64(*iamListMaxItems), *iamListPathPrefix)
			if err != nil {
				log.Fatal(err)
			}

			i.ReadableList(descs)
		case "upload":
			cm := certutils.NewCertificateManager()

			err = cm.LoadCertificate(*iamUploadCert, *iamUploadCertPath)
			if err != nil {
				log.Fatal(err)
			}

			err = cm.LoadChain(*iamUploadChain, *iamUploadChainPath)
			if err != nil {
				log.Fatal(err)
			}

			err = cm.LoadPrivateKey(*iamUploadPkey, *iamUploadPkeyPath)
			if err != nil {
				log.Fatal(err)
			}

			err = cm.CheckPrivateKeyBitLen()
			if err != nil {
				log.Fatal(err)
			}

			msg, err := i.Upload(cm.Cert, cm.Chain, cm.Pkey, *iamUploadPath, *iamUploadName)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(msg)
		case "update":
			msg, err := i.Update(*iamUpdateNewPath, *iamUpdateNewName, *iamUpdateName)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(msg)
		case "delete":
			name := *iamDeleteName
			if name == "" {
				names, err := i.ListNames(*iamDeleteMarker, int64(*iamDeleteMaxItems), *iamDeletePathPrefix)
				if err != nil {
					log.Fatal(err)
				}
				name = certutils.Choice(names, "Choose the server certificate you want to delete : ", 20)

				if name == "" {
					os.Exit(0)
				}
			}

			msg, err := i.Delete(name)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(msg)
		}
	case "cloudfront":
		cf := certutils.NewCloudFront(sess, *cfMarker, int64(*cfMaxItems))
		switch cmds[1] {
		case "list":
			dists, err := cf.List(*cfListCertFilter, *cfListAliasesFilter)
			if err != nil {
				log.Fatal(err)
			}

			cf.ReadableList(dists)
		case "update":
			var service, cert string
			if *cfUpdateACMArn == "" && *cfUpdateIAMId == "" {
				log.Fatal("--acm-arn or --iam-id is required.")
			} else if *cfUpdateACMArn != "" && *cfUpdateIAMId != "" {
				log.Fatal("--acm-arn or --iam-id but not both.")
			} else if *cfUpdateACMArn != "" {
				cert = *cfUpdateACMArn
				service = "acm"
			} else {
				cert = *cfUpdateIAMId
				service = "iam"
			}

			dist, err := cf.Update(*cfUpdateDistId, service, cert)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(dist)
		case "bulk-update":
			var service, srcCert, destCert string
			if *cfBUpdateSrcACMArn == "" && *cfBUpdateSrcIAMId == "" {
				log.Fatal("--source-acm-arn or --source-iam-id is required.")
			} else if *cfBUpdateSrcACMArn != "" && *cfBUpdateSrcIAMId != "" {
				log.Fatal("--source-acm-arn or --source-iam-id but not both.")
			} else if *cfBUpdateSrcACMArn != "" {
				srcCert = *cfBUpdateSrcACMArn
			} else {
				srcCert = *cfBUpdateSrcIAMId
			}

			if *cfBUpdateDestACMArn == "" && *cfBUpdateDestIAMId == "" {
				log.Fatal("--dest-acm-arn or --dest-iam-id is required.")
			} else if *cfBUpdateDestACMArn != "" && *cfBUpdateDestIAMId != "" {
				log.Fatal("--dest-acm-arn or --dest-iam-id but not both.")
			} else if *cfBUpdateDestACMArn != "" {
				destCert = *cfBUpdateDestACMArn
				service = "acm"
			} else {
				destCert = *cfBUpdateDestIAMId
				service = "iam"
			}

			dists, err := cf.BulkUpdate(service, srcCert, destCert, !*cfBUpdateNoDryRun)
			if err != nil {
				log.Fatal(err)
			}

			for _, dist := range dists {
				fmt.Println(dist)
			}
		}
	case "elb":
		e := certutils.NewELB(sess)
		switch cmds[1] {
		case "list":
			descs, err := e.List(*elbListCertFilter)
			if err != nil {
				log.Fatal(err)
			}

			e.ReadableList(descs)
		case "update":
			update, err := e.Update(*elbUpdateName, int64(*elbUpdatePort), *elbUpdateArn)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(update)
		case "bulk-update":
			updates, err := e.BulkUpdate(*elbBUpdateSrcCertArn, *elbBUpdateDestCertArn, !*elbBUpdateNoDryRun)
			if err != nil {
				log.Fatal(err)
			}

			for _, u := range updates {
				fmt.Println(u)
			}
		}
	case "alb":
		alb := certutils.NewALB(sess)
		switch cmds[1] {
		case "list":
			descs, err := alb.List(*albListCertFilter)
			if err != nil {
				log.Fatal(err)
			}

			alb.ReadableList(descs)
		case "update":
			err := alb.Update(*albUpdateName, *albUpdateArn)
			if err != nil {
				log.Fatal(err)
			}
		case "bulk-update":
			albs, err := alb.BulkUpdate(*albBUpdateSrcCertArn, *albBUpdateDestCertArn, !*albBUpdateNoDryRun)
			if err != nil {
				log.Fatal(err)
			}

			for _, alb := range albs {
				fmt.Println(alb)
			}
		}
	}
}
