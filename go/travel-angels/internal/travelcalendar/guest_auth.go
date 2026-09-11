package travelcalendar

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const guestHeader = "X-Travel-Guest-Token"

func (s *Service) registerGuestRoutes(r *gin.Engine) {
	g := r.Group("/travel-angels")
	g.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Header("X-Robots-Tag", "noindex, nofollow")
		c.Header("Referrer-Policy", "no-referrer")
	})
	g.POST("/calendar-guests/verification", s.requestGuestVerification)
	g.POST("/calendar-guests/verify", s.verifyGuest)
	g.POST("/calendar-guests/claim", s.claimGuest)
	g.GET("/calendar-guests/preferences/:token", s.getGuestPreferences)
	g.PUT("/calendar-guests/preferences/:token", s.updateGuestPreferences)
	g.GET("/calendar-subscriptions", s.listGuestSubscriptions)
	g.POST("/calendar-subscriptions", s.setGuestSubscription)
	g.GET("/calendar-invites", s.listCalendarInvitations)
	g.POST("/calendar-invites", s.inviteCalendarGuest)
	g.POST("/calendar-broadcasts/preview", s.previewCalendarBroadcast)
	g.POST("/calendar-broadcasts", s.createCalendarBroadcast)
	g.GET("/calendar-broadcasts", s.listCalendarBroadcasts)
	g.POST("/calendar-broadcasts/:id/retry", s.retryCalendarBroadcast)
	g.GET("/calendar-notifications", s.listCalendarNotifications)
	g.POST("/calendar-notifications/retry", s.retryCalendarNotifications)
}

func normalizedGuestEmail(input string) (string, error) {
	input = strings.ToLower(strings.TrimSpace(input))
	if len(input) > 254 || strings.ContainsAny(input, "\r\n") {
		return "", errors.New("Enter a valid email address")
	}
	address, err := mail.ParseAddress(input)
	if err != nil || address.Address != input || !strings.Contains(input, "@") {
		return "", errors.New("Enter a valid email address")
	}
	return input, nil
}

func (s *Service) ensureGuestRecipient(tx *gorm.DB, address, name string) (*GuestRecipient, error) {
	address, err := normalizedGuestEmail(address)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if len(name) > 200 {
		return nil, errors.New("Name is too long")
	}
	recipient := GuestRecipient{ID: uuid.New(), Email: address, Name: name, ParticipationNotices: true}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "email"}}, DoNothing: true}).Create(&recipient).Error; err != nil {
		return nil, err
	}
	// On conflict the candidate still holds its unused generated UUID. Clear
	// it so GORM does not add that UUID to the email identity lookup.
	recipient = GuestRecipient{}
	if err := tx.Where("email = ?", address).First(&recipient).Error; err != nil {
		return nil, err
	}
	return &recipient, nil
}

func guestDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

type guestRateLimit struct {
	KeyHash   string `gorm:"primaryKey"`
	Count     int
	ExpiresAt time.Time
}

func (guestRateLimit) TableName() string { return "tc_guest_rate_limits" }

// Fixed windows live in the database so restarting a server or using several
// replicas cannot reset verification and invitation abuse limits.
func (s *Service) allowGuestAction(key string, limit int) bool {
	now := time.Now().UTC()
	window := now.Unix() / 900
	row := guestRateLimit{KeyHash: guestDigest(fmt.Sprintf("%s:%d", key, window)), Count: 1, ExpiresAt: time.Unix((window+1)*900, 0).UTC()}
	err := s.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key_hash"}}, DoUpdates: clause.Assignments(map[string]interface{}{"count": gorm.Expr("tc_guest_rate_limits.count + 1")})}, clause.Returning{}).Create(&row).Error
	if err != nil {
		return false
	}
	return row.Count <= limit
}

func (s *Service) deliveryConfigured(c *gin.Context) bool {
	if s.guestDelivery == nil || s.guestDelivery.Sender == nil {
		c.JSON(503, gin.H{"error": "Email delivery is not configured. Ask the host to configure Travel Angels email delivery."})
		return false
	}
	u, err := url.Parse(s.guestDelivery.BaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		c.JSON(503, gin.H{"error": "The Travel Angels web address is not configured for email links."})
		return false
	}
	return true
}

func (s *Service) requestGuestVerification(c *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "Enter an email address"})
		return
	}
	address, err := normalizedGuestEmail(body.Email)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if !s.deliveryConfigured(c) {
		return
	}
	if !s.allowGuestAction("verify-ip:"+c.ClientIP(), 10) || !s.allowGuestAction("verify-email:"+address, 3) {
		c.JSON(429, gin.H{"error": "Too many verification requests. Try again in 15 minutes."})
		return
	}
	recipient, err := s.ensureGuestRecipient(s.db, address, "")
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not start email verification"})
		return
	}
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not start email verification"})
		return
	}
	code := fmt.Sprintf("%06d", number.Int64())
	challenge := guestChallenge{ID: uuid.New(), RecipientID: recipient.ID, ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	challenge.CodeHash = guestDigest(challenge.ID.String() + ":" + code)
	if err := s.db.Create(&challenge).Error; err != nil {
		c.JSON(500, gin.H{"error": "Could not start email verification"})
		return
	}
	err = s.guestDelivery.Sender.SendMail(email.Email{Email: address, Subject: "Your Travel Angels verification code", PlainTextContent: "Your Travel Angels email verification code is " + code + ".\n\nIt expires in 10 minutes. Enter it in the browser where you requested verification. Verifying your email lets you browse permitted plans and choose whether to follow updates. It does not create an account or join any plans.\n\nIf you did not request this code, ignore this message."})
	if err != nil {
		s.db.Model(&challenge).Update("consumed_at", time.Now().UTC())
		c.JSON(503, gin.H{"error": "The verification email could not be sent. Try again later."})
		return
	}
	c.JSON(202, gin.H{"challengeId": challenge.ID, "expiresAt": challenge.ExpiresAt})
}

