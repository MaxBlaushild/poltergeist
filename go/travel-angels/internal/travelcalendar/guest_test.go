package travelcalendar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGuestCodeExpiryAttemptsAndConsumption(t *testing.T) {
	now := time.Now().UTC()
	challenge := guestChallenge{ID: uuid.New(), ExpiresAt: now.Add(time.Minute)}
	challenge.CodeHash = guestDigest(challenge.ID.String() + ":001234")
	if !validGuestCode(challenge, "001234", now) {
		t.Fatal("valid code rejected")
	}
	if validGuestCode(challenge, "001235", now) || validGuestCode(challenge, "1234", now) {
		t.Fatal("invalid code accepted")
	}
	if validGuestCode(challenge, "001234", challenge.ExpiresAt) {
		t.Fatal("expired code accepted")
	}
	challenge.Attempts = 5
	if validGuestCode(challenge, "001234", now) {
		t.Fatal("exhausted challenge accepted")
	}
	challenge.Attempts = 0
	challenge.ConsumedAt = &now
	if validGuestCode(challenge, "001234", now) {
		t.Fatal("replayed challenge accepted")
	}
}

func TestGuestEmailNormalization(t *testing.T) {
	address, err := normalizedGuestEmail(" Friend@Example.com ")
	if err != nil || address != "friend@example.com" {
		t.Fatal("email identity did not normalize")
	}
	for _, invalid := range []string{"", "Friend <friend@example.com>", "one@example.com,two@example.com", "user@example.com\r\nBcc: other@example.com"} {
		if _, err := normalizedGuestEmail(invalid); err == nil {
			t.Fatalf("accepted invalid contact %q", invalid)
		}
	}
}

func TestMinimalRevocationNoticeDoesNotExposeTitle(t *testing.T) {
	for _, event := range []string{"removed", "access_removed", "unpublished"} {
		if strings.Contains(noticeMessage(event, "SECRET DESTINATION"), "SECRET") {
			t.Fatalf("%s leaked restricted details", event)
		}
	}
}

func TestBroadcastStatusKeepsUncertainSendsPending(t *testing.T) {
	if broadcastStatus([]calendarDelivery{{Status: "sent"}, {Status: "sending"}}) != "pending" {
		t.Fatal("uncertain delivery must remain pending")
	}
	if broadcastStatus([]calendarDelivery{{Status: "sent"}, {Status: "failed"}}) != "failed" {
		t.Fatal("failed deliveries need explicit retry")
	}
	if !deliveryErrorUncertain(&url.Error{Op: "Post", URL: "https://mail.example.test", Err: context.DeadlineExceeded}) {
		t.Fatal("network timeout must not be blindly retried")
	}
	if deliveryErrorUncertain(errors.New("provider rejected request")) {
		t.Fatal("known provider rejection should allow retry")
	}
}

type guestTestSender struct {
	Messages []email.Email
	Attempts int
	FailNext bool
}

func (sender *guestTestSender) SendMail(message email.Email) error {
	sender.Attempts++
	if sender.FailNext {
		sender.FailNext = false
		return errors.New("simulated provider rejection")
	}
	sender.Messages = append(sender.Messages, message)
	return nil
}

type guestTestFixture struct {
	db      *gorm.DB
	service *Service
	router  *gin.Engine
	sender  *guestTestSender
	owner   uuid.UUID
}

func newGuestTestFixture(t *testing.T) *guestTestFixture {
	t.Helper()
	dsn := os.Getenv("TC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Set TC_TEST_DATABASE_URL to run PostgreSQL guest flow integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	// A unique schema and a single connection isolate every test from the
	// application's tables, even when pointed at an existing test database.
	sqlDB.SetMaxOpenConns(1)
	schema := "tc_guest_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("SET search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP SCHEMA " + schema + " CASCADE"); sqlDB.Close() })
	if err := db.Exec("CREATE TABLE users (id uuid PRIMARY KEY, email text UNIQUE)").Error; err != nil {
		t.Fatal(err)
	}
	_, source, _, _ := runtime.Caller(0)
	for _, migration := range []string{"000473_travel_calendars.up.sql", "000474_travel_calendar_guests.up.sql"} {
		path := filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations", migration)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(data)).Error; err != nil {
			t.Fatalf("migration %s: %v", migration, err)
		}
	}
	owner := uuid.New()
	if err := db.Exec("INSERT INTO users(id) VALUES (?)", owner).Error; err != nil {
		t.Fatal(err)
	}
	service := New(db, func(c *gin.Context) (*models.User, error) {
		id, err := uuid.Parse(c.GetHeader("X-Test-User"))
		if err != nil {
			return nil, errors.New("not authenticated")
		}
		return &models.User{ID: id}, nil
	}, "https://travel.example.test")
	sender := &guestTestSender{}
	service.ConfigureDelivery(GuestDeliveryConfig{Sender: sender, BaseURL: "https://travel.example.test"})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	service.RegisterRoutes(router)
	return &guestTestFixture{db: db, service: service, router: router, sender: sender, owner: owner}
}

