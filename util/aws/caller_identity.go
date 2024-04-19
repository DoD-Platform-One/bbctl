package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// CallerIdentity holds the relevant output of the GetCallerIdentity function of the AWS SDK for STS
type CallerIdentity struct {
	sts.GetCallerIdentityOutput

	Username string
}
