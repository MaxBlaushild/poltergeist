package db

import (
	"context"
	"sort"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type albumHandle struct {
	db *gorm.DB
}

func (h *albumHandle) Create(ctx context.Context, userID uuid.UUID, name string, tags []string) (*models.Album, error) {
	album := &models.Album{
		UserID: userID,
		Name:   name,
	}
	if err := h.db.WithContext(ctx).Create(album).Error; err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		records := make([]models.AlbumTag, 0, len(tags))
		seen := make(map[string]bool)
		for _, tag := range tags {
			t := trimTag(tag)
			if t == "" || seen[t] {
				continue
			}
			seen[t] = true
			records = append(records, models.AlbumTag{AlbumID: album.ID, Tag: t})
		}
		if len(records) > 0 {
			if err := h.db.WithContext(ctx).Create(&records).Error; err != nil {
				return nil, err
			}
		}
	}
	return album, nil
}

func (h *albumHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Album, error) {
	var albums []models.Album
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}

// FindAccessibleAlbumIDs returns album IDs the user can access: owned, member, or accepted invite.
func (h *albumHandle) FindAccessibleAlbumIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Owned
	var owned []uuid.UUID
	h.db.WithContext(ctx).Model(&models.Album{}).Where("user_id = ?", userID).Pluck("id", &owned)
	seen := make(map[uuid.UUID]bool)
	for _, id := range owned {
		seen[id] = true
	}
	// Member (via album_members - but we need albumMemberHandle, which we don't have)
	// Use raw query
	var memberIDs []uuid.UUID
	h.db.WithContext(ctx).Table("album_members").Where("user_id = ?", userID).Pluck("album_id", &memberIDs)
	for _, id := range memberIDs {
		seen[id] = true
	}
	// Accepted invites
	var acceptedIDs []uuid.UUID
	h.db.WithContext(ctx).Table("album_invites").
		Where("invited_user_id = ? AND status = ?", userID, "accepted").
		Pluck("album_id", &acceptedIDs)
	for _, id := range acceptedIDs {
		seen[id] = true
	}
	result := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result, nil
}

func (h *albumHandle) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Album, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var albums []models.Album
	if err := h.db.WithContext(ctx).Where("id IN ?", ids).Order("created_at DESC").Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}

func (h *albumHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Album, error) {
	var album models.Album
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&album).Error; err != nil {
		return nil, err
	}
	return &album, nil
}

func (h *albumHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Album{}, "id = ?", id).Error
}

func (h *albumHandle) GetTags(ctx context.Context, albumID uuid.UUID) ([]string, error) {
	var rows []models.AlbumTag
	if err := h.db.WithContext(ctx).Where("album_id = ?", albumID).Find(&rows).Error; err != nil {
		return nil, err
	}
	tags := make([]string, len(rows))
	for i := range rows {
		tags[i] = rows[i].Tag
	}
	return tags, nil
}

func (h *albumHandle) AddTag(ctx context.Context, albumID uuid.UUID, tag string) error {
	t := trimTag(tag)
	if t == "" {
		return nil
	}
	return h.db.WithContext(ctx).Create(&models.AlbumTag{AlbumID: albumID, Tag: t}).Error
}

func (h *albumHandle) RemoveTag(ctx context.Context, albumID uuid.UUID, tag string) error {
	return h.db.WithContext(ctx).Where("album_id = ? AND tag = ?", albumID, tag).
		Delete(&models.AlbumTag{}).Error
}

