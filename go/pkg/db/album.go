package db

import (
	"context"

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
