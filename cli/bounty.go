package cli

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	init3 "github.com/filecoin-project/specs-actors/v3/actors/builtin/init"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/build"

	"github.com/urfave/cli/v2"
)

var bountyCmd = &cli.Command{
	Name:  "bounty",
	Usage: "Interact with bounty actors",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "confidence",
			Usage: "number of block confirmations to wait for",
			Value: int(build.MessageConfidence),
		},
	},
	Subcommands: []*cli.Command{
		bountyCreateCmd,
		bountyInfoCmd,
		bountyClaimCmd,
	},
}

var bountyCreateCmd = &cli.Command{
	Name:  "create",
	Usage: "Create a new bounty actor",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "piece-cid",
			Usage: "the cid of the piece this bounty will reward",
		},
		&cli.StringFlag{
			Name:     "token",
			Usage:    "optional token actor address to pay bounty in something other than filecoin",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "from",
			Usage:    "if token supplied, specifies token account from which bounty will pay",
			Required: false,
		},
		&cli.Uint64Flag{
			Name:     "value",
			Usage:    "total value of bounty (in FIL or token specified)",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:  "duration",
			Usage: "duration in chain epochs for which bounty will be supported",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:     "bounties",
			Usage:    "number of deals that may be rewarded for bounty at same time (replication factor)",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		w := cctx.App.Writer

		api, closer, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		var (
			confidence = uint64(cctx.Int("confidence"))

			pieceStr     = cctx.String("piece-cid")
			tokenStr     = cctx.String("token")
			fromStr     = cctx.String("from")
			value = cctx.Uint64("value")
			duration = cctx.Uint64("duration")
			bounties   = cctx.Uint64("bounties")

			token  *address.Address
			from address.Address
		)

		pieceCid, err := cid.Decode(pieceStr)
		if err != nil {
			return fmt.Errorf("failed to parse piece cid %s: %w", pieceStr, err)
		}

		sender, err := api.WalletDefaultAddress(ctx)
		if err != nil {
			return fmt.Errorf("failed to get wallet default address: %w", err)
		}

		if tokenStr == "" {
			from, err = api.WalletDefaultAddress(ctx)
			if err != nil {
				return fmt.Errorf("failed to get wallet default address: %w", err)
			}
		} else {
			t, err := address.NewFromString(tokenStr)
			if err != nil {
				return fmt.Errorf("failed to parse owner address: %w", err)
			}
			token = &t

			from, err = address.NewFromString(fromStr)
			if err != nil {
				return fmt.Errorf("failed to parse owner address: %w", err)
			}
		}

		_, _ = fmt.Fprintf(w, "creating bounty for %s with value %d, duration %d, and bounties %d\n", pieceStr, value, duration, bounties)

		mcid, err := api.BountyCreate(ctx, sender, pieceCid, token, from, abi.NewTokenAmount(int64(value)), abi.ChainEpoch(duration), bounties)
		if err != nil {
			return fmt.Errorf("bounty creation failed: %w", err)
		}

		_, _ = fmt.Fprintf(w, "message CID: %s\n", mcid)

		// wait for it to get mined into a block
		result, err := api.StateWaitMsg(ctx, mcid, confidence)
		if err != nil {
			return fmt.Errorf("failed to wait for message: %w", err)
		}

		if err = processResult(w, result); err != nil {
			return err
		}

		// get address of newly created bounty
		var ret init3.ExecReturn
		if err := ret.UnmarshalCBOR(bytes.NewReader(result.Receipt.Return)); err != nil {
			return err
		}

		_, _ = fmt.Fprintln(cctx.App.Writer, "created new bounty actor: ", ret.IDAddress, ret.RobustAddress)
		return nil
	},
}

var bountyInfoCmd = &cli.Command{
	Name:  "info",
	Usage: "Retrieve the basic info of a bounty actor",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "bounty",
			Usage:    "bounty actor address",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		t := cctx.String("bounty")
		addr, err := address.NewFromString(t)
		if err != nil {
			return fmt.Errorf("failed to parse address %s: %w", t, err)
		}

		api, closer, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		info, err := api.BountyInfo(ctx, addr)
		if err != nil {
			return fmt.Errorf("failed to retrieve bounty info: %w", err)
		}

		output, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("failed to serialize bounty info into JSON: %w", err)
		}

		_, _ = fmt.Fprintln(cctx.App.Writer, string(output))
		return nil
	},
}

var bountyClaimCmd = &cli.Command{
	Name:      "claim",
	Usage:     "Claims that existing deals have earned the bounty and optionally adds a new deal",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "bounty",
			Usage:    "bounty actor address",
			Required: true,
		},
		&cli.Int64Flag{
			Name:  "deal-id",
			Usage: "if provided, attempt to add the new deal as a bounty recipient",
			Required: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		w := cctx.App.Writer

		api, closer, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		ctx := ReqContext(cctx)

		var (
			confidence = uint64(cctx.Int("confidence"))
			newDealID *abi.DealID
		)

		bounty, err := address.NewFromString(cctx.String("bounty"))
		if err != nil {
			return fmt.Errorf("failed to parse bounty address: %w", err)
		}

		if c := cctx.String("deal-id"); c != "" {
			dealID := abi.DealID(cctx.Uint64("deal-id"))
			newDealID = &dealID
		}

		sender, err := api.WalletDefaultAddress(ctx)
		if err != nil {
			return fmt.Errorf("failed to get wallet default address: %w", err)
		}

		var mcid cid.Cid

		_, _ = fmt.Fprintf(w, "making claim on bounty %v\n", bounty)

		mcid, err = api.BountyClaim(ctx, bounty, sender, newDealID)
		if err != nil {
			return fmt.Errorf("transfer from failed: %w", err)
		}

		_, _ = fmt.Fprintf(w, "message CID: %s\n", mcid)
		_, _ = fmt.Fprintf(w, "awaiting %d confirmations...\n", confidence)

		// wait for it to get mined into a block
		result, err := api.StateWaitMsg(ctx, mcid, confidence)
		if err != nil {
			return fmt.Errorf("failed to wait for message: %w", err)
		}

		return processResult(w, result)
	},
}
