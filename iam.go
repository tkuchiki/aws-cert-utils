package certutils

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/olekukonko/tablewriter"
)

type IAM struct {
	client *iam.IAM
}

type IAMDescription struct {
	name string
	id   string
	path string
	arn  string
}

func NewIAM(sess *session.Session) *IAM {
	return &IAM{
		client: iam.New(sess),
	}
}

func createIAMUploadServerCertificateInput(body, chain, pkey, path, name string) *iam.UploadServerCertificateInput {
	input := &iam.UploadServerCertificateInput{}

	input.SetCertificateBody(body)
	input.SetCertificateChain(chain)
	input.SetPath(path)
	input.SetPrivateKey(pkey)
	input.SetServerCertificateName(name)

	return input
}

func (i *IAM) Upload(cert, chain, pkey []byte, path, name string) (string, error) {
	out, err := i.client.UploadServerCertificate(createIAMUploadServerCertificateInput(string(cert), string(chain), string(pkey), path, name))

	return fmt.Sprintf("Uploaded %s %s", name, *out.ServerCertificateMetadata.Arn), err
}

func createIAMListServerCertificatesInput(marker string, maxItems int64, path string) *iam.ListServerCertificatesInput {
	input := &iam.ListServerCertificatesInput{}

	if marker != "" {
		input.SetMarker(marker)
	}

	if maxItems > 0 {
		input.SetMaxItems(maxItems)
	}

	if path != "" {
		input.SetPathPrefix(path)
	}

	return input
}

func (i *IAM) List(marker string, maxItems int64, path string) ([]IAMDescription, error) {
	out, err := i.client.ListServerCertificates(createIAMListServerCertificatesInput(marker, maxItems, path))
	if err != nil {
		return []IAMDescription{}, err
	}

	descs := make([]IAMDescription, 0, len(out.ServerCertificateMetadataList))
	for _, metadata := range out.ServerCertificateMetadataList {
		desc := IAMDescription{
			name: *metadata.ServerCertificateName,
			id:   *metadata.ServerCertificateId,
			path: *metadata.Path,
			arn:  *metadata.Arn,
		}
		descs = append(descs, desc)
	}

	return descs, err
}

func (i *IAM) ListNames(marker string, maxItems int64, path string) ([]string, error) {
	descs, err := i.List(marker, maxItems, path)
	if err != nil {
		return []string{}, err
	}

	names := make([]string, 0, len(descs))
	for _, desc := range descs {
		names = append(names, desc.name)
	}

	return names, err
}

func createIAMUpdateServerCertificateInput(newPath, newName, name string) *iam.UpdateServerCertificateInput {
	input := &iam.UpdateServerCertificateInput{}

	input.SetNewPath(newPath)
	input.SetNewServerCertificateName(newName)
	input.SetServerCertificateName(name)

	return input
}

func (i *IAM) Update(newPath, newName, name string) (string, error) {
	_, err := i.client.UpdateServerCertificate(createIAMUpdateServerCertificateInput(newPath, newName, name))

	return fmt.Sprintf("Updated %s -> %s", name, newName), err
}

func (i *IAM) Delete(name string) (string, error) {
	input := &iam.DeleteServerCertificateInput{ServerCertificateName: aws.String(name)}
	_, err := i.client.DeleteServerCertificate(input)

	return fmt.Sprintf("Deleted %s", name), err
}

func (i *IAM) ReadableList(descs []IAMDescription) {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Name", "ID", "Path", "Arn"})

	for _, desc := range descs {
		table.Append([]string{desc.name, desc.id, desc.path, desc.arn})
	}

	table.Render()
}
