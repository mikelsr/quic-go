package handshake

import (
	"crypto/tls"
	"errors"
	"io"
	"time"

	"github.com/mikelsr/quic-go/internal/protocol"
	"github.com/mikelsr/quic-go/internal/wire"
)

var (
	// ErrKeysNotYetAvailable is returned when an opener or a sealer is requested for an encryption level,
	// but the corresponding opener has not yet been initialized
	// This can happen when packets arrive out of order.
	ErrKeysNotYetAvailable = errors.New("CryptoSetup: keys at this encryption level not yet available")
	// ErrKeysDropped is returned when an opener or a sealer is requested for an encryption level,
	// but the corresponding keys have already been dropped.
	ErrKeysDropped = errors.New("CryptoSetup: keys were already dropped")
	// ErrDecryptionFailed is returned when the AEAD fails to open the packet.
	ErrDecryptionFailed = errors.New("decryption failed")
)

type headerDecryptor interface {
	DecryptHeader(sample []byte, firstByte *byte, pnBytes []byte)
}

// LongHeaderOpener opens a long header packet
type LongHeaderOpener interface {
	headerDecryptor
	DecodePacketNumber(wirePN protocol.PacketNumber, wirePNLen protocol.PacketNumberLen) protocol.PacketNumber
	Open(dst, src []byte, pn protocol.PacketNumber, associatedData []byte) ([]byte, error)
}

// ShortHeaderOpener opens a short header packet
type ShortHeaderOpener interface {
	headerDecryptor
	DecodePacketNumber(wirePN protocol.PacketNumber, wirePNLen protocol.PacketNumberLen) protocol.PacketNumber
	Open(dst, src []byte, rcvTime time.Time, pn protocol.PacketNumber, kp protocol.KeyPhaseBit, associatedData []byte) ([]byte, error)
}

// LongHeaderSealer seals a long header packet
type LongHeaderSealer interface {
	Seal(dst, src []byte, packetNumber protocol.PacketNumber, associatedData []byte) []byte
	EncryptHeader(sample []byte, firstByte *byte, pnBytes []byte)
	Overhead() int
}

// ShortHeaderSealer seals a short header packet
type ShortHeaderSealer interface {
	LongHeaderSealer
	KeyPhase() protocol.KeyPhaseBit
}

type handshakeRunner interface {
	OnReceivedParams(*wire.TransportParameters)
	OnHandshakeComplete()
	OnReceivedReadKeys()
	DropKeys(protocol.EncryptionLevel)
}

type ConnectionState struct {
	tls.ConnectionState
	Used0RTT bool
}

// CryptoSetup handles the handshake and protecting / unprotecting packets
type CryptoSetup interface {
	StartHandshake() error
	io.Closer
	ChangeConnectionID(protocol.ConnectionID)
	GetSessionTicket() ([]byte, error)

	HandleMessage([]byte, protocol.EncryptionLevel) error
	SetLargest1RTTAcked(protocol.PacketNumber) error
	SetHandshakeConfirmed()
	ConnectionState() ConnectionState

	GetInitialOpener() (LongHeaderOpener, error)
	GetHandshakeOpener() (LongHeaderOpener, error)
	Get0RTTOpener() (LongHeaderOpener, error)
	Get1RTTOpener() (ShortHeaderOpener, error)

	GetInitialSealer() (LongHeaderSealer, error)
	GetHandshakeSealer() (LongHeaderSealer, error)
	Get0RTTSealer() (LongHeaderSealer, error)
	Get1RTTSealer() (ShortHeaderSealer, error)
}
