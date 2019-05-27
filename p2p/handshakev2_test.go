/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bufio"
	"bytes"
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func Test_baseWireHandshaker_writeWireHSRequest(t *testing.T) {
	tests := []struct {
		name     string
		args     p2pcommon.HSHeadReq
		wantErr  bool
		wantSize int
		wantErr2 bool
	}{
		{"TEmpty", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, nil}, false, 8, true},
		{"TSingle", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031}}, false, 12, false},
		{"TMulti", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{0x033333, 0x092fa10, p2pcommon.P2PVersion031, p2pcommon.P2PVersion030}}, false, 24, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &baseWireHandshaker{}
			buffer := bytes.NewBuffer(nil)
			wr := bufio.NewWriter(buffer)
			err := h.writeWireHSRequest(tt.args, wr)
			if (err != nil) != tt.wantErr {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if buffer.Len() != tt.wantSize {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", buffer.Len(), tt.wantSize)
			}

			got, err2 := h.readWireHSRequest(buffer)
			if (err2 != nil) != tt.wantErr2 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", err2, tt.wantErr2)
			}
			if !reflect.DeepEqual(tt.args, got) {
				t.Errorf("baseWireHandshaker.readWireHSRequest() = %v, want %v", got, tt.args)
			}
			if buffer.Len() != 0 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", buffer.Len(), 0)
			}

		})
	}
}

func Test_baseWireHandshaker_writeWireHSResponse(t *testing.T) {
	tests := []struct {
		name     string
		args     p2pcommon.HSHeadResp
		wantErr  bool
		wantSize int
		wantErr2 bool
	}{
		{"TSingle", p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion030.Uint32()}, false, 8, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &baseWireHandshaker{}
			buffer := bytes.NewBuffer(nil)
			wr := bufio.NewWriter(buffer)
			err := h.writeWireHSResponse(tt.args, wr)
			if (err != nil) != tt.wantErr {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if buffer.Len() != tt.wantSize {
				t.Errorf("baseWireHandshaker.writeWireHSRequest() error = %v, wantErr %v", buffer.Len(), tt.wantSize)
			}

			got, err2 := h.readWireHSResp(buffer)
			if (err2 != nil) != tt.wantErr2 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", err2, tt.wantErr2)
			}
			if !reflect.DeepEqual(tt.args, got) {
				t.Errorf("baseWireHandshaker.readWireHSRequest() = %v, want %v", got, tt.args)
			}
			if buffer.Len() != 0 {
				t.Errorf("baseWireHandshaker.readWireHSRequest() error = %v, wantErr %v", buffer.Len(), 0)
			}

		})
	}
}

