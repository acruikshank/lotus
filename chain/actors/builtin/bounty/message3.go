package bounty

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	init0 "github.com/filecoin-project/specs-actors/actors/builtin/init"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	"github.com/filecoin-project/specs-actors/v3/actors/builtin/bounty"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/chain/actors"
	init_ "github.com/filecoin-project/lotus/chain/actors/builtin/init"
	"github.com/filecoin-project/lotus/chain/types"
)

type message3 struct {
	from address.Address
}

var _ MessageBuilder = message3{}

// Create produces a message to construct a new token actor.
func (m message3) Create(pieceCid cid.Cid, token *address.Address, from address.Address, value abi.TokenAmount, duration abi.ChainEpoch, bounties uint64) (*types.Message, error) {
	params := &bounty.ConstructorParams{
		PieceCid: pieceCid,
		Token:    token,
		From:     from,
		Value:    value,
		Duration: duration,
		Bounties: bounties,
	}

	enc, err := actors.SerializeParams(params)
	if err != nil {
		return nil, err
	}

	execParams := &init0.ExecParams{
		CodeCID:           builtin3.BountyActorCodeID,
		ConstructorParams: enc,
	}

	enc, err = actors.SerializeParams(execParams)
	if err != nil {
		return nil, err
	}

	return &types.Message{
		To:     init_.Address,
		From:   m.from,
		Method: builtin0.MethodsInit.Exec,
		Params: enc,
		Value:  big.Zero(),
	}, nil
}

// Claim produces a message to make a claim against the bounty with an optional new deal id
func (m message3) Claim(bountyAddr address.Address, newDealId *abi.DealID) (*types.Message, error) {
	params := &bounty.ClaimParams{
		NewDealID: newDealId,
	}

	enc, err := actors.SerializeParams(params)
	if err != nil {
		return nil, err
	}

	return &types.Message{
		To:     bountyAddr,
		From:   m.from,
		Method: builtin3.MethodsToken.Transfer,
		Params: enc,
		Value:  big.Zero(),
	}, nil
}
