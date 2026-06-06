package room

import (
	"testing"
	"time"
)

func TestRoomMembersAndSnapshot(t *testing.T) {
	r := New("room1", "Standup", "", "invite", time.Now().Add(time.Hour))

	m := r.AddMember("peer-a", "Alice")
	if m.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", m.Name)
	}
	if r.Size() != 1 {
		t.Fatalf("expected 1 member, got %d", r.Size())
	}

	r.UpdateMember("peer-a", true, false, true)
	members := r.MemberList()
	if len(members) != 1 || !members[0].Muted || members[0].Speaking != true {
		t.Fatal("expected updated member flags")
	}

	r.SetHost("peer-a")
	snap := r.Snapshot()
	if snap["id"] != "room1" || snap["name"] != "Standup" || snap["hostId"] != "peer-a" {
		t.Fatal("unexpected snapshot identity")
	}

	r.RemoveMember("peer-a")
	if r.Size() != 0 {
		t.Fatal("expected empty room after removal")
	}
}

func TestRoomRenameMember(t *testing.T) {
	r := New("room1", "Team", "", "invite", time.Now().Add(time.Hour))
	r.AddMember("peer-a", "Alice")

	r.RenameMember("peer-a", "Alicia")
	members := r.MemberList()
	if len(members) != 1 || members[0].Name != "Alicia" {
		t.Fatalf("expected renamed member, got %+v", members)
	}

	r.RenameMember("missing", "Ghost")
	if r.Size() != 1 {
		t.Fatal("rename of missing member should not add one")
	}
}

func TestRoomTransfersHost(t *testing.T) {
	r := New("room1", "Team", "", "invite", time.Now().Add(time.Hour))
	r.AddMember("peer-a", "Alice")
	r.AddMember("peer-b", "Bob")
	r.SetHost("peer-a")

	r.RemoveMember("peer-a")
	r.TransferHost("peer-a")
	if r.HostID() != "peer-b" {
		t.Fatalf("expected peer-b to become host, got %q", r.HostID())
	}
}

func TestRoomAddChannelDedupes(t *testing.T) {
	r := New("room1", "Team", "", "invite", time.Now().Add(time.Hour))

	r.AddChannel("dev", "dev")
	r.AddChannel("dev", "dev-duplicate")

	channels := r.Snapshot()["channels"].([]Channel)
	if len(channels) != 2 {
		t.Fatalf("expected general + dev, got %d channels", len(channels))
	}
}