func TestInboundWireHandshker_handleInboundPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleChainID := &types.ChainID{}
	sampleStatus := &types.Status{}
	logger := log.NewLogger("p2p.test")
	sampleEmptyHSReq := p2pcommon.HSHeadReq{p2pcommon.MAGICMain, nil}
	sampleEmptyHSResp := p2pcommon.HSHeadResp{p2pcommon.HSError, p2pcommon.ErrWrongHSReq}

	type args struct {
		r []byte
	}
	tests := []struct {
		name string
		in   []byte

		bestVer   p2pcommon.P2PVersion
		ctxCancel int  // 0 is not , 1 is during read, 2 is during write
		vhErr     bool // version handshaker failed

		wantW   []byte // sent header
		wantErr bool
	}{
		// All valid
		{"TCurrentVersion", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030, 0x000101}}.Marshal(), p2pcommon.P2PVersion031, 0, false, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion031.Uint32()}.Marshal(), false},
		{"TOldVersion", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{0x000010, p2pcommon.P2PVersion030, 0x000101}}.Marshal(), p2pcommon.P2PVersion030, 0, false, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion030.Uint32()}.Marshal(), false},
		// wrong io read
		{"TWrongRead", sampleEmptyHSReq.Marshal()[:7], p2pcommon.P2PVersion031, 0, false, sampleEmptyHSResp.Marshal(), true},
		// empty version
		{"TEmptyVersion", sampleEmptyHSReq.Marshal(), p2pcommon.P2PVersion031, 0, false, sampleEmptyHSResp.Marshal(), true},
		// wrong io write
		// {"TWrongWrite", sampleEmptyHSReq.Marshal()[:7], sampleEmptyHSResp.Marshal(), true },
		// wrong magic
		{"TWrongMagic", p2pcommon.HSHeadReq{0x0001, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031}}.Marshal(), p2pcommon.P2PVersion031, 0, false, sampleEmptyHSResp.Marshal(), true},
		// not supported version (or wrong version)
		{"TNoVersion", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{0x000010, 0x030405, 0x000101}}.Marshal(), p2pcommon.P2PVersionUnknown, 0, false, p2pcommon.HSHeadResp{p2pcommon.HSError, p2pcommon.ErrNoMatchedVersion}.Marshal(), true},
		// protocol handshake failed
		{"TVersionHSFailed", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030, 0x000101}}.Marshal(), p2pcommon.P2PVersion031, 0, true, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion031.Uint32()}.Marshal(), true},

		// timeout while read, no reply to remote
		{"TTimeoutRead", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030, 0x000101}}.Marshal(), p2pcommon.P2PVersion031, 1, false, []byte{}, true},
		// timeout while writing, sent but remote not receiving fast
		{"TTimeoutWrite", p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030, 0x000101}}.Marshal(), p2pcommon.P2PVersion031, 2, false, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion031.Uint32()}.Marshal(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockVM := p2pmock.NewMockVersionedManager(ctrl)
			mockVH := p2pmock.NewMockVersionedHandshaker(ctrl)

			mockCtx := NewContextTestDouble(tt.ctxCancel) // TODO make mock
			wbuf := bytes.NewBuffer(nil)
			dummyReader := bufio.NewReader(bytes.NewBuffer(tt.in))
			dummyWriter := bufio.NewWriter(wbuf)
			dummyMsgRW := p2pmock.NewMockMsgReadWriter(ctrl)

			mockVM.EXPECT().FindBestP2PVersion(gomock.Any()).Return(tt.bestVer).MaxTimes(1)
			mockVM.EXPECT().GetVersionedHandshaker(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockVH, nil).MaxTimes(1)
			if !tt.vhErr {
				mockVH.EXPECT().DoForInbound(mockCtx).Return(sampleStatus, nil).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(dummyMsgRW).MaxTimes(1)
			} else {
				mockVH.EXPECT().DoForInbound(mockCtx).Return(nil, errors.New("version hs failed")).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(nil).MaxTimes(1)
			}

			h := NewInbountHSHandler(mockPM, mockActor, mockVM, logger, sampleChainID, samplePeerID).(*InboundWireHandshaker)
			got, got1, err := h.handleInboundPeer(mockCtx, dummyReader, dummyWriter)
			if (err != nil) != tt.wantErr {
				t.Errorf("InboundWireHandshaker.handleInboundPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(wbuf.Bytes(), tt.wantW) {
				t.Errorf("InboundWireHandshaker.handleInboundPeer() send resp %v, want %v", wbuf.Bytes(), tt.wantW)
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("InboundWireHandshaker.handleInboundPeer() got msgrw nil, want not")
				}
				if got1 == nil {
					t.Errorf("InboundWireHandshaker.handleInboundPeer() got status nil, want not")
				}
			}
		})
	}
}

