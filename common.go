package certutils

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/tkuchiki/aws-sdk-go-config"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

const minPrivateKeyBitLength = 1024
const maxPrivateKeyBitLength = 2048

type Tag struct {
	Key   string
	Value string
}

type CertificateManager struct {
	Cert  []byte
	Chain []byte
	Pkey  []byte
}

func NewCertificateManager() *CertificateManager {
	return &CertificateManager{}
}

func (cm *CertificateManager) LoadCertificate(cert, certPath string) error {
	var err error
	cm.Cert, err = GetCertificateData(cert, certPath)
	return err
}

func (cm *CertificateManager) LoadPrivateKey(pkey, pkeyPath string) error {
	var err error
	cm.Pkey, err = GetCertificateData(pkey, pkeyPath)
	return err
}

func (cm *CertificateManager) LoadChain(chain, chainPath string) error {
	var err error
	cm.Chain, err = GetCertificateData(chain, chainPath)
	return err
}

func (cm *CertificateManager) CheckPrivateKeyBitLen() error {
	bit, err := PrivateKeyBitLen(cm.Cert, cm.Pkey)
	if err != nil {
		return err
	}
	return CheckPrivateKeyBitLen(bit)
}

func NewAWSSession(accessKey, secretKey, arn, token, region, profile, config, creds string) (*session.Session, error) {
	conf := awsconfig.Option{
		Arn:         arn,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Region:      region,
		Token:       token,
		Profile:     profile,
		Config:      config,
		Credentials: creds,
	}

	return awsconfig.NewSession(conf)
}

func readFile(fpath string) ([]byte, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(f)
}

func GetCertificateData(data, fpath string) ([]byte, error) {
	if fpath != "" {
		return readFile(fpath)
	}

	return []byte(data), nil
}

func PrivateKeyBitLen(certBlock, keyBlock []byte) (int, error) {
	cert, err := tls.X509KeyPair(certBlock, keyBlock)
	if err != nil {
		return 0, err
	}

	var bit int
	switch privateKey := cert.PrivateKey.(type) {
	case *rsa.PrivateKey:
		bit = privateKey.N.BitLen()
	case *ecdsa.PrivateKey:
		bit = privateKey.Curve.Params().BitSize
	default:
		return 0, fmt.Errorf("unsupported private key")
	}

	return bit, nil
}

func CheckPrivateKeyBitLen(bit int) error {
	if bit > maxPrivateKeyBitLength {
		return fmt.Errorf("Invalid private key length (%d bit). AWS supports %d and %d bit RSA private key", bit, minPrivateKeyBitLength, maxPrivateKeyBitLength)
	}

	return nil
}

func SplitStatuses(s string) []string {
	if strings.ToUpper(s) == "ALL" {
		return []string{
			"PENDING_VALIDATION",
			"ISSUED",
			"INACTIVE",
			"EXPIRED",
			"VALIDATION_TIMED_OUT",
			"REVOKED",
			"FAILED",
		}
	}

	splited := strings.Split(s, ",")

	statuses := make([]string, 0, len(splited))
	for _, val := range splited {
		statuses = append(statuses, strings.ToUpper(strings.TrimSpace(val)))
	}

	return statuses
}

func CheckTagValuePattern(val string) error {
	if val == "" {
		return nil
	}

	pattern := `[\p{L}\p{Z}\p{N}_.:\/=+\-@]*`
	re := regexp.MustCompile(pattern)

	group := re.FindStringSubmatch(val)
	if len(group) > 0 && group[0] == val {
		return nil
	}

	return fmt.Errorf("Invalid tag value. Tag value supports %s", pattern)
}

func Choice(choices []string, msg string, pagesize int) string {
	val := ""
	prompt := &survey.Select{
		Message:  msg,
		Options:  choices,
		PageSize: pagesize,
	}
	survey.AskOne(prompt, &val, nil)

	return val
}

func toFlatten(strs []*string) string {
	return strings.Join(aws.StringValueSlice(strs), " ")
}

func dryRunMsg() []string {
	return []string{
		"# Dry run mode",
		"",
	}
}
