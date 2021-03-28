package keeper

import (
	"github.com/benwolfaardt/lending/x/lending/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeeper_CreateDebtWithAvailableID(t *testing.T) {
	debtor, _ := sdk.AccAddressFromBech32("cosmos1lwmppctrr6ssnrmuyzu554dzf50apkfvd53jx0")
	creditor, _ := sdk.AccAddressFromBech32("cosmos1tupew4x3rhh0lpqha9wvzmzxjr4e37mfy3qefm")

	debt1 := types.Debt{
		ID:			"A1",
		Debtor:		debtor,
		Amount:		sdk.NewCoin("foo", sdk.NewInt(20000)),
		Creditor: 	creditor,
	}
	_, ctx, _, _, keeper := SetupTestInput()
	err := keeper.CreateDebt(ctx, debt1)
	require.NoError(t, err)

	debt2 := types.Debt{
		ID:			"A2",
		Debtor:		debtor,
		Amount:		sdk.NewCoin("foo", sdk.NewInt(20000)),
		Creditor: 	creditor,
	}
	err = keeper.CreateDebt(ctx, debt2)

	require.NoError(t, err)
}

func TestKeeper_CreateDebtWithUnavailableID(t *testing.T) {
	debtor, _ := sdk.AccAddressFromBech32("cosmos1lwmppctrr6ssnrmuyzu554dzf50apkfvd53jx0")
	creditor, _ := sdk.AccAddressFromBech32("cosmos1tupew4x3rhh0lpqha9wvzmzxjr4e37mfy3qefm")

	debt1 := types.Debt{
		ID:       "A1",
		Debtor:   debtor,
		Amount:   sdk.NewCoin("foo", sdk.NewInt(20000)),
		Creditor: creditor,
	}
	_, ctx, _, _, keeper := SetupTestInput()
	err := keeper.CreateDebt(ctx, debt1)
	require.NoError(t, err)

	debt2 := types.Debt{
		ID:       "A1", // same ID as before!
		Debtor:   creditor,
		Amount:   sdk.NewCoin("foo", sdk.NewInt(10000)),
		Creditor: debtor,
	}
	err = keeper.CreateDebt(ctx, debt2)

	require.Error(t, err)
}

