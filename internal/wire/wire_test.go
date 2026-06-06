package wire

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	body := []byte("payload")
	frame, err := EncodeFrame(MsgPing, body)
	if err != nil {
		t.Fatal(err)
	}
	typ, decoded, err := DecodeFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	if typ != MsgPing || !bytes.Equal(decoded, body) {
		t.Fatalf("unexpected frame: %d %q", typ, decoded)
	}
}

func TestJoinRoundTrip(t *testing.T) {
	in := Join{
		RoomID:       "room1",
		Invite:       "invite",
		Password:     "secret",
		Name:         "Ada",
		ResumePeerID: "peer1",
		ResumeToken:  "token",
		Pow:          &Pow{ID: "pow1", Nonce: 42},
	}
	raw, err := EncodeJoin(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeJoin(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.RoomID != in.RoomID || out.Invite != in.Invite || out.Name != in.Name || out.Pow.Nonce != in.Pow.Nonce {
		t.Fatalf("join mismatch: %+v vs %+v", out, in)
	}
}

func TestSignalRoundTrip(t *testing.T) {
	in := Signal{
		To:    "peer-b",
		From:  "peer-a",
		Nonce: "nonce",
		Sig:   "sig",
		Kind:  SignalCandidate,
		Body:  json.RawMessage(`{"candidate":"x"}`),
	}
	raw, err := EncodeSignal(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeSignal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.To != in.To || out.From != in.From || out.Kind != in.Kind || !bytes.Equal(out.Body, in.Body) {
		t.Fatalf("signal mismatch: %+v vs %+v", out, in)
	}
}

func TestPingFixedSize(t *testing.T) {
	raw := EncodePing(1234567890)
	if len(raw) != 8 {
		t.Fatalf("expected 8 bytes, got %d", len(raw))
	}
	got, err := DecodePing(raw)
	if err != nil || got != 1234567890 {
		t.Fatalf("unexpected ping value: %d err=%v", got, err)
	}
}

func TestValidCandidateBody(t *testing.T) {
	if !ValidCandidateBody([]byte(`{"candidate":"1"}`)) {
		t.Fatal("expected valid candidate")
	}
	if ValidCandidateBody([]byte(`not-json`)) {
		t.Fatal("expected invalid candidate")
	}
}
