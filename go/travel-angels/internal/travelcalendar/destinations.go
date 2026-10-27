package travelcalendar

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DestinationSnapshot is saved independently in draft and published content.
// An absent ID denotes a trip-local label, never a shared catalog entry.
type DestinationSnapshot struct {
	ID               *uuid.UUID     `json:"id,omitempty"`
	Name             string         `json:"name"`
	Kind             string         `json:"kind"`
	CountryCode      string         `json:"countryCode,omitempty"`
	CountryName      string         `json:"countryName,omitempty"`
	ParentID         *uuid.UUID     `json:"parentId,omitempty"`
	FormattedAddress string         `json:"formattedAddress,omitempty"`
	Latitude         *float64       `json:"latitude,omitempty"`
	Longitude        *float64       `json:"longitude,omitempty"`
	TimeZone         string         `json:"timeZone,omitempty"`
	Aliases          []string       `json:"aliases,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	Theme            map[string]any `json:"theme,omitempty"`
}

type Destination struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name             string
	Kind             string
	CountryCode      string
	CountryName      string
	ParentID         *uuid.UUID `gorm:"type:uuid"`
	FormattedAddress string
	Latitude         *float64
	Longitude        *float64
	TimeZone         string
	Aliases          []string       `gorm:"serializer:json;type:jsonb"`
	Metadata         map[string]any `gorm:"serializer:json;type:jsonb"`
	Theme            map[string]any `gorm:"serializer:json;type:jsonb"`
	Provider         string
	ProviderID       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (Destination) TableName() string { return "tc_destinations" }

func (d Destination) snapshot() DestinationSnapshot {
	id := d.ID
	return DestinationSnapshot{ID: &id, Name: d.Name, Kind: d.Kind, CountryCode: d.CountryCode, CountryName: d.CountryName, ParentID: d.ParentID, FormattedAddress: d.FormattedAddress, Latitude: d.Latitude, Longitude: d.Longitude, TimeZone: d.TimeZone, Aliases: d.Aliases, Metadata: d.Metadata, Theme: d.Theme}
}

type DestinationCandidate struct {
	DestinationSnapshot
	Provider   string
	ProviderID string
}

// Providers return public place data only; they never receive the viewer or
// their private saved destinations. Errors leave the offline catalog usable.
type DestinationSearchProvider func(context.Context, string) ([]DestinationCandidate, error)

func (s *Service) ConfigureDestinationSearch(provider DestinationSearchProvider) {
	s.destinationSearch = provider
}

func DestinationCountryID(code string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("travel-angels/destination/country/"+strings.ToUpper(strings.TrimSpace(code))))
}

type destinationUsage struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	StopID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	DestinationID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	LastSelectedAt time.Time
}

func (destinationUsage) TableName() string { return "tc_destination_usage" }

func normalizeDestination(d *DestinationSnapshot) error {
	d.Name = strings.TrimSpace(d.Name)
	d.Kind = strings.ToLower(strings.TrimSpace(d.Kind))
	if d.Kind == "" {
		d.Kind = "place"
	}
	if d.Name == "" || utf8.RuneCountInString(d.Name) > 300 {
		return bad("Each destination needs a name of at most 300 characters")
	}
	if d.Kind != "country" && d.Kind != "region" && d.Kind != "city" && d.Kind != "place" {
		return bad("Choose a country, region, city, or place destination")
	}
	d.CountryCode = strings.ToUpper(strings.TrimSpace(d.CountryCode))
	if d.CountryCode != "" && (len(d.CountryCode) != 2 || d.CountryCode[0] < 'A' || d.CountryCode[0] > 'Z' || d.CountryCode[1] < 'A' || d.CountryCode[1] > 'Z') {
		return bad("Invalid destination country code")
	}
	if (d.Latitude == nil) != (d.Longitude == nil) {
		return bad("Destination coordinates require both latitude and longitude")
	}
	if d.Latitude != nil && (math.IsNaN(*d.Latitude) || math.IsNaN(*d.Longitude) || math.Abs(*d.Latitude) > 90 || math.Abs(*d.Longitude) > 180) {
		return bad("Invalid destination coordinates")
	}
	d.TimeZone = strings.TrimSpace(d.TimeZone)
	if d.TimeZone != "" {
		if _, err := time.LoadLocation(d.TimeZone); err != nil || d.TimeZone == "Local" {
			return bad("Invalid destination time zone")
		}
	}
	if utf8.RuneCountInString(d.FormattedAddress) > 500 || utf8.RuneCountInString(d.CountryName) > 200 || len(d.Aliases) > 30 {
		return bad("Destination details exceed limits")
	}
	for _, alias := range d.Aliases {
		if utf8.RuneCountInString(alias) > 200 {
			return bad("Destination alias exceeds limits")
		}
	}
	data, err := json.Marshal([]any{d.Metadata, d.Theme})
	if err != nil || len(data) > 8192 {
		return bad("Destination metadata exceeds limits")
	}
	if d.Aliases == nil {
		d.Aliases = []string{}
	}
	if d.Metadata == nil {
		d.Metadata = map[string]any{}
	}
	if d.Theme == nil {
		d.Theme = map[string]any{}
	}
	return nil
}

func contentDestinations(content *StopContent) []DestinationSnapshot {
	if len(content.Destinations) > 0 {
		return content.Destinations
	}
	if name := strings.TrimSpace(content.Destination); name != "" {
		return []DestinationSnapshot{{Name: name, Kind: "place"}}
	}
	return []DestinationSnapshot{}
}

func destinationSummary(destinations []DestinationSnapshot) string {
	names := make([]string, 0, len(destinations))
	for _, d := range destinations {
		names = append(names, d.Name)
	}
	return strings.Join(names, ", ")
}

func matchesDestinationQuery(content *StopContent, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if strings.Contains(strings.ToLower(content.Destination), query) {
		return true
	}
	for _, destination := range contentDestinations(content) {
		fields := append([]string{destination.Name, destination.CountryName, destination.CountryCode, destination.FormattedAddress}, destination.Aliases...)
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field), query) {
				return true
			}
		}
	}
	return false
}

func destinationIdentities(content *StopContent) string {
	identities := []string{}
	for _, d := range contentDestinations(content) {
		if d.ID != nil {
			identities = append(identities, "id:"+d.ID.String())
		} else {
			identities = append(identities, "custom:"+d.Kind+":"+strings.TrimSpace(d.Name)+":"+d.CountryCode)
		}
	}
	encoded, _ := json.Marshal(identities)
	return string(encoded)
}

// Missing arrays are old clients. Preserve structured choices on unrelated
// edits; an explicit changed legacy label becomes one custom destination.
func (s *Service) prepareDestinations(tx *gorm.DB, content, prior *StopContent) error {
	if content.Destinations == nil {
		if prior != nil && (strings.TrimSpace(content.Destination) == "" || strings.TrimSpace(content.Destination) == prior.Destination) {
			content.Destinations = append([]DestinationSnapshot{}, contentDestinations(prior)...)
		} else {
			content.Destinations = contentDestinations(content)
		}
	}
	if len(content.Destinations) == 0 || len(content.Destinations) > 12 {
		return bad("Choose between 1 and 12 destinations")
	}
	content.Destinations = append([]DestinationSnapshot{}, content.Destinations...)
	seen := map[string]bool{}
	for i := range content.Destinations {
		d := &content.Destinations[i]
		if d.ID != nil {
			var canonical Destination
			if err := tx.First(&canonical, "id = ?", *d.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return bad("A selected destination no longer exists; choose it again")
				}
				return err
			}
			*d = canonical.snapshot()
		} else {
			// Custom labels cannot invent catalog ancestry or provider metadata.
			d.ParentID, d.Metadata, d.Theme, d.Aliases = nil, nil, nil, nil
		}
		if err := normalizeDestination(d); err != nil {
			return err
		}
		identity := strings.ToLower(d.Kind + ":" + d.Name + ":" + d.CountryCode)
		if d.ID != nil {
			identity = d.ID.String()
		}
		if seen[identity] {
			return bad("Choose each destination only once")
		}
		seen[identity] = true
	}
	content.Destination = destinationSummary(content.Destinations)
	return nil
}

func (s *Service) recordDestinationUsage(tx *gorm.DB, stop *Stop, userID uuid.UUID) error {
	ids := []uuid.UUID{}
	for _, d := range stop.Draft.Destinations {
		if d.ID != nil {
			ids = append(ids, *d.ID)
		}
	}
	query := tx.Where("user_id = ? AND stop_id = ?", userID, stop.ID)
	if len(ids) > 0 {
		query = query.Where("destination_id NOT IN ?", ids)
	}
	if err := query.Delete(&destinationUsage{}).Error; err != nil {
		return err
	}
	for _, id := range ids {
		usage := destinationUsage{UserID: userID, StopID: stop.ID, DestinationID: id, LastSelectedAt: time.Now().UTC()}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&usage).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) persistDestinationCandidate(tx *gorm.DB, candidate DestinationCandidate) (*Destination, error) {
	d := candidate.DestinationSnapshot
	if err := normalizeDestination(&d); err != nil {
		return nil, err
	}
	provider, providerID := strings.TrimSpace(candidate.Provider), strings.TrimSpace(candidate.ProviderID)
	if provider == "" || len(provider) > 80 || providerID == "" || len(providerID) > 300 {
		return nil, bad("Destination provider identity is missing")
	}
	if d.CountryCode != "" {
		var country Destination
		if err := tx.First(&country, "id = ?", DestinationCountryID(d.CountryCode)).Error; err != nil {
			return nil, err
		}
		if d.Kind == "country" {
			return &country, nil
		}
		d.CountryName = country.Name
		if d.ParentID == nil {
			parent := country.ID
			d.ParentID = &parent
		}
	}
	if d.ParentID != nil {
		var parent Destination
		if err := tx.First(&parent, "id = ?", *d.ParentID).Error; err != nil {
			return nil, err
		}
	}
	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte("travel-angels/destination/provider/"+provider+"/"+providerID))
	row := Destination{ID: id, Name: d.Name, Kind: d.Kind, CountryCode: d.CountryCode, CountryName: d.CountryName, ParentID: d.ParentID, FormattedAddress: d.FormattedAddress, Latitude: d.Latitude, Longitude: d.Longitude, TimeZone: d.TimeZone, Aliases: d.Aliases, Metadata: d.Metadata, Theme: d.Theme, Provider: provider, ProviderID: providerID}
	// Provider refreshes never rewrite already published snapshots.
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&row).Error; err != nil {
		return nil, err
	}
	var existing Destination
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&existing, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// A later provider response can enrich initially missing details. Existing
	// canonical labels and curated metadata/theme always take precedence.
	if existing.CountryCode == "" {
		existing.CountryCode = row.CountryCode
	}
	if existing.CountryName == "" {
		existing.CountryName = row.CountryName
	}
	if existing.ParentID == nil {
		existing.ParentID = row.ParentID
	}
	if existing.FormattedAddress == "" {
		existing.FormattedAddress = row.FormattedAddress
	}
	if existing.Latitude == nil {
		existing.Latitude, existing.Longitude = row.Latitude, row.Longitude
	}
	if existing.TimeZone == "" {
		existing.TimeZone = row.TimeZone
	}
	if len(existing.Metadata) == 0 {
		existing.Metadata = row.Metadata
	}
	if len(existing.Theme) == 0 {
		existing.Theme = row.Theme
	}
	seenAliases := map[string]bool{}
	for _, alias := range existing.Aliases {
		seenAliases[alias] = true
	}
	for _, alias := range row.Aliases {
		if !seenAliases[alias] && len(existing.Aliases) < 30 {
			existing.Aliases = append(existing.Aliases, alias)
			seenAliases[alias] = true
		}
	}
	if existing.Aliases == nil {
		existing.Aliases = []string{}
	}
	if err := tx.Save(&existing).Error; err != nil {
		return nil, err
	}
	row = existing
	terms := append([]string{row.Name}, row.Aliases...)
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" {
			continue
		}
		if err := tx.Exec("INSERT INTO tc_destination_terms(destination_id, term) VALUES (?, ?) ON CONFLICT DO NOTHING", row.ID, term).Error; err != nil {
			return nil, err
		}
	}
	return &row, nil
}

type destinationRank struct {
	DestinationID  uuid.UUID
	Frequency      int
	LastSelectedAt time.Time
}

type destinationResult struct {
	DestinationSnapshot
	Recent    bool `json:"recent"`
	Frequency int  `json:"frequency"`
}

func (s *Service) searchDestinations(c *gin.Context) {
	u, ok := s.requireUser(c)
	if !ok {
		return
	}
	query := strings.ToLower(strings.TrimSpace(c.Query("q")))
	if utf8.RuneCountInString(query) > 100 {
		respondError(c, bad("Destination search is limited to 100 characters"))
		return
	}
	rows := []Destination{}
	lookupUnavailable := false
	base := s.db.WithContext(c.Request.Context())
	var err error
	if query == "" {
		err = base.Where("id IN (SELECT destination_id FROM tc_destination_usage WHERE user_id = ? GROUP BY destination_id ORDER BY max(last_selected_at) DESC LIMIT 50)", u.ID).Find(&rows).Error
		if err == nil && len(rows) == 0 {
			err = base.Where("kind = ?", "country").Order("name").Limit(20).Find(&rows).Error
		}
	} else {
		prefix := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(query) + "%"
		// Rank before bounding the candidate set, so popular prefixes cannot
		// displace an exact match or this user's own frequently selected place.
		err = base.Raw(`SELECT d.* FROM tc_destinations d
		 JOIN (SELECT destination_id, min(CASE WHEN term = ? THEN 0 ELSE 1 END) AS exactness
		       FROM tc_destination_terms WHERE term LIKE ? GROUP BY destination_id) m ON m.destination_id = d.id
		 LEFT JOIN LATERAL (SELECT count(*) AS frequency, max(last_selected_at) AS recent
		       FROM tc_destination_usage WHERE user_id = ? AND destination_id = d.id) u ON true
		 ORDER BY m.exactness, CASE WHEN lower(d.name) LIKE ? THEN 0 ELSE 1 END,
		          u.frequency DESC, u.recent DESC NULLS LAST, d.name, d.id LIMIT 200`, query, prefix, u.ID, prefix).Scan(&rows).Error
	}
	if err != nil {
		respondError(c, err)
		return
	}
	if len(rows) == 0 && query != "" && utf8.RuneCountInString(query) >= 2 && s.destinationSearch != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		candidates, err := s.destinationSearch(ctx, query)
		cancel()
		lookupUnavailable = err != nil
		if c.Request.Context().Err() == nil {
			if len(candidates) > 20 {
				candidates = candidates[:20]
			}
			found := map[uuid.UUID]bool{}
			for _, candidate := range candidates {
				// Each candidate has its own transaction; one invalid provider row
				// must not prevent the user from finding offline country records.
				var destination *Destination
				err := base.Transaction(func(tx *gorm.DB) error {
					var err error
					destination, err = s.persistDestinationCandidate(tx, candidate)
					return err
				})
				if err == nil && !found[destination.ID] {
					rows = append(rows, *destination)
					found[destination.ID] = true
				}
			}
			if len(candidates) > 0 && len(rows) == 0 {
				lookupUnavailable = true
			}
		}
	}
	ids := []uuid.UUID{}
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	ranks := []destinationRank{}
	if len(ids) > 0 {
		err = base.Model(&destinationUsage{}).Select("destination_id, count(*) AS frequency, max(last_selected_at) AS last_selected_at").Where("user_id = ? AND destination_id IN ?", u.ID, ids).Group("destination_id").Scan(&ranks).Error
		if err != nil {
			respondError(c, err)
			return
		}
	}
	usage := map[uuid.UUID]destinationRank{}
	for _, rank := range ranks {
		usage[rank.DestinationID] = rank
	}
	match := func(d Destination) int {
		if strings.EqualFold(d.Name, query) {
			return 0
		}
		for _, alias := range d.Aliases {
			if strings.EqualFold(alias, query) {
				return 0
			}
		}
		if strings.HasPrefix(strings.ToLower(d.Name), query) {
			return 1
		}
		return 2
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if query != "" && match(a) != match(b) {
			return match(a) < match(b)
		}
		ua, ub := usage[a.ID], usage[b.ID]
		if ua.Frequency != ub.Frequency {
			return ua.Frequency > ub.Frequency
		}
		if !ua.LastSelectedAt.Equal(ub.LastSelectedAt) {
			return ua.LastSelectedAt.After(ub.LastSelectedAt)
		}
		return a.Name < b.Name
	})
	if len(rows) > 20 {
		rows = rows[:20]
	}
	result := make([]destinationResult, 0, len(rows))
	for _, row := range rows {
		rank := usage[row.ID]
		result = append(result, destinationResult{DestinationSnapshot: row.snapshot(), Recent: rank.LastSelectedAt.After(time.Now().AddDate(0, 0, -30)), Frequency: rank.Frequency})
	}
	response := gin.H{"destinations": result}
	if lookupUnavailable && len(result) == 0 {
		response["lookupUnavailable"] = true
	}
	c.JSON(200, response)
}
