package protocol

import "errors"

var (
	ErrInvalidPacket  = errors.New("invalid packet format")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrVersionMismatch = errors.New("protocol version mismatch")
	ErrPacketTooLarge = errors.New("packet exceeds maximum size")
)
