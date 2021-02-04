package bounty

import (
	"github.com/filecoin-project/specs-actors/v3/actors/builtin/bounty"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/chain/actors/adt"
)

var _ State = (*state3)(nil)

// load3 loads the actor state for a token v3 actor.
func load3(store adt.Store, root cid.Cid) (State, error) {
	out := state3{store: store}
	err := store.Get(store.Context(), root, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type state3 struct {
	bounty.State
	store adt.Store
}

func (s *state3) BountyInfo() (*Info, error) {
	return &s.State, nil
}