func TestOutboundWireHandshaker_handleOutboundPeer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleChainID := &types.ChainID{}
	sampleStatus := &types.Status{}
	logger := log.NewLogger("p2p.test")
	outBytes := p2pcommon.HSHeadReq{p2pcommon.MAGICMain, []p2pcommon.P2PVersion{p2pcommon.P2PVersion031, p2pcommon.P2PVersion030}}.Marshal()

	type args struct {
		r []byte
	}
	tests := []struct {
		name string

		remoteBestVer p2pcommon.P2PVersion
		ctxCancel     int    // 0 is not , 1 is during write, 2 is during read
		vhErr         bool   // version handshaker failed
		receingBuf    []byte // received resp

		wantErr bool
	}{
		// remote listening peer accept my best p2p version
		{"TCurrentVersion", p2pcommon.P2PVersion031, 0, false, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion031.Uint32()}.Marshal(), false},
		// remote listening peer can connect, but old p2p version
		{"TOldVersion", p2pcommon.P2PVersion030, 0, false, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion030.Uint32()}.Marshal(), false},
		// wrong io read
		{"TWrongResp", p2pcommon.P2PVersion031, 0, false, outBytes[:6], true},
		// {"TWrongWrite", sampleEmptyHSReq.Marshal()[:7], sampleEmptyHSResp.Marshal(), true },
		// wrong magic
		{"TWrongMagic", p2pcommon.P2PVersion031, 0, false, p2pcommon.HSHeadResp{p2pcommon.HSError, p2pcommon.ErrWrongHSReq}.Marshal(), true},
		// not supported version (or wrong version)
		{"TNoVersion", p2pcommon.P2PVersionUnknown, 0, false, p2pcommon.HSHeadResp{p2pcommon.HSError, p2pcommon.ErrNoMatchedVersion}.Marshal(), true},
		// protocol handshake failed
		{"TVersionHSFailed", p2pcommon.P2PVersion031, 0, true, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion031.Uint32()}.Marshal(), true},

		// timeout while read, no reply to remote
		{"TTimeoutRead", p2pcommon.P2PVersion031, 1, false, []byte{}, true},
		// timeout while writing, sent but remote not receiving fast
		{"TTimeoutWrite", p2pcommon.P2PVersion031, 2, false, p2pcommon.HSHeadResp{p2pcommon.MAGICMain, p2pcommon.P2PVersion031.Uint32()}.Marshal(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := p2pmock.NewMockPeerManager(ctrl)
			mockActor := p2pmock.NewMockActorService(ctrl)
			mockVM := p2pmock.NewMockVersionedManager(ctrl)
			mockVH := p2pmock.NewMockVersionedHandshaker(ctrl)

			mockCtx := NewContextTestDouble(tt.ctxCancel) // TODO make mock
			wbuf := bytes.NewBuffer(nil)
			dummyReader := bufio.NewReader(bytes.NewBuffer(tt.receingBuf))
			dummyWriter := bufio.NewWriter(wbuf)
			dummyMsgRW := p2pmock.NewMockMsgReadWriter(ctrl)

			mockVM.EXPECT().GetVersionedHandshaker(tt.remoteBestVer, gomock.Any(), gomock.Any(), gomock.Any()).Return(mockVH, nil).MaxTimes(1)
			if !tt.vhErr {
				mockVH.EXPECT().DoForOutbound(mockCtx).Return(sampleStatus, nil).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(dummyMsgRW).MaxTimes(1)
			} else {
				mockVH.EXPECT().DoForOutbound(mockCtx).Return(nil, errors.New("version hs failed")).MaxTimes(1)
				mockVH.EXPECT().GetMsgRW().Return(nil).MaxTimes(1)
			}

			h := NewOutbountHSHandler(mockPM, mockActor, mockVM, logger, sampleChainID, samplePeerID).(*OutboundWireHandshaker)
			got, got1, err := h.handleOutboundPeer(mockCtx, dummyReader, dummyWriter)
			if (err != nil) != tt.wantErr {
				t.Errorf("OutboundWireHandshaker.handleOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !bytes.Equal(wbuf.Bytes(), outBytes) {
				t.Errorf("OutboundWireHandshaker.handleOutboundPeer() send resp %v, want %v", wbuf.Bytes(), tt.receingBuf)
			}
			if !tt.wantErr {
				if got == nil {
					t.Errorf("OutboundWireHandshaker.handleOutboundPeer() got msgrw nil, want not")
				}
				if got1 == nil {
					t.Errorf("OutboundWireHandshaker.handleOutboundPeer() got status nil, want not")
				}
			}
		})
	}
}

type ContextTestDouble struct {
	doneChannel chan struct{}
	expire      uint32
	callCnt     uint32
}

var _ context.Context = (*ContextTestDouble)(nil)

func NewContextTestDouble(expire int) *ContextTestDouble {
	if expire <= 0 {
		expire = 9999999
	}
	return &ContextTestDouble{expire: uint32(expire), doneChannel: make(chan struct{}, 1)}
}

func (*ContextTestDouble) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (c *ContextTestDouble) Done() <-chan struct{} {
	current := atomic.AddUint32(&c.callCnt, 1)
	if current >= c.expire {
		c.doneChannel <- struct{}{}
	}
	return c.doneChannel
}

func (c *ContextTestDouble) Err() error {
	if atomic.LoadUint32(&c.callCnt) >= c.expire {
		return errors.New("timeout")
	} else {
		return nil
	}
}

func (*ContextTestDouble) Value(key interface{}) interface{} {
	panic("implement me")
}