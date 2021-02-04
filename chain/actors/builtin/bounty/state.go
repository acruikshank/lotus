package bounty

import (
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/go-state-types/cbor"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	bounty3 "github.com/filecoin-project/specs-actors/v3/actors/builtin/bounty"

	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
)

func init() {
	builtin.RegisterActorState(builtin3.BountyActorCodeID, func(store adt.Store, root cid.Cid) (cbor.Marshaler, error) {
		return load3(store, root)
	})
}

// BountyINfo is returned by the BountyInfo() method. Currently an alias to the
// appropriate actor return type.
type Info = bounty3.State
type DealBounty = bounty3.DealBounty

// Load returns an abstract copy of the bounty3 actor state, regardless of
// the actor version.
func Load(store adt.Store, act *types.Actor) (State, error) {
	switch act.Code {
	case builtin3.BountyActorCodeID:
		return load3(store, act.Head)
	}
	return nil, xerrors.Errorf("unknown actor code %s", act.Code)
}

// State is an abstract version of the token3 actor's state that works across
// versions.
type State interface {
	cbor.Marshaler

	BountyInfo() (*Info, error)
}
