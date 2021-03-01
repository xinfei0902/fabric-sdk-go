package fabclient

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"

	"../dlog"
)

func MyExecute(cc *channel.Client, async bool, request channel.Request, options ...channel.RequestOption) (ret channel.Response, err error) {
	// channelContext, err := mainSetup.WholeChannelContext()
	// if err != nil {
	// 	return
	// }

	options = append(options, channel.WithTimeout(fab.Execute, 30*time.Second))
	options = append(options, channel.WithRetry(retry.DefaultOpts))
	// options = append(options, channel.WithTargetFilter(filter.NewEndpointFilter(channelContext, filter.EndorsingPeer)))

	if async || mainSetup.AsyncCall {
		return cc.InvokeHandler(AsyncOverHalfExecuteHandler(), request, options...)
	}
	return cc.InvokeHandler(SyncOverHalfExecuteHandler(), request, options...)
}

func AsyncOverHalfExecuteHandler(next ...invoke.Handler) invoke.Handler {
	return NewProposalProcessorHandler(
		NewEndorsementHandler(
			NewEndorsementValidationHandler(
				invoke.NewSignatureValidationHandler(
					// use fix
					MyNewCommitHandler(next...),
				),
			),
		),
	)
}

func SyncOverHalfExecuteHandler(next ...invoke.Handler) invoke.Handler {
	return NewProposalProcessorHandler(
		NewEndorsementHandler(
			// use fix
			NewEndorsementValidationHandler(
				invoke.NewSignatureValidationHandler(
					invoke.NewCommitHandler(next...),
				),
			),
		),
	)
}

////////////////////////////  1111111111111111111111111  /////////////////////////////////////////
//NewProposalProcessorHandler returns a handler that selects proposal processors
func NewProposalProcessorHandler(next ...invoke.Handler) *ProposalProcessorHandler {
	return &ProposalProcessorHandler{next: getNext(next)}
}

//ProposalProcessorHandler for selecting proposal processors
type ProposalProcessorHandler struct {
	next invoke.Handler
}

//Handle selects proposal processors
func (h *ProposalProcessorHandler) Handle(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) {
	//Get proposal processor, if not supplied then use selection service to get available peers as endorser
	if len(requestContext.Opts.Targets) == 0 {
		// var selectionOpts []options.Opt
		// if requestContext.SelectionFilter != nil {
		// 	selectionOpts = append(selectionOpts, selectopts.WithPeerFilter(requestContext.SelectionFilter))
		// }

		// endorsers, err := clientContext.Selection.GetEndorsersForChaincode(newInvocationChain(requestContext), selectionOpts...)

		list, _ := AllTargetPeers("org")

		endorsers := make([]fab.Peer, 0, len(list))

		ctx, err := mainSetup.WholeChannelContext()
		if err != nil {
			requestContext.Error = errors.WithMessage(err, "Failed to get endorsing peers")
			return
		}

		for _, v := range list {
			if len(v) == 0 || len(v[0]) == 0 {
				continue
			}
			one, err := getPeer(ctx, v[0])
			if err != nil {
				requestContext.Error = errors.WithMessage(err, "Failed to get endorsing peers")
				return
			}
			endorsers = append(endorsers, one)
		}

		if err != nil {
			requestContext.Error = errors.WithMessage(err, "Failed to get endorsing peers")
			return
		}
		requestContext.Opts.Targets = endorsers
	}

	//Delegate to next step if any
	if h.next != nil {
		h.next.Handle(requestContext, clientContext)
	}
}

func newInvocationChain(requestContext *invoke.RequestContext) []*fab.ChaincodeCall {
	invocChain := []*fab.ChaincodeCall{{ID: requestContext.Request.ChaincodeID}}
	for _, ccCall := range requestContext.Request.InvocationChain {
		if ccCall.ID == invocChain[0].ID {
			invocChain[0].Collections = ccCall.Collections
		} else {
			invocChain = append(invocChain, ccCall)
		}
	}
	return invocChain
}

