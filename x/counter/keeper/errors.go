package keeper

import "cosmossdk.io/errors"

var (
	ErrNumTooLarge       = errors.Register("counter", 0, "requested integer to add is too large")
	ErrExceedsMaxAdd     = errors.Register("counter", 1, "add value exceeds max allowed")
	ErrInsufficientFunds = errors.Register("counter", 2, "insufficient funds to pay add cost")
)
