package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/decks/domain"
	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/internal/mapper"
)

var _ domain.IDeckRepository = &DeckRepository{}

type DeckRepository struct {
	db     *gorm.DB
	mapper *mapper.DeckMapper
}

func NewDeckRepository(db *gorm.DB, mapper *mapper.DeckMapper) domain.IDeckRepository {
	return &DeckRepository{db: db, mapper: mapper}
}

type deckWithCount struct {
	models.Deck
	CardCount int
}

func (r *DeckRepository) query(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Table("decks").
		Select("decks.*, (SELECT count(*) FROM cards WHERE cards.deck_id = decks.id) AS card_count").
		Where("decks.archived_at IS NULL")
}

func (r *DeckRepository) List(ctx context.Context, userID string) ([]entity.DeckEntity, error) {
	var rows []deckWithCount
	if err := r.query(ctx).Where("decks.user_id = ?", userID).Order("decks.created_at ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.DeckEntity, 0, len(rows))
	for i := range rows {
		out = append(out, *r.mapper.DeckModelToEntity(&rows[i].Deck, rows[i].CardCount))
	}
	return out, nil
}

func (r *DeckRepository) FindByID(ctx context.Context, userID, deckID string) (*entity.DeckEntity, error) {
	var rows []deckWithCount
	if err := r.query(ctx).Where("decks.user_id = ? AND decks.id = ?", userID, deckID).Limit(1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return r.mapper.DeckModelToEntity(&rows[0].Deck, rows[0].CardCount), nil
}

func (r *DeckRepository) NameTaken(ctx context.Context, userID, name, exceptDeckID string) (bool, error) {
	q := database.GetDBFromContext(ctx, r.db).WithContext(ctx).Model(&models.Deck{}).
		Where("user_id = ? AND lower(name) = lower(?) AND archived_at IS NULL", userID, name)
	if exceptDeckID != "" {
		q = q.Where("id <> ?", exceptDeckID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *DeckRepository) Create(ctx context.Context, deck *entity.DeckEntity) (*entity.DeckEntity, error) {
	model := r.mapper.DeckEntityToModel(deck)
	model.ID = ""
	if err := database.GetDBFromContext(ctx, r.db).WithContext(ctx).Omit("ID").Create(model).Error; err != nil {
		return nil, err
	}
	return r.mapper.DeckModelToEntity(model, 0), nil
}

func (r *DeckRepository) Update(ctx context.Context, deck *entity.DeckEntity) error {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Model(&models.Deck{ID: deck.ID}).
		Where("user_id = ?", deck.UserID).
		Updates(map[string]any{
			"name":            deck.Name,
			"description":     deck.Description,
			"new_per_day":     deck.NewPerDay,
			"reviews_per_day": deck.ReviewsPerDay,
			"updated_at":      time.Now(),
		}).Error
}

func (r *DeckRepository) Archive(ctx context.Context, userID, deckID string, at time.Time) error {
	res := database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Model(&models.Deck{}).
		Where("id = ? AND user_id = ? AND archived_at IS NULL", deckID, userID).
		Update("archived_at", at)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