/////////////////////////////  2222222222222222222222222  ///////////////////////////////

//NewEndorsementHandler returns a handler that endorses a transaction proposal
func NewEndorsementHandler(next ...invoke.Handler) *EndorsementHandler {
	return &EndorsementHandler{next: getNext(next)}
}

//EndorsementHandler for handling endorse transactions
type EndorsementHandler struct {
	next invoke.Handler
}

//Handle for endorsing transactions
func (e *EndorsementHandler) Handle(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) {

	if len(requestContext.Opts.Targets) == 0 {
		requestContext.Error = status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "targets were not provided", nil)
		return
	}

	// Endorse Tx
	transactionProposalResponses, proposal, err := createAndSendTransactionProposal(clientContext.Transactor, &requestContext.Request, peer.PeersToTxnProcessors(requestContext.Opts.Targets))

	requestContext.Response.Proposal = proposal
	requestContext.Response.TransactionID = proposal.TxnID // TODO: still needed?

	if err != nil {
		if len(transactionProposalResponses) == 0 {
			requestContext.Error = err
			return
		}

		tmp := make([]fab.Peer, 0, len(requestContext.Opts.Targets))

	Loop:
		for _, i := range requestContext.Opts.Targets {
			for _, j := range transactionProposalResponses {
				fmt.Println(j.Endorser, i.URL())
				if j.Endorser == i.URL() {
					tmp = append(tmp, i)

					continue Loop
				}
			}

			// transactionProposalResponses = append(transactionProposalResponses,
			// 	&fab.TransactionProposalResponse{
			// 		Endorser: err.Error(),
			// 	})
		}

		fmt.Println(len(requestContext.Opts.Targets), len(tmp))
		requestContext.Opts.Targets = tmp
	}

	requestContext.Response.Responses = transactionProposalResponses
	if len(transactionProposalResponses) > 0 {
		requestContext.Response.Payload = transactionProposalResponses[0].ProposalResponse.GetResponse().Payload
		requestContext.Response.ChaincodeStatus = transactionProposalResponses[0].ChaincodeStatus
	}

	//Delegate to next step if any
	if e.next != nil {
		e.next.Handle(requestContext, clientContext)
	}
}

func createAndSendTransactionProposal(transactor fab.ProposalSender, chrequest *invoke.Request, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, error) {
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID:  chrequest.ChaincodeID,
		Fcn:          chrequest.Fcn,
		Args:         chrequest.Args,
		TransientMap: chrequest.TransientMap,
	}

	txh, err := transactor.CreateTransactionHeader()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction header failed")
	}

	proposal, err := txn.CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction proposal failed")
	}

	transactionProposalResponses, err := transactor.SendTransactionProposal(proposal, targets)

	return transactionProposalResponses, proposal, err
}

///////////////////////////////  33333333333333333333333 ///////////////////////////

func NewEndorsementValidationHandler(next ...invoke.Handler) invoke.Handler {
	return &EndorsementValidationHandler{next: getNext(next)}
}

//EndorsementValidationHandler for transaction proposal response filtering
type EndorsementValidationHandler struct {
	next invoke.Handler
}

//Handle for Filtering proposal response
func (f *EndorsementValidationHandler) Handle(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) {

	list, _ := AllTargetPeers("org")
	//Filter tx proposal responses
	pair, err := f.validate(requestContext.Response.Responses, len(list))
	if err != nil {
		requestContext.Error = errors.WithMessage(err, "endorsement validation failed")
		return
	}

	requestContext.Response.Responses = pair

	if len(pair) > 0 {
		requestContext.Response.Payload = pair[0].ProposalResponse.GetResponse().Payload
		requestContext.Response.ChaincodeStatus = pair[0].ChaincodeStatus
	}

	//Delegate to next step if any
	if f.next != nil {
		f.next.Handle(requestContext, clientContext)
	}
}