func (fixture *guestTestFixture) request(t *testing.T, method, path string, body interface{}, user uuid.UUID, token string) (int, map[string]interface{}) {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, "/travel-angels"+path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	if user != uuid.Nil {
		request.Header.Set("X-Test-User", user.String())
	}
	if token != "" {
		request.Header.Set(guestHeader, token)
	}
	recorder := httptest.NewRecorder()
	fixture.router.ServeHTTP(recorder, request)
	result := map[string]interface{}{}
	if recorder.Body.Len() > 0 {
		if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
			t.Fatalf("non-JSON %d response: %s", recorder.Code, recorder.Body.String())
		}
	}
	return recorder.Code, result
}

func (fixture *guestTestFixture) verify(t *testing.T, address string) (string, uuid.UUID) {
	t.Helper()
	code, result := fixture.request(t, "POST", "/calendar-guests/verification", gin.H{"email": address}, uuid.Nil, "")
	if code != 202 {
		t.Fatalf("request verification: %d %#v", code, result)
	}
	challenge := result["challengeId"].(string)
	message := fixture.sender.Messages[len(fixture.sender.Messages)-1].PlainTextContent
	match := regexp.MustCompile(`code is (\d{6})`).FindStringSubmatch(message)
	if len(match) != 2 {
		t.Fatal("verification message missing code")
	}
	code, result = fixture.request(t, "POST", "/calendar-guests/verify", gin.H{"challengeId": challenge, "code": match[1]}, uuid.Nil, "")
	if code != 200 {
		t.Fatalf("verify: %d %#v", code, result)
	}
	token := result["guestToken"].(string)
	recipient := uuid.MustParse(result["recipientId"].(string))
	code, _ = fixture.request(t, "POST", "/calendar-guests/verify", gin.H{"challengeId": challenge, "code": match[1]}, uuid.Nil, "")
	if code != 401 {
		t.Fatal("verification code replay succeeded")
	}
	return token, recipient
}

func (fixture *guestTestFixture) plans(t *testing.T) (Calendar, Stop) {
	t.Helper()
	cal := Calendar{ID: uuid.New(), OwnerID: fixture.owner, Title: "Year abroad", Visibility: "link", ShareToken: newToken(), SharingEnabled: true}
	content := StopContent{Title: "Lisbon", Destination: "Portugal", StartDate: "2028-05-01", EndDate: "2028-05-10", DatePrecision: "fixed", Status: "confirmed", JoinOpen: true}
	stop := Stop{ID: uuid.New(), CalendarID: cal.ID, Draft: content, Published: &content, Visibility: "inherit", ShareToken: newToken(), SharingEnabled: true}
	if err := fixture.db.Create(&cal).Error; err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.Create(&stop).Error; err != nil {
		t.Fatal(err)
	}
	return cal, stop
}

