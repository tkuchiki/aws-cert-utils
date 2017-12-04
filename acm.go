package certutils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/olekukonko/tablewriter"
)

type ACM struct {
	client *acm.ACM
}

type ACMDescription struct {
	arn                     string
	nameTag                 string
	status                  string
	inUseBy                 []string
	notAfter                time.Time
	domainName              string
	subjectAlternativeNames []string
}

func NewACM(sess *session.Session) *ACM {
	return &ACM{
		client: acm.New(sess),
	}
}

func createACMImportCertificateInput(cert, chain, pkey []byte) *acm.ImportCertificateInput {
	return &acm.ImportCertificateInput{
		Certificate:      cert,
		CertificateChain: chain,
		PrivateKey:       pkey,
	}
}

func (a *ACM) Import(cert, chain, pkey []byte) (string, string, error) {
	out, err := a.client.ImportCertificate(createACMImportCertificateInput(cert, chain, pkey))

	return *out.CertificateArn, fmt.Sprintf("Imported %s", *out.CertificateArn), err
}

func createACMListCertificatesInput(statuses []string, maxItems int64, nextToken string) *acm.ListCertificatesInput {
	linput := &acm.ListCertificatesInput{}

	if len(statuses) > 0 {
		linput.SetCertificateStatuses(aws.StringSlice(statuses))
	}

	if maxItems > 0 {
		linput.SetMaxItems(maxItems)
	}

	if nextToken != "" {
		linput.SetNextToken(nextToken)
	}

	return linput
}

func (a *ACM) listTags(arn string) (*acm.ListTagsForCertificateOutput, error) {
	input := &acm.ListTagsForCertificateInput{}
	input.SetCertificateArn(arn)

	return a.client.ListTagsForCertificate(input)
}

func (a *ACM) getNameTag(arn string) (string, error) {
	out, err := a.listTags(arn)
	if err != nil {
		return "", err
	}

	for _, tag := range out.Tags {
		if strings.ToLower(aws.StringValue(tag.Key)) == "name" {
			return *tag.Value, nil
		}
	}

	return "", fmt.Errorf("Name tag not found")
}

func (a *ACM) List(statuses string, maxItems int64, nextToken string) ([]ACMDescription, error) {
	cout, err := a.client.ListCertificates(createACMListCertificatesInput(SplitStatuses(statuses), maxItems, nextToken))

	descs := make([]ACMDescription, 0, len(cout.CertificateSummaryList))
	for _, summary := range cout.CertificateSummaryList {
		dcout, err := a.client.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: summary.CertificateArn,
		})
		if err != nil {
			return []ACMDescription{}, err
		}

		cert := dcout.Certificate
		nameTag, _ := a.getNameTag(*cert.CertificateArn)

		desc := ACMDescription{
			arn:                     *cert.CertificateArn,
			nameTag:                 nameTag,
			status:                  *cert.Status,
			inUseBy:                 aws.StringValueSlice(cert.InUseBy),
			notAfter:                aws.TimeValue(cert.NotAfter),
			domainName:              *cert.DomainName,
			subjectAlternativeNames: aws.StringValueSlice(cert.SubjectAlternativeNames),
		}

		descs = append(descs, desc)
	}

	return descs, err
}

func (a *ACM) ListDeleteTargets(statuses string, maxItems int64, nextToken string) ([]string, map[string]string, error) {
	descs, err := a.List(statuses, maxItems, nextToken)
	if err != nil {
		return []string{}, map[string]string{}, err
	}

	targets := make(map[string]string, 0)
	arns := make([]string, 0, len(descs))
	for _, desc := range descs {
		tagArn := fmt.Sprintf("[%s] %s", desc.nameTag, desc.arn)
		arns = append(arns, tagArn)
		targets[tagArn] = desc.arn
	}

	return arns, targets, err
}

func toACMTags(tags []Tag) []*acm.Tag {
	if len(tags) <= 0 {
		return []*acm.Tag{}
	}

	acmTags := make([]*acm.Tag, 0, len(tags))

	for _, t := range tags {
		acmTag := &acm.Tag{
			Key:   aws.String(t.Key),
			Value: aws.String(t.Value),
		}

		acmTags = append(acmTags, acmTag)
	}

	return acmTags
}

func createACMAddTagsToCertificateInput(arn string, tags []Tag) *acm.AddTagsToCertificateInput {
	tinput := &acm.AddTagsToCertificateInput{}
	tinput.SetCertificateArn(arn)
	tinput.SetTags(toACMTags(tags))

	return tinput
}

func (a *ACM) AddTags(arn string, tags []Tag) error {
	_, err := a.client.AddTagsToCertificate(createACMAddTagsToCertificateInput(arn, tags))

	return err
}

func createACMDeleteCertificateInput(arn string) *acm.DeleteCertificateInput {
	dinput := &acm.DeleteCertificateInput{}

	dinput.SetCertificateArn(arn)

	return dinput
}

func (a *ACM) Delete(arn string) (string, error) {
	_, err := a.client.DeleteCertificate(createACMDeleteCertificateInput(arn))

	return fmt.Sprintf("Deleted %s", arn), err
}

func (a *ACM) ReadableList(descs []ACMDescription) {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Name tag", "Domain Name", "Additional Name", "In Use?", "Not After", "Certificate Arn"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	for _, desc := range descs {
		inUse := "No"
		if len(desc.inUseBy) > 0 {
			inUse = "Yes"
		}
		for _, name := range desc.subjectAlternativeNames {
			if name == desc.domainName {
				continue
			}
			table.Append([]string{desc.nameTag, desc.domainName, name, inUse, desc.notAfter.String(), desc.arn})
		}
	}

	table.Render()
}
