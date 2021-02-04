package full

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.uber.org/fx"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/builtin/token"
	"github.com/filecoin-project/lotus/chain/actors/builtin/bounty"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
)

type BountyAPI struct {
	fx.In

	Chain    *store.ChainStore
	StateAPI StateAPI
	MpoolAPI MpoolAPI
}

var _ api.BountyAPI = (*BountyAPI)(nil)

func (b *BountyAPI) BountyInfo(ctx context.Context, bountyAddr address.Address) (*bounty.Info, error) {
	actor, err := b.StateAPI.StateGetActor(ctx, bountyAddr, types.EmptyTSK)
	if err != nil {
		return nil, fmt.Errorf("failed to load bounty actor at address %s: %w", bountyAddr, err)
	}

	state, err := bounty.Load(b.Chain.Store(ctx), actor)
	if err != nil {
		return nil, fmt.Errorf("failed to load actor state: %w", err)
	}

	return state.BountyInfo()
}

func (b *BountyAPI) BountyCreate(ctx context.Context, creator address.Address, pieceCid cid.Cid, token *address.Address, from address.Address, value abi.TokenAmount, duration abi.ChainEpoch, bounties uint64) (cid.Cid, error) {
	return b.pushMessage(ctx, creator, func(mb bounty.MessageBuilder) (*types.Message, error) {
		return mb.Create(pieceCid, token, from, value, duration, bounties)
	})
}

func (b *BountyAPI) BountyClaim(ctx context.Context, bountyAddr address.Address, from address.Address, newDealId *abi.DealID) (cid.Cid, error) {
	_, err := b.StateAPI.StateGetActor(ctx, bountyAddr, types.EmptyTSK)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to load bounty actor at address %s: %w", bountyAddr, err)
	}

	return b.pushMessage(ctx, from, func(mb bounty.MessageBuilder) (*types.Message, error) {
		return mb.Claim(bountyAddr, newDealId)
	})
}

func (b *BountyAPI) pushMessage(ctx context.Context, from address.Address, fn func(mb bounty.MessageBuilder) (*types.Message, error)) (cid.Cid, error) {
	nver, err := b.StateAPI.StateNetworkVersion(ctx, types.EmptyTSK)
	if err != nil {
		return cid.Undef, err
	}

	builder := bounty.Message(actors.VersionForNetwork(nver), from)
	msg, err := fn(builder)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to build message: %w", err)
	}

	// send the message out to the network
	smsg, err := b.MpoolAPI.MpoolPushMessage(ctx, msg, nil)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to push message: %w", err)
	}

	return smsg.Cid(), nil
}

func (b *BountyAPI) messageBuilder(ctx context.Context, from address.Address) (token.MessageBuilder, error) {
	nver, err := b.StateAPI.StateNetworkVersion(ctx, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	return token.Message(actors.VersionForNetwork(nver), from), nil
}