func TestGuestVerificationClaimAndPurposeIsolation(t *testing.T) {
	fixture := newGuestTestFixture(t)
	token, recipientID := fixture.verify(t, "friend@example.com")
	var recipient GuestRecipient
	if err := fixture.db.First(&recipient, "id = ?", recipientID).Error; err != nil {
		t.Fatal(err)
	}
	if recipient.UserID != nil {
		t.Fatal("guest verification claimed an account")
	}
	var users int64
	fixture.db.Table("users").Count(&users)
	if users != 1 {
		t.Fatal("verification created a user")
	}
	var stored guestToken
	fixture.db.Where("recipient_id = ? AND purpose = ?", recipientID, "view").First(&stored)
	if stored.TokenHash == token || stored.TokenHash != guestDigest(token) {
		t.Fatal("guest token stored unhashed")
	}
	preferences, err := fixture.service.preferenceToken(fixture.db, recipientID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.findGuestToken(preferences, "view"); err == nil {
		t.Fatal("preference capability grants viewing")
	}
	if status, _ := fixture.request(t, "GET", "/calendar-guests/preferences/"+token, nil, uuid.Nil, ""); status != 401 {
		t.Fatal("view token accepted as preferences")
	}
	if status, _ := fixture.request(t, "POST", "/calendar-guests/claim", gin.H{}, uuid.Nil, token); status != 401 {
		t.Fatal("guest can claim without app auth")
	}
	if status, result := fixture.request(t, "POST", "/calendar-guests/claim", gin.H{}, fixture.owner, token); status != 200 {
		t.Fatalf("claim failed %d %#v", status, result)
	}
	fixture.db.First(&recipient, "id = ?", recipientID)
	if recipient.UserID == nil || *recipient.UserID != fixture.owner {
		t.Fatal("claim did not link existing identity")
	}
	other := uuid.New()
	fixture.db.Exec("INSERT INTO users(id) VALUES (?)", other)
	if status, _ := fixture.request(t, "POST", "/calendar-guests/claim", gin.H{}, other, token); status != 409 {
		t.Fatal("verified contact transferred to another account")
	}
	fixture.db.Model(&guestToken{}).Where("token_hash = ?", guestDigest(token)).Update("verified_at", time.Now().Add(-16*time.Minute))
	if status, _ := fixture.request(t, "POST", "/calendar-guests/claim", gin.H{}, fixture.owner, token); status != 401 {
		t.Fatal("stale verification accepted for claiming")
	}
}

func TestClaimRefusesEmailOwnedByExistingAccount(t *testing.T) {
	fixture := newGuestTestFixture(t)
	token, recipientID := fixture.verify(t, "existing@example.com")
	other := uuid.New()
	fixture.db.Exec("INSERT INTO users(id,email) VALUES (?,?)", other, "existing@example.com")
	if status, _ := fixture.request(t, "POST", "/calendar-guests/claim", gin.H{}, fixture.owner, token); status != 409 {
		t.Fatal("claim ignored existing account email ownership")
	}
	var recipient GuestRecipient
	fixture.db.First(&recipient, "id = ?", recipientID)
	if recipient.UserID != nil {
		t.Fatal("conflicting claim mutated recipient")
	}
}

func TestGuestVerificationAttemptAndContactLimits(t *testing.T) {
	fixture := newGuestTestFixture(t)
	status, result := fixture.request(t, "POST", "/calendar-guests/verification", gin.H{"email": "limited@example.com"}, uuid.Nil, "")
	if status != 202 {
		t.Fatalf("verification request failed: %d %#v", status, result)
	}
	challenge := result["challengeId"].(string)
	code := regexp.MustCompile(`code is (\d{6})`).FindStringSubmatch(fixture.sender.Messages[0].PlainTextContent)[1]
	wrong := "000000"
	if code == wrong {
		wrong = "999999"
	}
	for i := 0; i < 5; i++ {
		if status, _ := fixture.request(t, "POST", "/calendar-guests/verify", gin.H{"challengeId": challenge, "code": wrong}, uuid.Nil, ""); status != 401 {
			t.Fatal("incorrect code accepted")
		}
	}
	if status, _ := fixture.request(t, "POST", "/calendar-guests/verify", gin.H{"challengeId": challenge, "code": code}, uuid.Nil, ""); status != 401 {
		t.Fatal("correct code accepted after exhausting attempts")
	}
	for i := 0; i < 2; i++ {
		if status, _ := fixture.request(t, "POST", "/calendar-guests/verification", gin.H{"email": "LIMITED@example.com"}, uuid.Nil, ""); status != 202 {
			t.Fatal("allowed contact retry rejected")
		}
	}
	if status, _ := fixture.request(t, "POST", "/calendar-guests/verification", gin.H{"email": "limited@example.com"}, uuid.Nil, ""); status != 429 {
		t.Fatal("contact verification mail limit bypassed")
	}
	fixture.service.ConfigureDelivery(GuestDeliveryConfig{})
	if status, _ := fixture.request(t, "POST", "/calendar-guests/verification", gin.H{"email": "unconfigured@example.com"}, uuid.Nil, ""); status != 503 {
		t.Fatal("unconfigured mail provider pretended to send verification")
	}
}

func TestInvitationDoesNotSubscribeAndDeduplicates(t *testing.T) {
	fixture := newGuestTestFixture(t)
	cal, _ := fixture.plans(t)
	body := gin.H{"calendarId": cal.ID, "email": "friend@example.com", "name": "Friend"}
	for i := 0; i < 2; i++ {
		status, result := fixture.request(t, "POST", "/calendar-invites", body, fixture.owner, "")
		if status != 200 || result["status"] != "sent" {
			t.Fatalf("invite %d: %d %#v", i, status, result)
		}
	}
	if len(fixture.sender.Messages) != 1 {
		t.Fatal("repeated invitation sent duplicate successful email")
	}
	var count int64
	fixture.db.Model(&calendarSubscription{}).Count(&count)
	if count != 0 {
		t.Fatal("invitation silently subscribed guest")
	}
	fixture.db.Model(&GuestRecipient{}).Count(&count)
	if count != 1 {
		t.Fatal("invitation created duplicate recipient")
	}
	if strings.Contains(fixture.sender.Messages[0].PlainTextContent, "Lisbon") || strings.Contains(fixture.sender.Messages[0].PlainTextContent, "Portugal") {
		t.Fatal("generic invitation leaked itinerary")
	}
}

func TestBroadcastRechecksPermissionsRetriesAndPreferences(t *testing.T) {
	fixture := newGuestTestFixture(t)
	cal, stop := fixture.plans(t)
	token1, recipient1 := fixture.verify(t, "one@example.com")
	token2, recipient2 := fixture.verify(t, "two@example.com")
	for _, token := range []string{token1, token2} {
		status, result := fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"calendarId": cal.ID, "shareToken": cal.ShareToken, "subscribed": true}, uuid.Nil, token)
		if status != 200 {
			t.Fatalf("subscribe %d %#v", status, result)
		}
	}
	fixture.sender.Messages = nil
	fixture.sender.Attempts = 0
	fixture.sender.FailNext = true
	body := gin.H{"calendarId": cal.ID, "message": "Come visit our plans"}
	status, result := fixture.request(t, "POST", "/calendar-broadcasts", body, fixture.owner, "")
	if status != 200 {
		t.Fatalf("queue %d %#v", status, result)
	}
	broadcastID := result["id"].(string)
	if err := fixture.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if fixture.sender.Attempts != 2 || len(fixture.sender.Messages) != 1 {
		t.Fatal("expected one failed and one successful delivery")
	}
	status, result = fixture.request(t, "POST", "/calendar-broadcasts/"+broadcastID+"/retry", gin.H{}, fixture.owner, "")
	if status != 200 {
		t.Fatalf("retry %d %#v", status, result)
	}
	if err := fixture.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if err := fixture.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	if fixture.sender.Attempts != 3 || len(fixture.sender.Messages) != 2 {
		t.Fatal("retry duplicated an already successful send")
	}
	status, result = fixture.request(t, "POST", "/calendar-broadcasts", body, fixture.owner, "")
	if status != 200 {
		t.Fatalf("queue second %d %#v", status, result)
	}
	queuedID := uuid.MustParse(result["id"].(string))
	// Revoke while a message waits in the outbox. A generic notice is allowed;
	// the queued broadcast must be skipped without leaking its free-form text.
	stop.Visibility = "specific"
	if err := fixture.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&stop).Update("visibility", "specific").Error; err != nil {
			return err
		}
		return fixture.service.reconcileGuestAccess(tx, &stop)
	}); err != nil {
		t.Fatal(err)
	}
	if err := fixture.service.DispatchPendingDeliveries(context.Background(), 25); err != nil {
		t.Fatal(err)
	}
	var deliveries []calendarDelivery
	fixture.db.Where("broadcast_id = ?", queuedID).Find(&deliveries)
	for _, delivery := range deliveries {
		if delivery.Status != "skipped" {
			t.Fatalf("queued unauthorized message %s", delivery.Status)
		}
	}
	for _, recipientID := range []uuid.UUID{recipient1, recipient2} {
		var count int64
		fixture.db.Model(&subscriptionExclusion{}).Where("recipient_id = ? AND stop_id = ?", recipientID, stop.ID).Count(&count)
		if count != 1 {
			t.Fatal("revocation did not persist stop subscription exclusion")
		}
	}
	fixture.db.Model(&stop).Update("visibility", "inherit")
	stop.Visibility = "inherit"
	var recipient GuestRecipient
	fixture.db.First(&recipient, "id = ?", recipient1)
	if fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("restoring permission silently resumed stop updates")
	}
	status, result = fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"stopId": stop.ID, "shareToken": stop.ShareToken, "subscribed": true}, uuid.Nil, token1)
	if status != 200 {
		t.Fatalf("explicit resubscribe %d %#v", status, result)
	}
	if !fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("explicit follow did not resume allowed updates")
	}
	preferences, err := fixture.service.preferenceToken(fixture.db, recipient1)
	if err != nil {
		t.Fatal(err)
	}
	status, result = fixture.request(t, "GET", "/calendar-guests/preferences/"+preferences, nil, uuid.Nil, "")
	if status != 200 {
		t.Fatalf("preferences %d %#v", status, result)
	}
	serialized, _ := json.Marshal(result)
	if strings.Contains(string(serialized), "Lisbon") {
		t.Fatal("preferences token exposed stop details")
	}
	status, result = fixture.request(t, "PUT", "/calendar-guests/preferences/"+preferences, gin.H{"unsubscribeAll": true}, uuid.Nil, "")
	if status != 200 {
		t.Fatalf("unsubscribe %d %#v", status, result)
	}
	if fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("unsubscribe did not affect dispatch eligibility")
	}
}

