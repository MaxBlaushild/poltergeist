package main

import "testing"

func TestLookupTestAuthUserFormatsPhoneNumber(t *testing.T) {
	user, ok := lookupTestAuthUser("+1 (202) 555-0101")
	if !ok {
		t.Fatalf("expected second test user to be allowlisted")
	}

	if user.PhoneNumber != secondTestPhone {
		t.Fatalf("expected phone %s, got %s", secondTestPhone, user.PhoneNumber)
	}
}

func TestLookupTestAuthUserForLoginChecksCode(t *testing.T) {
	if _, ok := lookupTestAuthUserForLogin(demoPhone, testAuthCode); !ok {
		t.Fatal("expected demo user login to be allowlisted with test auth code")
	}

	if _, ok := lookupTestAuthUserForLogin(demoPhone, "654321"); ok {
		t.Fatal("expected incorrect code to be rejected for allowlisted test user")
	}
}
