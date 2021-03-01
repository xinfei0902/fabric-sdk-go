package fabclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func MyQuery(cc *channel.Client, request channel.Request, options ...channel.RequestOption) (ret channel.Response, err error) {
	options = append(options, channel.WithTimeout(fab.Query, 30*time.Second))
	options = append(options, channel.WithRetry(retry.DefaultOpts))

	return cc.InvokeHandler(NewQueryHandler(), request, options...)
}

//NewQueryHandler returns query handler with EndorseTxHandler & EndorsementValidationHandler Chained
func NewQueryHandler(next ...invoke.Handler) invoke.Handler {
	return NewProposalProcessorHandler(
		NewEndorsementHandler(
			NewEndorsementValidationHandler(
				invoke.NewSignatureValidationHandler(next...),
			),
		),
	)
}