func TestBroadcastSharedTextRequiresEveryStopToBeAuthorized(t *testing.T) {
	fixture := newGuestTestFixture(t)
	cal, stop := fixture.plans(t)
	token, _ := fixture.verify(t, "friend@example.com")
	fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"calendarId": cal.ID, "shareToken": cal.ShareToken, "subscribed": true}, uuid.Nil, token)
	private := stop
	private.ID = uuid.New()
	private.ShareToken = newToken()
	private.Visibility = "private"
	if err := fixture.db.Create(&private).Error; err != nil {
		t.Fatal(err)
	}
	status, result := fixture.request(t, "POST", "/calendar-broadcasts/preview", gin.H{"calendarId": cal.ID, "message": "Secret details about both stops"}, fixture.owner, "")
	if status != 200 || result["recipientCount"] != float64(0) {
		t.Fatalf("mixed audience broadcast leaks free-form text %d %#v", status, result)
	}
}

func TestSharedStopCalendarFollowersDeduplicateWithoutExpandingAccess(t *testing.T) {
	fixture := newGuestTestFixture(t)
	calendar, stop := fixture.plans(t)
	cohost := uuid.New()
	if err := fixture.db.Exec("INSERT INTO users(id) VALUES (?)", cohost).Error; err != nil {
		t.Fatal(err)
	}
	second := Calendar{ID: uuid.New(), OwnerID: cohost, Title: "Co-host calendar", Visibility: "link", ShareToken: newToken(), SharingEnabled: true}
	if err := fixture.db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.Create(&Collaborator{ID: uuid.New(), StopID: &stop.ID, UserID: cohost, Accepted: true}).Error; err != nil {
		t.Fatal(err)
	}
	token, recipientID := fixture.verify(t, "shared@example.com")
	if status, result := fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"calendarId": second.ID, "shareToken": second.ShareToken, "subscribed": true}, uuid.Nil, token); status != 200 {
		t.Fatalf("follow co-host: %d %#v", status, result)
	}
	body := gin.H{"stopIds": []uuid.UUID{stop.ID}, "message": "Our shared stop"}
	status, result := fixture.request(t, "POST", "/calendar-broadcasts/preview", body, fixture.owner, "")
	if status != 200 || result["recipientCount"] != float64(0) {
		t.Fatalf("following second calendar expanded original link permissions: %d %#v", status, result)
	}
	stop.Visibility = "specific"
	if err := fixture.db.Model(&stop).Update("visibility", "specific").Error; err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.Create(&Grant{ID: uuid.New(), StopID: &stop.ID, RecipientID: &recipientID}).Error; err != nil {
		t.Fatal(err)
	}
	status, result = fixture.request(t, "POST", "/calendar-broadcasts/preview", body, cohost, "")
	if status != 200 || result["recipientCount"] != float64(1) {
		t.Fatalf("authorized co-host follower omitted: %d %#v", status, result)
	}
	fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"calendarId": calendar.ID, "shareToken": calendar.ShareToken, "subscribed": true}, uuid.Nil, token)
	status, result = fixture.request(t, "POST", "/calendar-broadcasts/preview", body, fixture.owner, "")
	if status != 200 || result["recipientCount"] != float64(1) {
		t.Fatalf("both host follows duplicate recipient: %d %#v", status, result)
	}
}

