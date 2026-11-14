package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/decks/domain"
	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/internal/mapper"
)

var _ domain.ICardRepository = &CardRepository{}

type CardRepository struct {
	db     *gorm.DB
	mapper *mapper.DeckMapper
}

func NewCardRepository(db *gorm.DB, mapper *mapper.DeckMapper) domain.ICardRepository {
	return &CardRepository{db: db, mapper: mapper}
}

func (r *CardRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *CardRepository) List(ctx context.Context, userID string, f entity.CardFilter) ([]entity.CardEntity, int64, error) {
	q := r.conn(ctx).Model(&models.Card{}).Where("user_id = ? AND deck_id = ?", userID, f.DeckID)
	if f.Query != "" {
		like := "%" + escapeLike(f.Query) + "%"
		q = q.Where("(headword ILIKE ? OR content->>'meaning_vi' ILIKE ?)", like, like)
	}
	switch f.State {
	case "new":
		q = q.Where("state = 0 AND NOT suspended")
	case "learning":
		q = q.Where("state IN (1, 3) AND NOT suspended")
	case "review":
		q = q.Where("state = 2 AND NOT suspended")
	case "suspended":
		q = q.Where("suspended")
	case "leech":
		q = q.Where("is_leech")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.Card
	err := q.Order("position ASC, created_at ASC").
		Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	out := make([]entity.CardEntity, 0, len(rows))
	for i := range rows {
		e, err := r.mapper.CardModelToEntity(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *e)
	}
	return out, total, nil
}

func (r *CardRepository) FindByID(ctx context.Context, userID, cardID string) (*entity.CardEntity, error) {
	var row models.Card
	err := r.conn(ctx).First(&row, "id = ? AND user_id = ?", cardID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapper.CardModelToEntity(&row)
}

func (r *CardRepository) HeadwordExists(ctx context.Context, deckID, headword, exceptCardID string) (bool, error) {
	q := r.conn(ctx).Model(&models.Card{}).Where("deck_id = ? AND lower(headword) = lower(?)", deckID, headword)
	if exceptCardID != "" {
		q = q.Where("id <> ?", exceptCardID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *CardRepository) NextPosition(ctx context.Context, deckID string) (int, error) {
	var max *int
	err := r.conn(ctx).Model(&models.Card{}).Where("deck_id = ?", deckID).Select("max(position)").Scan(&max).Error
	if err != nil || max == nil {
		return 1, err
	}
	return *max + 1, nil
}

func (r *CardRepository) Create(ctx context.Context, card *entity.CardEntity) (*entity.CardEntity, error) {
	content, err := json.Marshal(card.Content)
	if err != nil {
		return nil, err
	}
	model := &models.Card{
		UserID:         card.UserID,
		DeckID:         card.DeckID,
		LexemeID:       card.LexemeID,
		Headword:       card.Headword,
		Content:        content,
		ContentVersion: 1,
		State:          entity.StateNew,
		Due:            card.Due,
		Position:       card.Position,
	}
	if err := r.conn(ctx).Omit("ID").Create(model).Error; err != nil {
		return nil, err
	}
	return r.mapper.CardModelToEntity(model)
}

func (r *CardRepository) UpdateContent(ctx context.Context, card *entity.CardEntity) error {
	content, err := json.Marshal(card.Content)
	if err != nil {
		return err
	}
	return r.conn(ctx).Model(&models.Card{}).
		Where("id = ? AND user_id = ?", card.ID, card.UserID).
		Updates(map[string]any{"headword": card.Headword, "content": content, "updated_at": time.Now()}).Error
}

func (r *CardRepository) Bulk(ctx context.Context, userID string, ids []string, action domain.BulkAction, targetDeckID string, now time.Time) (int64, error) {
	db := r.conn(ctx)
	var res *gorm.DB
	switch action {
	case domain.BulkSuspend, domain.BulkUnsuspend:
		res = db.Model(&models.Card{}).Where("user_id = ? AND id IN ?", userID, ids).
			Updates(map[string]any{"suspended": action == domain.BulkSuspend, "updated_at": now})
	case domain.BulkDelete:
		res = db.Where("user_id = ? AND id IN ?", userID, ids).Delete(&models.Card{})
	case domain.BulkReset:
		// Back to "new" at the end of its deck's new-card queue. Review history is kept.
		res = db.Exec(`
			UPDATE cards c SET
				state = 0, due = ?, stability = 0, difficulty = 0, scheduled_days = 0,
				remaining_steps = 0, reps = 0, lapses = 0, last_review_at = NULL,
				is_leech = FALSE, suspended = FALSE, updated_at = ?,
				position = (SELECT coalesce(max(position), 0) FROM cards c2 WHERE c2.deck_id = c.deck_id) + s.rn
			FROM (SELECT id, row_number() OVER (ORDER BY position) AS rn FROM cards WHERE user_id = ? AND id IN ?) s
			WHERE c.id = s.id`, now, now, userID, ids)
	case domain.BulkMove:
		res = db.Exec(`
			UPDATE cards c SET
				deck_id = ?, updated_at = ?,
				position = (SELECT coalesce(max(position), 0) FROM cards c2 WHERE c2.deck_id = ?) + s.rn
			FROM (SELECT id, row_number() OVER (ORDER BY position) AS rn FROM cards WHERE user_id = ? AND id IN ? AND deck_id <> ?) s
			WHERE c.id = s.id`, targetDeckID, now, targetDeckID, userID, ids, targetDeckID)
	default:
		return 0, errors.New("unknown bulk action")
	}
	return res.RowsAffected, res.Error
}