func (f *EndorsementValidationHandler) validate(txProposalResponse []*fab.TransactionProposalResponse, count int) (ret []*fab.TransactionProposalResponse, err error) {

	HashTable := make(map[string][]*fab.TransactionProposalResponse)

	var errResp error
	var errCount int

	for _, r := range txProposalResponse {
		if r == nil || r.ProposalResponse == nil {
			errCount++
			if errResp == nil && len(r.Endorser) > 0 {
				errResp = errors.New(r.Endorser)
			}
			continue
		}
		response := r.ProposalResponse.GetResponse()
		if response.Status < int32(common.Status_SUCCESS) || response.Status >= int32(common.Status_BAD_REQUEST) {
			errCount++
			if errResp == nil {
				errResp = status.NewFromProposalResponse(r.ProposalResponse, r.Endorser)
			}
			continue
		}

		hashOpt := sha256.New()
		_, err = hashOpt.Write(r.ProposalResponse.Payload)
		if err != nil {
			// memory limit error
			return nil, err
		}
		_, err = hashOpt.Write(response.Payload)
		if err != nil {
			// memory limit error
			return nil, err
		}

		key := string(hashOpt.Sum(nil))

		pair, ok := HashTable[key]
		if !ok {
			HashTable[key] = make([]*fab.TransactionProposalResponse, 0, count)
		}

		pair = append(pair, r)
		HashTable[key] = pair
	}

	threshold := mainSetup.Threshold
	logOpt := dlog.DebugLog("validate", dlog.DebugGetSerialNumber()).WithField("threshold", threshold).WithField("total", count)

	if threshold < 1 {
		if float64(errCount)/float64(count) > threshold {
			logOpt.WithField("count", errCount).WithError(errResp).Infoln("validate failed")
			return nil, errResp
		}

		for _, pair := range HashTable {
			if float64(len(pair))/float64(count) > threshold {
				logOpt.WithField("count", len(pair)).Infoln("validate")
				return pair, nil
			}
		}

	} else {
		if float64(errCount) > threshold {
			logOpt.WithField("count", errCount).WithError(errResp).Infoln("validate failed")
			return nil, errResp
		}

		for _, pair := range HashTable {
			if float64(len(pair)) > threshold {
				logOpt.WithField("count", len(pair)).Infoln("validate")
				return pair, nil
			}
		}
	}

	return nil, status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(),
		"ProposalResponsePayloads do not match", nil)
}

//////////////////////////////////  555555555555555///////////////////

func MyNewCommitHandler(next ...invoke.Handler) invoke.Handler {
	return &AsyncCommitTxHandler{next: getNext(next)}
}

func getNext(next []invoke.Handler) invoke.Handler {
	if len(next) > 0 {
		return next[0]
	}
	return nil
}

//AsyncCommitTxHandler for committing transactions
type AsyncCommitTxHandler struct {
	next invoke.Handler
}

//Handle handles commit tx
func (c *AsyncCommitTxHandler) Handle(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) {
	txnID := requestContext.Response.TransactionID

	start := time.Now().Unix()

	_, err := createAndSendTransaction(clientContext.Transactor, requestContext.Response.Proposal, requestContext.Response.Responses)
	end := time.Now().Unix()
	if err != nil {
		dlog.DebugLog(string(txnID), 0).Debug(start, end, "error", end-start)
		requestContext.Error = errors.Wrap(err, "CreateAndSendTransaction failed")
		return
	}
	dlog.DebugLog(string(txnID), 0).Debug(start, end, "success", end-start)

	//Delegate to next step if any
	if c.next != nil {
		c.next.Handle(requestContext, clientContext)
	}
}

func createAndSendTransaction(sender fab.Sender, proposal *fab.TransactionProposal, resps []*fab.TransactionProposalResponse) (*fab.TransactionResponse, error) {

	txnRequest := fab.TransactionRequest{
		Proposal:          proposal,
		ProposalResponses: resps,
	}

	tx, err := sender.CreateTransaction(txnRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := sender.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}

	return transactionResponse, nil
}