func TestGuestSubscriptionOnlyRetainsObservedShareCapabilities(t *testing.T) {
	fixture := newGuestTestFixture(t)
	calendar, stop := fixture.plans(t)
	token, recipientID := fixture.verify(t, "capabilities@example.com")
	status, result := fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"calendarId": calendar.ID, "shareToken": calendar.ShareToken, "subscribed": true}, uuid.Nil, token)
	if status != 200 {
		t.Fatalf("follow calendar: %d %#v", status, result)
	}
	var subscription calendarSubscription
	fixture.db.Where("calendar_id = ? AND recipient_id = ?", calendar.ID, recipientID).First(&subscription)
	if subscription.CalendarLinkTokenHash != guestDigest(calendar.ShareToken) || subscription.StopLinkTokenHash != "" {
		t.Fatal("subscription did not retain only the actual calendar capability as a hash")
	}
	var recipient GuestRecipient
	fixture.db.First(&recipient, "id = ?", recipientID)
	calendar.ShareToken = newToken()
	if err := fixture.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&calendar).Update("share_token", calendar.ShareToken).Error; err != nil {
			return err
		}
		return fixture.service.reconcileGuestAccess(tx, &stop)
	}); err != nil {
		t.Fatal(err)
	}
	if fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("rotated calendar token retained access through an unobserved stop link")
	}
	status, result = fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"stopId": stop.ID, "shareToken": stop.ShareToken, "subscribed": true}, uuid.Nil, token)
	if status != 200 {
		t.Fatalf("follow current stop link: %d %#v", status, result)
	}
	if !fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("newly observed stop link was not usable")
	}
	stop.SharingEnabled = false
	if err := fixture.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&stop).Update("sharing_enabled", false).Error; err != nil {
			return err
		}
		return fixture.service.reconcileGuestAccess(tx, &stop)
	}); err != nil {
		t.Fatal(err)
	}
	if fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("disabled stop link retained access through an unobserved new calendar link")
	}
	stop.Visibility = "specific"
	fixture.db.Model(&stop).Update("visibility", "specific")
	fixture.db.Create(&Grant{ID: uuid.New(), StopID: &stop.ID, RecipientID: &recipientID})
	if fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("new grant silently restored a paused subscription")
	}
	status, result = fixture.request(t, "POST", "/calendar-subscriptions", gin.H{"stopId": stop.ID, "subscribed": true}, uuid.Nil, token)
	if status != 200 {
		t.Fatalf("follow by specific grant: %d %#v", status, result)
	}
	if !fixture.service.broadcastRecipientEligible(fixture.db, &recipient, []Stop{stop}) {
		t.Fatal("specific grant did not authorize explicit resubscription")
	}
}

