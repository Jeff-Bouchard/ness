package coin

import (
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
)

// helper to build a delegated-hours transaction and corresponding uxIn
func buildDelegatedTx(t *testing.T, headTime uint64, inputHours uint64, maxHours uint64, minInterval uint64, delegateLast uint64, expiry uint64, payoutCoins uint64, payoutHours uint64, ownerCoins uint64) (Transaction, UxArray) {
	t.Helper()

	ownerPub, _ := cipher.GenerateKeyPair()
	ownerAddr := cipher.AddressFromPubKey(ownerPub)

	delegatePub, delegateSec := cipher.GenerateKeyPair()
	delegateAddr := cipher.AddressFromPubKey(delegatePub)

	providerPub, _ := cipher.GenerateKeyPair()
	providerAddr := cipher.AddressFromPubKey(providerPub)

	// single input ux
	ux := UxOut{
		Head: UxHead{Time: headTime, BkSeq: 1},
		Body: UxBody{
			SrcTransaction: cipher.SHA256{},
			Address:        ownerAddr,
			Coins:          ownerCoins,
			Hours:          inputHours,
			Delegate:       delegateAddr,
			MaxHours:       maxHours,
			Expiry:         expiry,
			MinInterval:    minInterval,
			DelegateLast:   delegateLast,
		},
	}

	// outputs: owner gets coins back minus droplet to provider; provider gets payout
	ownerOutCoins := ownerCoins - payoutCoins
	if ownerOutCoins > ownerCoins {
		t.Fatalf("ownerOutCoins underflow")
	}

	// choose owner hours so that fee = inputHours - (ownerHours+payoutHours)
	// keep owner hours non-negative
	var ownerHours uint64
	if inputHours < payoutHours+1 {
		t.Fatalf("invalid inputHours for test setup")
	}
	ownerHours = inputHours - payoutHours - 1 // leave fee=1 hour for burn

	txn := Transaction{Type: TxTypeDelegatedHours}
	must := func(err error) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	must(txn.PushInput(ux.Hash()))
	must(txn.PushOutput(ownerAddr, ownerOutCoins, ownerHours))
	must(txn.PushOutput(providerAddr, payoutCoins, payoutHours))

	// set length/inner hash
	must(txn.UpdateHeader())
	// sign with delegate
	must(txn.SignInput(delegateSec, 0))

	return txn, UxArray{ux}
}

func TestVerifyDelegatedConstraints_Success(t *testing.T) {
	headTime := uint64(1_000_000)
	inputHours := uint64(200)
	txn, uxIn := buildDelegatedTx(t, headTime, inputHours, 100, 86_400, 0, headTime+100_000, 1, 50, 1_000_000)

	if err := txn.VerifyInputSignatures(uxIn); err != nil {
		t.Fatalf("signature verify failed: %v", err)
	}
	if err := txn.VerifyDelegated(headTime, uxIn); err != nil {
		t.Fatalf("delegated verify failed: %v", err)
	}
}

func TestVerifyDelegatedConstraints_MinInterval(t *testing.T) {
	headTime := uint64(1_000_000)
	inputHours := uint64(200)
	delegateLast := headTime - 100 // less than 86400 ago
	txn, uxIn := buildDelegatedTx(t, headTime, inputHours, 100, 86_400, delegateLast, headTime+100_000, 1, 50, 1_000_000)

	if err := txn.VerifyDelegated(headTime, uxIn); err == nil {
		t.Fatalf("expected min interval violation")
	}
}

func TestVerifyDelegatedConstraints_Expiry(t *testing.T) {
	headTime := uint64(1_000_000)
	inputHours := uint64(200)
	expiry := headTime // expired at head time
	txn, uxIn := buildDelegatedTx(t, headTime, inputHours, 100, 0, 0, expiry, 1, 50, 1_000_000)

	if err := txn.VerifyDelegated(headTime, uxIn); err == nil {
		t.Fatalf("expected expiry violation")
	}
}

func TestVerifyDelegatedConstraints_MaxHours(t *testing.T) {
	headTime := uint64(1_000_000)
	inputHours := uint64(200)
	// MaxHours too low for fee+payout (fee=1, payout=50 => consumed=51)
	txn, uxIn := buildDelegatedTx(t, headTime, inputHours, 40, 0, 0, headTime+100_000, 1, 50, 1_000_000)

	if err := txn.VerifyDelegated(headTime, uxIn); err == nil {
		t.Fatalf("expected max hours violation")
	}
}

func TestVerifyDelegatedConstraints_PayoutCoinLimit(t *testing.T) {
	headTime := uint64(1_000_000)
	inputHours := uint64(200)
	// payout coins exceed 1 droplet
	txn, uxIn := buildDelegatedTx(t, headTime, inputHours, 100, 0, 0, headTime+100_000, 2, 50, 1_000_000)

	if err := txn.VerifyDelegated(headTime, uxIn); err == nil {
		t.Fatalf("expected payout coin limit violation")
	}
}
