package models

import "testing"

func TestIsSystemView(t *testing.T) {
	// The system view (authentication disabled) is nil-user → admin rights.
	var anonymous *User
	if !anonymous.IsSystemView() {
		t.Error("nil user must be a system view")
	}
	if !(&User{Username: "root", IsAdmin: true}).IsSystemView() {
		t.Error("admin must be a system view")
	}
	if (&User{Username: "alice"}).IsSystemView() {
		t.Error("regular user must not be a system view")
	}
}