func TestParticipationNoticeDoesNotInventLinkCapabilities(t *testing.T) {
	fixture := newGuestTestFixture(t)
	calendar, stop := fixture.plans(t)
	_, recipientID := fixture.verify(t, "notices@example.com")
	participantID := uuid.New()
	fixture.db.Exec("INSERT INTO users(id) VALUES (?)", participantID)
	fixture.db.Model(&GuestRecipient{}).Where("id = ?", recipientID).Update("user_id", participantID)
	participation := Participation{ID: uuid.New(), StopID: stop.ID, UserID: participantID, Status: "requested", StartDate: "2028-05-01", EndDate: "2028-05-10", CalendarLinkTokenHash: guestDigest(calendar.ShareToken)}
	if err := fixture.db.Create(&participation).Error; err != nil {
		t.Fatal(err)
	}
	if err := fixture.service.queueParticipationNotice(fixture.db, &stop, &participation, "requested"); err != nil {
		t.Fatal(err)
	}
	fixture.db.Model(&calendar).Update("share_token", newToken())
	var recipient GuestRecipient
	fixture.db.First(&recipient, "id = ?", recipientID)
	delivery := calendarDelivery{RecipientID: recipientID, StopID: &stop.ID, ParticipationID: &participation.ID, Event: "requested"}
	_, content, allowed := fixture.service.participationEmail(fixture.db, &delivery, &recipient, &stop)
	if !allowed || strings.Contains(content, "Lisbon") || strings.Contains(content, "/#/stops/") {
		t.Fatal("notice invented an unobserved stop link after calendar token rotation")
	}
	status, result := fixture.request(t, "GET", "/calendar-notifications", nil, participantID, "")
	serialized, _ := json.Marshal(result)
	if status != 200 || strings.Contains(string(serialized), "Lisbon") || strings.Contains(string(serialized), "stopId") {
		t.Fatalf("in-app notice exposed revoked details: %d %s", status, serialized)
	}
}
