package api

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/chain/actors/builtin/bounty"
)

type BountyInfo = bounty.Info

type BountyAPI interface {
	// BountyInfo returns the bounty's information.
	BountyInfo(ctx context.Context, token address.Address) (*BountyInfo, error)

	// BountyCreate creates a new bounty with the specified info.
	BountyCreate(ctx context.Context, creator address.Address, pieceCid cid.Cid, token *address.Address, from address.Address, value abi.TokenAmount, duration abi.ChainEpoch, bounties uint64) (cid.Cid, error)

	// BountyClaim creates a new claim for the bounty which will transfer funds for known deals and optionally add a new deal if provided
	BountyClaim(ctx context.Context, bountyAddr, from address.Address, newDealId *abi.DealID) (cid.Cid, error)
}
