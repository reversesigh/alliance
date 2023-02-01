package alliance

import (
	"fmt"
	"time"

	"github.com/terra-money/alliance/x/alliance/keeper"
	"github.com/terra-money/alliance/x/alliance/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// EndBlocker
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Re-delegations are stored for the entire duration of the staking module's unbonding period. This is so that
	// delegators will not be able to escape a slashing event by re-delegating away from an offending validator.
	// This hook checks that the re-delegations are mature and deletes the re-delegation event from the store.
	k.CompleteRedelegations(ctx)

	// Un-delegations are held for the entire duration of the staking module's unbonding period.
	// This hook checks that the un-delegations are mature and once mature, returns the bonded tokens back to the delegator.
	if err := k.CompleteUndelegations(ctx); err != nil {
		panic(fmt.Errorf("failed to complete undelegations from x/alliance module: %s", err))
	}

	assets := k.GetAllAssets(ctx)

	// Deduct assets hook applies the take rate on alliance assets.
	if _, err := k.DeductAssetsHook(ctx, assets); err != nil {
		panic(fmt.Errorf("failed to apply take rate from alliance in x/alliance module: %s", err))
	}

	// Updates the reward weight based on the alliance asset configuration.
	k.RewardWeightChangeHook(ctx, assets)
	if err := k.RebalanceHook(ctx, assets); err != nil {
		panic(fmt.Errorf("failed to rebalance assets in x/alliance module: %s", err))
	}
	return []abci.ValidatorUpdate{}
}
