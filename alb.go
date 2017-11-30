package certutils

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/olekukonko/tablewriter"
)

type ALB struct {
	client *elbv2.ELBV2
}

type ALBDescription struct {
	name    string
	dnsname string
	certs   []ALBCertificate
}

type ALBCertificate struct {
	arn         string
	port        int64
	listenerArn string
}

func NewALB(sess *session.Session) *ALB {
	return &ALB{
		client: elbv2.New(sess),
	}
}

func createALBDescribeListenersInput(lbArn string) *elbv2.DescribeListenersInput {
	input := &elbv2.DescribeListenersInput{}

	input.SetLoadBalancerArn(lbArn)

	return input
}

func (alb *ALB) getLBs(certFilter string) ([]ALBDescription, error) {
	input := &elbv2.DescribeLoadBalancersInput{}
	out, err := alb.client.DescribeLoadBalancers(input)
	lbs := make([]ALBDescription, 0, len(out.LoadBalancers))
	for _, lb := range out.LoadBalancers {
		albdesc := ALBDescription{}

		albdesc.dnsname = *lb.DNSName
		albdesc.name = *lb.LoadBalancerName

		lout, err := alb.client.DescribeListeners(createALBDescribeListenersInput(*lb.LoadBalancerArn))
		if err != nil {
			return []ALBDescription{}, err
		}

		for _, l := range lout.Listeners {
			for _, cert := range l.Certificates {
				if certFilter != "" && certFilter != *cert.CertificateArn {
					continue
				}

				albcert := ALBCertificate{
					arn:         *cert.CertificateArn,
					port:        *l.Port,
					listenerArn: *l.ListenerArn,
				}
				albdesc.certs = append(albdesc.certs, albcert)
			}
		}

		if len(albdesc.certs) < 1 {
			continue
		}

		lbs = append(lbs, albdesc)
	}
	return lbs, err
}

func (alb *ALB) List(certFilter string) ([]ALBDescription, error) {
	return alb.getLBs(certFilter)
}

func (alb *ALB) getListener(name string) (*elbv2.Listener, error) {
	lbinput := &elbv2.DescribeLoadBalancersInput{}
	names := []string{name}
	lbinput.SetNames(aws.StringSlice(names))
	lbout, err := alb.client.DescribeLoadBalancers(lbinput)
	if err != nil {
		return &elbv2.Listener{}, err
	}

	if len(lbout.LoadBalancers) < 1 {
		return &elbv2.Listener{}, fmt.Errorf("Listener not found")
	}

	linput := &elbv2.DescribeListenersInput{}
	linput.SetLoadBalancerArn(*lbout.LoadBalancers[0].LoadBalancerArn)

	lout, err := alb.client.DescribeListeners(linput)
	if err != nil {
		return &elbv2.Listener{}, err
	}

	if len(lout.Listeners) < 1 {
		return &elbv2.Listener{}, fmt.Errorf("Listener not found")
	}

	return lout.Listeners[0], nil
}

func createALBModifyListenerInput(listenerArn, certArn string) *elbv2.ModifyListenerInput {
	input := &elbv2.ModifyListenerInput{}

	cert := &elbv2.Certificate{}
	cert.SetCertificateArn(certArn)

	certs := []*elbv2.Certificate{cert}

	input.SetCertificates(certs)

	input.SetListenerArn(listenerArn)

	return input
}

func (alb *ALB) Update(name string, certArn string) error {
	l, err := alb.getListener(name)
	if err != nil {
		return err
	}

	_, err = alb.client.ModifyListener(createALBModifyListenerInput(*l.ListenerArn, certArn))

	return err
}

func albUpdateMsg(name string, port int64, src, dest string) string {
	return fmt.Sprintf("Updated %s:%d %s -> %s", name, port, src, dest)
}

func (alb *ALB) BulkUpdate(srcCertArn, destCertArn string, dryRun bool) ([]string, error) {
	lbs, err := alb.getLBs(srcCertArn)
	if err != nil {
		return []string{}, err
	}

	updates := make([]string, 0)
	if dryRun {
		updates = append(updates, dryRunMsg()...)
	}

	for _, lb := range lbs {
		for _, cert := range lb.certs {
			if dryRun {
				updates = append(updates, albUpdateMsg(lb.name, cert.port, srcCertArn, destCertArn))
			} else {
				_, err := alb.client.ModifyListener(createALBModifyListenerInput(cert.listenerArn, destCertArn))

				if err != nil {
					return []string{}, err
				}
				updates = append(updates, albUpdateMsg(lb.name, cert.port, srcCertArn, destCertArn))
			}
		}
	}

	return updates, nil
}

func (alb *ALB) ReadableList(descs []ALBDescription) {
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
