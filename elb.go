package certutils

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/olekukonko/tablewriter"
)

type ELB struct {
	client *elb.ELB
}

type ELBDescription struct {
	name    string
	dnsname string
	certs   []ELBCertificate
}

type ELBCertificate struct {
	arn  string
	port int64
}

func NewELB(sess *session.Session) *ELB {
	return &ELB{
		client: elb.New(sess),
	}
}

func (e *ELB) getDescriptions(marker string, certFilter string) ([]ELBDescription, error) {
	input := &elb.DescribeLoadBalancersInput{}
	out, err := e.client.DescribeLoadBalancers(input)
	descs := make([]ELBDescription, 0, len(out.LoadBalancerDescriptions))
	for _, desc := range out.LoadBalancerDescriptions {
		elbdesc := ELBDescription{}

		elbdesc.dnsname = *desc.DNSName
		elbdesc.name = *desc.LoadBalancerName

		for _, ld := range desc.ListenerDescriptions {
			l := ld.Listener
			certArn := aws.StringValue(l.SSLCertificateId)
			if certArn != "" {
				if certFilter != "" && certArn != certFilter {
					continue
				}

				elbcert := ELBCertificate{
					arn:  certArn,
					port: *l.LoadBalancerPort,
				}
				elbdesc.certs = append(elbdesc.certs, elbcert)
			}
		}

		if len(elbdesc.certs) < 1 {
			continue
		}

		descs = append(descs, elbdesc)
	}
	return descs, err
}

func (e *ELB) List(certFilter string) ([]ELBDescription, error) {
	return e.getDescriptions("", certFilter)
}

func createELBSetLoadBalancerListenerSSLCertificateInput(name string, port int64, certArn string) *elb.SetLoadBalancerListenerSSLCertificateInput {
	input := &elb.SetLoadBalancerListenerSSLCertificateInput{}

	input.SetLoadBalancerName(name)
	input.SetLoadBalancerPort(port)
	input.SetSSLCertificateId(certArn)

	return input
}

func (e *ELB) getLB(name string) (*elb.DescribeLoadBalancersOutput, error) {
	input := &elb.DescribeLoadBalancersInput{}

	names := []string{name}
	input.SetLoadBalancerNames(aws.StringSlice(names))
	return e.client.DescribeLoadBalancers(input)
}

func getListenerCertificateByPort(out *elb.DescribeLoadBalancersOutput, port int64) (string, error) {
	for _, desc := range out.LoadBalancerDescriptions {
		for _, ld := range desc.ListenerDescriptions {
			l := ld.Listener
			if *l.LoadBalancerPort == port {
				return *l.SSLCertificateId, nil
			}
		}
	}

	return "", fmt.Errorf("Listener not found")
}

func (e *ELB) Update(name string, port int64, certArn string) (string, error) {
	lb, err := e.getLB(name)
	if err != nil {
		return "", err
	}

	destCert, err := getListenerCertificateByPort(lb, port)
	if err != nil {
		return "", err
	}

	_, err = e.client.SetLoadBalancerListenerSSLCertificate(createELBSetLoadBalancerListenerSSLCertificateInput(name, port, certArn))

	return elbUpdateMsg(name, port, destCert, certArn), err
}

func elbUpdateMsg(name string, port int64, src, dest string) string {
	return fmt.Sprintf("Updated %s:%d %s -> %s", name, port, src, dest)
}

func (e *ELB) BulkUpdate(srcCertArn, destCertArn string, dryRun bool) ([]string, error) {
	descs, err := e.getDescriptions("", srcCertArn)
	if err != nil {
		return []string{}, err
	}

	updates := make([]string, 0)
	if dryRun {
		updates = append(updates, dryRunMsg()...)
	}

	for _, desc := range descs {
		for _, cert := range desc.certs {
			if dryRun {
				updates = append(updates, elbUpdateMsg(desc.name, cert.port, srcCertArn, destCertArn))
			} else {
				_, err := e.client.SetLoadBalancerListenerSSLCertificate(createELBSetLoadBalancerListenerSSLCertificateInput(desc.name, cert.port, destCertArn))

				if err != nil {
					return []string{}, err
				}
				updates = append(updates, elbUpdateMsg(desc.name, cert.port, srcCertArn, destCertArn))
			}
		}
	}

	return updates, nil
}

func (e *ELB) ReadableList(descs []ELBDescription) {
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Name", "Port", "Listener SSL Certificate"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	for _, desc := range descs {
		for _, cert := range desc.certs {
			table.Append([]string{desc.name, fmt.Sprint(cert.port), cert.arn})
		}
	}

	table.Render()
}
