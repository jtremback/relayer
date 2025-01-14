package relayer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	chantypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
)

type relayPacket interface {
	Msg(src, dst *Chain) (sdk.Msg, error)
	FetchCommitResponse(src, dst *Chain, queryHeight uint64) error
	Data() []byte
	Seq() uint64
	Timeout() clienttypes.Height
}

type relayMsgTimeout struct {
	packetData   []byte
	seq          uint64
	timeout      clienttypes.Height
	timeoutStamp uint64
	dstRecvRes   *chantypes.QueryPacketReceiptResponse

	pass bool
}

func (rp *relayMsgTimeout) Data() []byte {
	return rp.packetData
}

func (rp *relayMsgTimeout) Seq() uint64 {
	return rp.seq
}

func (rp *relayMsgTimeout) Timeout() clienttypes.Height {
	return rp.timeout
}

func (rp *relayMsgTimeout) FetchCommitResponse(src, dst *Chain, queryHeight uint64) (err error) {
	dstRecvRes, err := dst.QueryPacketReceipt(int64(queryHeight)-1, rp.seq)
	switch {
	case err != nil:
		return err
	case dstRecvRes.Proof == nil:
		return fmt.Errorf("timeout packet receipt proof seq(%d) is nil", rp.seq)
	default:
		rp.dstRecvRes = dstRecvRes
		return nil
	}
}

func (rp *relayMsgTimeout) Msg(src, dst *Chain) (sdk.Msg, error) {
	if rp.dstRecvRes == nil {
		return nil, fmt.Errorf("timeout packet [%s]seq{%d} has no associated proofs", src.ChainID, rp.seq)
	}
	msg := chantypes.NewMsgTimeout(
		chantypes.NewPacket(
			rp.packetData,
			rp.seq,
			src.PathEnd.PortID,
			src.PathEnd.ChannelID,
			dst.PathEnd.PortID,
			dst.PathEnd.ChannelID,
			rp.timeout,
			rp.timeoutStamp,
		),
		rp.seq,
		rp.dstRecvRes.Proof,
		rp.dstRecvRes.ProofHeight,
		src.MustGetAddress(),
	)
	return msg, nil
}

type relayMsgRecvPacket struct {
	packetData   []byte
	seq          uint64
	timeout      clienttypes.Height
	timeoutStamp uint64
	dstComRes    *chantypes.QueryPacketCommitmentResponse

	pass bool
}

func (rp *relayMsgRecvPacket) timeoutPacket() *relayMsgTimeout {
	return &relayMsgTimeout{
		packetData:   rp.packetData,
		seq:          rp.seq,
		timeout:      rp.timeout,
		timeoutStamp: rp.timeoutStamp,
		dstRecvRes:   nil,
		pass:         false,
	}
}

func (rp *relayMsgRecvPacket) Data() []byte {
	return rp.packetData
}

func (rp *relayMsgRecvPacket) Seq() uint64 {
	return rp.seq
}

func (rp *relayMsgRecvPacket) Timeout() clienttypes.Height {
	return rp.timeout
}

func (rp *relayMsgRecvPacket) FetchCommitResponse(src, dst *Chain, queryHeight uint64) (err error) {
	dstCommitRes, err := dst.QueryPacketCommitment(int64(queryHeight)-1, rp.seq)
	switch {
	case err != nil:
		return err
	case dstCommitRes.Proof == nil:
		return fmt.Errorf("recv packet commitment proof seq(%d) is nil", rp.seq)
	case dstCommitRes.Commitment == nil:
		return fmt.Errorf("recv packet commitment query seq(%d) is nil", rp.seq)
	default:
		rp.dstComRes = dstCommitRes
		return nil
	}
}

func (rp *relayMsgRecvPacket) Msg(src, dst *Chain) (sdk.Msg, error) {
	if rp.dstComRes == nil {
		return nil, fmt.Errorf("receive packet [%s]seq{%d} has no associated proofs", src.ChainID, rp.seq)
	}
	packet := chantypes.NewPacket(
		rp.packetData,
		rp.seq,
		dst.PathEnd.PortID,
		dst.PathEnd.ChannelID,
		src.PathEnd.PortID,
		src.PathEnd.ChannelID,
		rp.timeout,
		rp.timeoutStamp,
	)
	msg := chantypes.NewMsgRecvPacket(
		packet,
		rp.dstComRes.Proof,
		rp.dstComRes.ProofHeight,
		src.MustGetAddress(),
	)
	return msg, nil
}

type relayMsgPacketAck struct {
	packetData   []byte
	ack          []byte
	seq          uint64
	timeout      clienttypes.Height
	timeoutStamp uint64
	dstComRes    *chantypes.QueryPacketAcknowledgementResponse

	pass bool
}

func (rp *relayMsgPacketAck) Data() []byte {
	return rp.packetData
}
func (rp *relayMsgPacketAck) Seq() uint64 {
	return rp.seq
}
func (rp *relayMsgPacketAck) Timeout() clienttypes.Height {
	return rp.timeout
}

func (rp *relayMsgPacketAck) Msg(src, dst *Chain) (sdk.Msg, error) {
	if rp.dstComRes == nil {
		return nil, fmt.Errorf("ack packet [%s]seq{%d} has no associated proofs", src.ChainID, rp.seq)
	}
	msg := chantypes.NewMsgAcknowledgement(
		chantypes.NewPacket(
			rp.packetData,
			rp.seq,
			src.PathEnd.PortID,
			src.PathEnd.ChannelID,
			dst.PathEnd.PortID,
			dst.PathEnd.ChannelID,
			rp.timeout,
			rp.timeoutStamp,
		),
		rp.ack,
		rp.dstComRes.Proof,
		rp.dstComRes.ProofHeight,
		src.MustGetAddress(),
	)
	return msg, nil
}

func (rp *relayMsgPacketAck) FetchCommitResponse(src, dst *Chain, queryHeight uint64) (err error) {
	dstCommitRes, err := dst.QueryPacketAcknowledgement(int64(queryHeight)-1, rp.seq)
	switch {
	case err != nil:
		return err
	case dstCommitRes.Proof == nil:
		return fmt.Errorf("ack packet acknowledgement proof seq(%d) is nil", rp.seq)
	case dstCommitRes.Acknowledgement == nil:
		return fmt.Errorf("ack packet acknowledgement query seq(%d) is nil", rp.seq)
	default:
		rp.dstComRes = dstCommitRes
		return nil
	}
}
