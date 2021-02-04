package bounty

import (
	"fmt"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"

	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"

	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/types"
)

var Methods = builtin3.MethodsMultisig

func Message(version actors.Version, from address.Address) MessageBuilder {
	switch version {
	case actors.Version3:
		return message3{from}
	default:
		panic(fmt.Sprintf("unsupported actors version: %d", version))
	}
}

type MessageBuilder interface {
	// Create produces a message to construct a new token actor.
	Create(pieceCid cid.Cid, token *address.Address, from address.Address, value abi.TokenAmount, duration abi.ChainEpoch, bounties uint64) (*types.Message, error)

	// Claim produces a message to make a claim against the bounty with an optional new deal id
	Claim(bountyAddress address.Address, newDealId *abi.DealID) (*types.Message, error)
}