// FindPostsForAlbum returns posts from the user's feed (user + friends) that have at least one of the album's tags
func (h *albumHandle) FindPostsForAlbum(ctx context.Context, userID uuid.UUID, albumTags []string) ([]models.Post, error) {
	if len(albumTags) == 0 {
		return []models.Post{}, nil
	}
	// Get friend IDs + self
	friendIDs, err := h.getFriendIDsAndSelf(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(friendIDs) == 0 {
		return []models.Post{}, nil
	}
	// Find post IDs that have any of the album tags
	var postIDs []uuid.UUID
	if err := h.db.WithContext(ctx).Table("post_tags").
		Select("DISTINCT post_id").
		Where("tag IN ?", albumTags).
		Pluck("post_id", &postIDs).Error; err != nil {
		return nil, err
	}
	if len(postIDs) == 0 {
		return []models.Post{}, nil
	}
	// Get posts where id in postIDs and user_id in friendIDs, ordered by created_at DESC
	var posts []models.Post
	if err := h.db.WithContext(ctx).
		Where("id IN ? AND user_id IN ?", postIDs, friendIDs).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetAlbumTagsForUserOrderedByAssociation returns tags from albums the user can access,
// ordered reverse chronologically by when they gained access (owner=created, member=joined, invite=accepted).
func (h *albumHandle) GetAlbumTagsForUserOrderedByAssociation(ctx context.Context, userID uuid.UUID) ([]string, error) {
	tagToTime := make(map[string]string) // tag -> max sortable timestamp string for consistency

	// Owners: albums where user_id = userID
	var ownerRows []struct {
		Tag  string
		Time string
	}
	h.db.WithContext(ctx).Table("album_tags").
		Select("album_tags.tag as tag, albums.created_at::text as time").
		Joins("JOIN albums ON albums.id = album_tags.album_id").
		Where("albums.user_id = ?", userID).
		Scan(&ownerRows)
	for _, r := range ownerRows {
		t := trimTag(r.Tag)
		if t == "" {
			continue
		}
		if r.Time > tagToTime[t] {
			tagToTime[t] = r.Time
		}
	}

	// Members: album_members where user_id = userID
	var memberRows []struct {
		Tag  string
		Time string
	}
	h.db.WithContext(ctx).Table("album_tags").
		Select("album_tags.tag as tag, album_members.created_at::text as time").
		Joins("JOIN album_members ON album_members.album_id = album_tags.album_id").
		Where("album_members.user_id = ?", userID).
		Scan(&memberRows)
	for _, r := range memberRows {
		t := trimTag(r.Tag)
		if t == "" {
			continue
		}
		if r.Time > tagToTime[t] {
			tagToTime[t] = r.Time
		}
	}

	// Accepted invites: use accepted_at if set, else created_at (for legacy rows)
	var inviteRows []struct {
		Tag  string
		Time string
	}
	h.db.WithContext(ctx).Table("album_tags").
		Select("album_tags.tag as tag, COALESCE(album_invites.accepted_at::text, album_invites.created_at::text) as time").
		Joins("JOIN album_invites ON album_invites.album_id = album_tags.album_id").
		Where("album_invites.invited_user_id = ? AND album_invites.status = ?", userID, "accepted").
		Scan(&inviteRows)
	for _, r := range inviteRows {
		t := trimTag(r.Tag)
		if t == "" {
			continue
		}
		if r.Time > tagToTime[t] {
			tagToTime[t] = r.Time
		}
	}

	// Sort tags by sort time descending
	type pair struct {
		tag  string
		time string
	}
	pairs := make([]pair, 0, len(tagToTime))
	for tag, t := range tagToTime {
		pairs = append(pairs, pair{tag, t})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].time > pairs[j].time })
	result := make([]string, len(pairs))
	for i, p := range pairs {
		result[i] = p.tag
	}
	return result, nil
}

func (h *albumHandle) getFriendIDsAndSelf(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var friends []models.Friend
	if err := h.db.WithContext(ctx).Find(&friends).Error; err != nil {
		return nil, err
	}
	ids := []uuid.UUID{userID}
	for _, f := range friends {
		if f.FirstUserID == userID {
			ids = append(ids, f.SecondUserID)
		} else if f.SecondUserID == userID {
			ids = append(ids, f.FirstUserID)
		}
	}
	return ids, nil
}