func validGuestCode(challenge guestChallenge, code string, now time.Time) bool {
	if challenge.ConsumedAt != nil || !challenge.ExpiresAt.After(now) || challenge.Attempts >= 5 || len(code) != 6 {
		return false
	}
	actual := guestDigest(challenge.ID.String() + ":" + code)
	return subtle.ConstantTimeCompare([]byte(actual), []byte(challenge.CodeHash)) == 1
}

func (s *Service) verifyGuest(c *gin.Context) {
	var body struct {
		ChallengeID uuid.UUID `json:"challengeId"`
		Code        string    `json:"code"`
	}
	if c.ShouldBindJSON(&body) != nil || body.ChallengeID == uuid.Nil {
		c.JSON(400, gin.H{"error": "Enter the verification code"})
		return
	}
	if !s.allowGuestAction("verify-code:"+c.ClientIP(), 50) {
		c.JSON(429, gin.H{"error": "Too many attempts. Try again in 15 minutes."})
		return
	}
	var recipient GuestRecipient
	var token guestToken
	raw := ""
	valid := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var challenge guestChallenge
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&challenge, "id = ?", body.ChallengeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		now := time.Now().UTC()
		if !validGuestCode(challenge, strings.TrimSpace(body.Code), now) {
			if challenge.ConsumedAt == nil && challenge.Attempts < 5 {
				return tx.Model(&challenge).Update("attempts", gorm.Expr("attempts + 1")).Error
			}
			return nil
		}
		if err := tx.Model(&challenge).Updates(map[string]interface{}{"consumed_at": now, "attempts": gorm.Expr("attempts + 1")}).Error; err != nil {
			return err
		}
		if err := tx.First(&recipient, "id = ?", challenge.RecipientID).Error; err != nil {
			return err
		}
		if err := tx.Model(&recipient).Update("verified_at", now).Error; err != nil {
			return err
		}
		raw = newToken()
		token = guestToken{ID: uuid.New(), RecipientID: recipient.ID, TokenHash: guestDigest(raw), Purpose: "view", VerifiedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour)}
		if err := tx.Create(&token).Error; err != nil {
			return err
		}
		valid = true
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not verify email. Try again."})
		return
	}
	if !valid {
		c.JSON(401, gin.H{"error": "That code is invalid, expired, or already used. Request a fresh code."})
		return
	}
	c.JSON(200, gin.H{"guestToken": raw, "expiresAt": token.ExpiresAt, "recipientId": recipient.ID, "email": recipient.Email})
}

func (s *Service) findGuestToken(raw, purpose string) (*guestToken, error) {
	if len(raw) < 32 || len(raw) > 256 {
		return nil, gorm.ErrRecordNotFound
	}
	var token guestToken
	err := s.db.Where("token_hash = ? AND purpose = ? AND expires_at > ?", guestDigest(raw), purpose, time.Now().UTC()).First(&token).Error
	return &token, err
}

func (s *Service) guestRecipient(c *gin.Context) uuid.UUID {
	if token, err := s.findGuestToken(c.GetHeader(guestHeader), "view"); err == nil {
		return token.RecipientID
	}
	if s.authenticate != nil {
		if user, err := s.authenticate(c); err == nil && user != nil {
			var recipient GuestRecipient
			if s.db.Where("user_id = ? AND verified_at IS NOT NULL", user.ID).First(&recipient).Error == nil {
				return recipient.ID
			}
		}
	}
	return uuid.Nil
}

func (s *Service) requireGuestRecipient(c *gin.Context) (*GuestRecipient, bool) {
	id := s.guestRecipient(c)
	var recipient GuestRecipient
	if id == uuid.Nil || s.db.First(&recipient, "id = ?", id).Error != nil {
		c.JSON(401, gin.H{"error": "Verify your email to manage updates. An app account is not required."})
		return nil, false
	}
	return &recipient, true
}

func (s *Service) claimGuest(c *gin.Context) {
	user, ok := s.requireUser(c)
	if !ok {
		return
	}
	token, err := s.findGuestToken(c.GetHeader(guestHeader), "view")
	if err != nil || token.VerifiedAt.Before(time.Now().UTC().Add(-15*time.Minute)) {
		c.JSON(401, gin.H{"error": "Verify your email again before claiming. Your selections have not been submitted."})
		return
	}
	conflict := false
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var recipient GuestRecipient
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&recipient, "id = ?", token.RecipientID).Error; err != nil {
			return err
		}
		if recipient.UserID != nil && *recipient.UserID != user.ID {
			conflict = true
			return nil
		}
		// The shared user directory may already own this email via another app.
		// Phone verification alone can never move that address to this account.
		var existing models.User
		err := tx.Where("LOWER(email) = ? AND id <> ?", recipient.Email, user.ID).First(&existing).Error
		if err == nil {
			conflict = true
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Model(&recipient).Update("user_id", user.ID).Error
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not claim this invitation"})
		return
	}
	if conflict {
		c.JSON(409, gin.H{"error": "This email is already connected to another account. Sign in to or recover that account. Accounts cannot be merged here."})
		return
	}
	c.JSON(200, gin.H{"recipientId": token.RecipientID, "claimed": true})
}
