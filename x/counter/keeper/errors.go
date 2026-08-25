package keeper

import "cosmossdk.io/errors"

var (
	// Codes start at 2: code 0 is reserved for success and code 1 for internal errors.
	ErrNumTooLarge       = errors.Register("counter", 2, "requested integer to add is too large")
	ErrExceedsMaxAdd     = errors.Register("counter", 3, "add value exceeds max allowed")
	ErrInsufficientFunds = errors.Register("counter", 4, "insufficient funds to pay add cost")
)
