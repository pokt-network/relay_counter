package main

import (
	"fmt"
)

const (
	HTTPStatusCodeError = "status code non 200"
	InvalidUnitError    = "ERROR: unrecognized unit: "
	Default
)

func NewHTTPStatusCode(code int, body string) error {
	return fmt.Errorf(HTTPStatusCodeError+": %d: With body %s", code, body)
}

func NewInvalidStartEndError(s, e int64, unit string) error {
	return fmt.Errorf("ERROR: unable to interpret timeline: (Start) %d%s Ago to (End) %d%s Ago is not a valid range. Start must come before End", s, unit, e, unit)
}

func NewInvalidUnitError(unit string) error {
	return fmt.Errorf("ERROR: %s%s, valid units: (minutes, hours, days, weeks, blocks, sessions)", InvalidUnitError, unit)
}

func NewInvalidMinimumHeightError(minHeight int64) error {
	return fmt.Errorf("ERROR: the start height is less than 0 (%d), ensure your pocket client is synced and the start and end values are within bounds", minHeight)
}

func NewProofMsgInterfaceError() error {
	return fmt.Errorf("ERROR: unable to convert interface to ProofMsg")
}

func NewPublicKeyError() error {
	return fmt.Errorf("ERROR: unable to convert string public key into ED25519 public key")
}
