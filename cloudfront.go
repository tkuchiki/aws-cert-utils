package certutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/olekukonko/tablewriter"
)

type CloudFront struct {
	client    *cloudfront.CloudFront
	iamClient *IAM
	marker    string
	maxItems  int64
}

type CFDistribution struct {
	id         string
	domain     string
	cert       string
	aliasesStr string
	aliases    []string
}

func NewCloudFront(sess *session.Session, marker string, maxItems int64) *CloudFront {
	return &CloudFront{
		client: cloudfront.New(sess),
		iamClient: &IAM{
			client: iam.New(sess),
		},
		marker:   marker,
		maxItems: int64(maxItems),
	}
}

func createCFListDistributionsInput(marker string, maxItems int64) *cloudfront.ListDistributionsInput {
	dinput := &cloudfront.ListDistributionsInput{}

	if marker != "" {
		dinput.SetMarker(marker)
	}

	if maxItems > 0 {
		dinput.SetMaxItems(int64(maxItems))
	}

	return dinput
}

func (cf *CloudFront) getDistributions(certFilter, aliasesFilter string) ([]CFDistribution, error) {
	out, err := cf.client.ListDistributions(createCFListDistributionsInput(cf.marker, cf.maxItems))
	if err != nil {
		return []CFDistribution{}, err
	}

	iamDescs, err := cf.iamClient.ListMap("", int64(0), "")
	if err != nil {
		return []CFDistribution{}, err
	}

	dists := make([]CFDistribution, 0, len(out.DistributionList.Items))
	for _, summary := range out.DistributionList.Items {
		dist := CFDistribution{}

		vCert := summary.ViewerCertificate
		var iamCert string
		if aws.StringValue(vCert.ACMCertificateArn) != "" {
			dist.cert = *vCert.ACMCertificateArn
		} else if aws.StringValue(vCert.IAMCertificateId) != "" {
			dist.cert = *vCert.IAMCertificateId
			iamCert = fmt.Sprintf("%s | %s", *vCert.IAMCertificateId, iamDescs[*vCert.IAMCertificateId].name)
		} else {
			continue
		}

		if certFilter != "" && dist.cert != certFilter {
			continue
		}

		if iamCert != "" {
			dist.cert = iamCert
		}

		aliases := summary.Aliases.Items

		if len(aliases) > 0 {
			dist.aliases = aws.StringValueSlice(aliases)
			dist.aliasesStr = toFlatten(aliases)
		}

		if aliasesFilter != "" && (dist.aliasesStr == "" || strings.Index(dist.aliasesStr, aliasesFilter) < 0) {
			continue
		}

		dist.id = *summary.Id
		dist.domain = *summary.DomainName

		dists = append(dists, dist)
	}
	return dists, err
}

func (cf *CloudFront) List(certFilter, aliasesFilter string) ([]CFDistribution, error) {
	return cf.getDistributions(certFilter, aliasesFilter)
}

func createCFDistributionConfig(vc *cloudfront.ViewerCertificate) *cloudfront.DistributionConfig {
	dc := &cloudfront.DistributionConfig{}

	dc.SetViewerCertificate(vc)

	return dc
}

func createCFViewerCertificate(vc *cloudfront.ViewerCertificate, service, cert string) *cloudfront.ViewerCertificate {
	newvc := &cloudfront.ViewerCertificate{}

	switch service {
	case "acm":
		newvc.SetACMCertificateArn(cert)
	case "iam":
		newvc.SetIAMCertificateId(cert)
	}
	newvc.SetCloudFrontDefaultCertificate(false)
	newvc.SetMinimumProtocolVersion(*vc.MinimumProtocolVersion)
	newvc.SetSSLSupportMethod(*vc.SSLSupportMethod)

	return newvc
}

func createCFUpdateDistributionInput(distOut *cloudfront.GetDistributionOutput, service, cert string) *cloudfront.UpdateDistributionInput {
	dinput := &cloudfront.UpdateDistributionInput{}

	dist := distOut.Distribution

	dinput.SetId(*dist.Id)
	dinput.SetIfMatch(*distOut.ETag)

	dc := dist.DistributionConfig
	vc := dc.ViewerCertificate

	newvc := createCFViewerCertificate(vc, service, cert)

	dc.SetViewerCertificate(newvc)

	dinput.SetDistributionConfig(dc)

	return dinput
}

func createGetDistributionInput(id string) *cloudfront.GetDistributionInput {
	dinput := &cloudfront.GetDistributionInput{}

	dinput.SetId(id)

	return dinput
}

func (cf *CloudFront) GetDistribution(id string) (*cloudfront.GetDistributionOutput, error) {
	return cf.client.GetDistribution(createGetDistributionInput(id))
}

func getCertificate(vc *cloudfront.ViewerCertificate) string {
	if aws.StringValue(vc.ACMCertificateArn) != "" {
		return aws.StringValue(vc.ACMCertificateArn)
	}
	if aws.StringValue(vc.IAMCertificateId) != "" {
		return aws.StringValue(vc.IAMCertificateId)
	}

	return ""
}

func (cf *CloudFront) Update(id, service, cert string) (string, error) {
	distOut, err := cf.GetDistribution(id)
	if err != nil {
		return "", err
	}

	srcCert := getCertificate(distOut.Distribution.DistributionConfig.ViewerCertificate)

	_, err = cf.client.UpdateDistribution(createCFUpdateDistributionInput(distOut, service, cert))
	if err != nil {
		return "", err
	}

	aliases := toFlatten(distOut.Distribution.DistributionConfig.Aliases.Items)

	return cfUpdateMsg(id, aliases, srcCert, cert), nil
}

func cfUpdateMsg(id, aliases, src, dest string) string {
	return fmt.Sprintf("Updated %s %s %s -> %s", id, aliases, src, dest)
}

func (cf *CloudFront) BulkUpdate(service, srcCert, destCert string, dryRun bool) ([]string, error) {
	dists, err := cf.getDistributions(srcCert, "")
	if err != nil {
		return []string{}, err
	}

	updates := make([]string, 0)
	if dryRun {
		updates = append(updates, dryRunMsg()...)
	}

	for _, dist := range dists {
		if dryRun {
			updates = append(updates, cfUpdateMsg(dist.id, dist.aliasesStr, srcCert, destCert))
		} else {
			distOut, err := cf.GetDistribution(dist.id)
			if err != nil {
				return []string{}, err
			}

			_, err = cf.client.UpdateDistribution(createCFUpdateDistributionInput(distOut, service, destCert))

			if err != nil {
				return []string{}, err
			}
			updates = append(updates, cfUpdateMsg(dist.id, dist.aliasesStr, srcCert, destCert))
		}

	}

	return updates, nil
}

func (cf *CloudFront) ReadableList(dists []CFDistribution) {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Distribution ID", "Aliases", "SSL Certificate"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	for _, dist := range dists {
		for _, alias := range dist.aliases {
			table.Append([]string{dist.id, alias, dist.cert})
		}
	}

	table.Render()
}