func TestKeeper_PayDebt(t *testing.T) {
	debtor, _ := sdk.AccAddressFromBech32("cosmos1lwmppctrr6ssnrmuyzu554dzf50apkfvd53jx0")
	creditor, _ := sdk.AccAddressFromBech32("cosmos1tupew4x3rhh0lpqha9wvzmzxjr4e37mfy3qefm")

	ID := "A1"
	amount := sdk.NewCoin("foo", sdk.NewInt(20000))

	// a slice of test inputs; each input is a struct instance
	tests := []struct {
		name 					string // name of the test
		preExistingDebt 		*types.Debt // original debt, if any
		msgPayDebt 				types.MsgPayDebt // message to send
		startingDebtorBalance 	sdk.Coins // balance of the debtor, if any
		wantErr 				bool // expected error, or not
	}{
		{
		"pay not existing debt",
		nil,
		types.MsgPayDebt{
			ID: 	ID,
			Amount: amount,
			Debtor: debtor,
		},
		nil,
		true,
		},
		{
			"debtor has not enough funds to pay debt",
			&types.Debt{
			ID: 		ID,
			Debtor: 	debtor,
			Amount: 	amount,
			Creditor: 	creditor,
		},
			types.MsgPayDebt{
			ID: 	ID,
			Amount: amount,
			Debtor: debtor,
		},
		nil,
		true,
		},
		{
			"payer is not the debtor",
			&types.Debt{
			ID: 		ID,
			Debtor: 	debtor,
			Amount: 	amount,
			Creditor: 	creditor,
		},
			types.MsgPayDebt{
			ID: 	ID,
			Amount: amount,
			Debtor: creditor,
		},
		nil,
		true,
		},
		{
		"totally pay debt with exact amount",
		&types.Debt{
			ID: 		ID,
			Debtor: 	debtor,
			Amount: 	amount,
			Creditor: creditor,
		},
			types.MsgPayDebt{
			ID: 	ID,
			Amount: amount,
			Debtor: debtor,
		},
			sdk.NewCoins(amount),
			false,
		},
		{
			"totally pay debt with more amount than necessary",
			&types.Debt{
			ID: 		ID,
			Debtor: 	debtor,
			Amount: 	amount,
			Creditor: 	creditor,
		},
			types.MsgPayDebt{
			ID: 	ID,
			Amount: sdk.NewCoin("foo", sdk.NewInt(40000)),
			Debtor: debtor,
		},
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(50000))),
			false,
		},
		{
		"partially pay debt",
		&types.Debt{
		ID: 		ID,
		Debtor: 	debtor,
		Amount: 	amount,
		Creditor: 	creditor,
		},
		types.MsgPayDebt{
		ID: 	ID,
		Amount: sdk.NewCoin("foo", sdk.NewInt(10000)),
		Debtor: debtor,
		},
			sdk.NewCoins(amount),
			false,
		},
	}
	// test runner
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// the message msut be syntactically correct
			require.NoError(t, tt.msgPayDebt.ValidateBasic())
			_, ctx, authKeeper, bankKeeper, keeper := SetupTestInput()
			// if there is a preexisting debt:
			if tt.preExistingDebt != nil {
				// it must be syntactically correct
				require.NoError(t, tt.preExistingDebt.Validate())
				// we create it in store
				require.NoError(t, keeper.CreateDebt(ctx, *tt.preExistingDebt))
				// its creditor initially has no coins
				require.NoError(t, bankKeeper.SetCoins(ctx, tt.preExistingDebt.Creditor, sdk.NewCoins()))
			}
			if tt.startingDebtorBalance == nil {
				tt.startingDebtorBalance = sdk.NewCoins()
			}
			// we set the balance of the debtor as specified in the test input
			require.NoError(t, bankKeeper.SetCoins(ctx, tt.msgPayDebt.Debtor,
				tt.startingDebtorBalance))
			// we call the actual pay debt operation to test
			err := keeper.PayDebt(ctx, tt.msgPayDebt)
			if tt.wantErr {
				// if an error was expected, if must have really occurred
				require.Error(t, err)
				if tt.preExistingDebt != nil {
					// the creditor did not get any coin
					creditorAccount := authKeeper.GetAccount(ctx, creditor)
					require.True(t, creditorAccount.GetCoins().Empty())
				}
				debtorAccount := authKeeper.GetAccount(ctx, tt.msgPayDebt.Debtor)
				// the debtor kept its coins unchanged
				require.True(t, debtorAccount.GetCoins().IsEqual(tt.startingDebtorBalance))
				return
			}
			// if no error was expected
			creditorAccount := authKeeper.GetAccount(ctx, tt.preExistingDebt.Creditor)
			// the creditor git its coins
			require.True(t, creditorAccount.GetCoins().IsEqual(sdk.NewCoins(tt.msgPayDebt.Amount)))
			debtorAccount := authKeeper.GetAccount(ctx, tt.preExistingDebt.Debtor)
			// the debtor paid as expected
			require.True(t, debtorAccount.GetCoins().IsEqual(tt.startingDebtorBalance.Sub(sdk.NewCoins(tt.msgPayDebt.Amount))))
			// the debt is still in store
			newDebt, err := keeper.getDebtByID(ctx, tt.msgPayDebt.ID)
			require.NoError(t, err)
			expectedAmount := sdk.NewCoin(tt.preExistingDebt.Amount.Denom, sdk.NewInt(0))
			if tt.msgPayDebt.Amount.IsLT(tt.preExistingDebt.Amount) {
				expectedAmount = tt.preExistingDebt.Amount.Sub(tt.msgPayDebt.Amount)
			}
			// the debt in store has been updated to the remaining amount
			require.True(t, newDebt.Amount.IsEqual(expectedAmount))
		})
	}
}