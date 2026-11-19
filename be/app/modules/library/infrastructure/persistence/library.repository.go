// Package persistence is the PostgreSQL (GORM) adapter of the library module.
package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/library/domain"
	"ielts-practice/app/modules/library/domain/entity"
	"ielts-practice/app/modules/library/domain/rules"
)

var _ domain.ILibraryRepository = &LibraryRepository{}

type LibraryRepository struct {
	db *gorm.DB
}

func NewLibraryRepository(db *gorm.DB) domain.ILibraryRepository {
	return &LibraryRepository{db: db}
}

func (r *LibraryRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func toCollection(m *models.LibraryCollection) entity.Collection {
	c := entity.Collection{ID: m.ID, Slug: m.Slug, Kind: m.Kind, Sort: int(m.Sort), Status: m.Status, ContentHash: m.ContentHash}
	_ = json.Unmarshal(m.Title, &c.Title)
	_ = json.Unmarshal(m.Description, &c.Description)
	return c
}

func (r *LibraryRepository) ListCollections(ctx context.Context) ([]entity.Collection, error) {
	var rows []models.LibraryCollection
	if err := r.conn(ctx).Where("status = ?", entity.StatusPublished).Order("sort, slug").Find(&rows).Error; err != nil {
		return nil, err
	}
	var counts []struct {
		CollectionID string
		TopicID      string
		Words        int
	}
	err := r.conn(ctx).Raw(`
		SELECT t.collection_id, t.topic_id, count(w.id) AS words
		FROM library_topics t LEFT JOIN library_words w ON w.library_topic_id = t.id
		GROUP BY t.id ORDER BY t.sort, t.topic_id`).Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	out := make([]entity.Collection, 0, len(rows))
	for i := range rows {
		c := toCollection(&rows[i])
		for _, tc := range counts {
			if tc.CollectionID == c.ID && tc.Words > 0 {
				c.Topics = append(c.Topics, entity.TopicCount{TopicID: tc.TopicID, Words: tc.Words})
			}
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *LibraryRepository) FindCollection(ctx context.Context, slug string) (*entity.Collection, error) {
	var row models.LibraryCollection
	err := r.conn(ctx).Where("slug = ?", slug).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c := toCollection(&row)
	return &c, nil
}

func (r *LibraryRepository) ListWords(ctx context.Context, userID, collectionID string, f entity.WordFilter) ([]entity.Word, error) {
	q := r.conn(ctx).Table("library_words w").
		Select(`w.id, t.topic_id, w.lexeme_id, w.cefr, w.source, l.content,
			EXISTS (SELECT 1 FROM cards c WHERE c.user_id = ? AND (c.lexeme_id = w.lexeme_id OR lower(c.headword) = lower(l.headword))) AS in_deck,
			EXISTS (SELECT 1 FROM user_known_words k WHERE k.user_id = ? AND k.lemma = lower(l.headword)) AS known`, userID, userID).
		Joins("JOIN library_topics t ON t.id = w.library_topic_id").
		Joins("JOIN lexemes l ON l.id = w.lexeme_id AND l.status = 'ready'").
		Where("t.collection_id = ?", collectionID)
	if f.TopicID != "" {
		q = q.Where("t.topic_id = ?", f.TopicID)
	}
	if f.Level != "" {
		q = q.Where("w.cefr = ?", f.Level)
	}
	if f.Query != "" {
		q = q.Where("w.search_text LIKE ?", rules.LikePattern(f.Query))
	}
	var rows []struct {
		ID       string
		TopicID  string
		LexemeID string
		CEFR     string `gorm:"column:cefr"`
		Source   string
		Content  []byte
		InDeck   bool
		Known    bool
	}
	if err := q.Order("t.sort, t.topic_id, w.sort").Limit(f.Limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Word, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		if seen[row.LexemeID] { // a word in two topics is listed once, under its first topic
			continue
		}
		seen[row.LexemeID] = true
		w := entity.Word{ID: row.ID, TopicID: row.TopicID, LexemeID: row.LexemeID, CEFR: row.CEFR, Source: row.Source, InDeck: row.InDeck, Known: row.Known}
		if err := json.Unmarshal(row.Content, &w.Card); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func (r *LibraryRepository) TopicHeadwords(ctx context.Context, topicID string) ([]string, error) {
	var words []string
	err := r.conn(ctx).Raw(`
		SELECT DISTINCT lower(l.headword) FROM library_words w
		JOIN library_topics t ON t.id = w.library_topic_id
		JOIN library_collections c ON c.id = t.collection_id AND c.status = 'published'
		JOIN lexemes l ON l.id = w.lexeme_id
		WHERE t.topic_id = ?`, topicID).Scan(&words).Error
	return words, err
}

func (r *LibraryRepository) FindCards(ctx context.Context, lexemeIDs []string) (map[string]entity.CardContent, error) {
	var rows []models.Lexeme
	err := r.conn(ctx).Where("id IN ? AND origin = ? AND status = 'ready' AND content IS NOT NULL", lexemeIDs, entity.OriginLibrary).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]entity.CardContent, len(rows))
	for _, row := range rows {
		var c entity.CardContent
		if err := json.Unmarshal(row.Content, &c); err != nil {
			return nil, err
		}
		out[row.ID] = c
	}
	return out, nil
}

// ---- seed ---------------------------------------------------------------------

func (r *LibraryRepository) UpsertCollection(ctx context.Context, c *entity.Collection) error {
	title, _ := json.Marshal(c.Title)
	desc, _ := json.Marshal(c.Description)
	var ids []string
	err := r.conn(ctx).Raw(`
		INSERT INTO library_collections (slug, kind, title, description, sort, status, content_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (slug) DO UPDATE SET kind = EXCLUDED.kind, title = EXCLUDED.title, description = EXCLUDED.description,
			sort = EXCLUDED.sort, status = EXCLUDED.status, content_hash = EXCLUDED.content_hash, updated_at = now()
		RETURNING id`, c.Slug, c.Kind, title, desc, c.Sort, c.Status, c.ContentHash).Scan(&ids).Error
	if err != nil {
		return err
	}
	if len(ids) == 1 {
		c.ID = ids[0]
	}
	return nil
}

func (r *LibraryRepository) TopicHash(ctx context.Context, collectionID, topicID string) (string, error) {
	var hashes []string
	err := r.conn(ctx).Model(&models.LibraryTopic{}).Where("collection_id = ? AND topic_id = ?", collectionID, topicID).
		Pluck("content_hash", &hashes).Error
	if err != nil || len(hashes) == 0 {
		return "", err
	}
	return hashes[0], nil
}

func (r *LibraryRepository) ReplaceTopic(ctx context.Context, collectionID, topicID string, sort int, hash string, words []entity.SeedWord) error {
	return r.conn(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []string
		err := tx.Raw(`
			INSERT INTO library_topics (collection_id, topic_id, sort, content_hash) VALUES (?, ?, ?, '')
			ON CONFLICT (collection_id, topic_id) DO UPDATE SET sort = EXCLUDED.sort, updated_at = now()
			RETURNING id`, collectionID, topicID, sort).Scan(&ids).Error
		if err != nil {
			return err
		}
		libTopicID := ids[0]
		keep := make([]string, 0, len(words))
		for _, w := range words {
			lexemeID, err := upsertLexeme(tx, w)
			if err != nil {
				return err
			}
			keep = append(keep, lexemeID)
			err = tx.Exec(`
				INSERT INTO library_words (library_topic_id, lexeme_id, cefr, source, sort, search_text)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT (library_topic_id, lexeme_id) DO UPDATE SET cefr = EXCLUDED.cefr, source = EXCLUDED.source,
					sort = EXCLUDED.sort, search_text = EXCLUDED.search_text`,
				libTopicID, lexemeID, w.CEFR, w.Source, w.Sort, w.SearchText).Error
			if err != nil {
				return err
			}
		}
		del := tx.Where("library_topic_id = ?", libTopicID)
		if len(keep) > 0 {
			del = del.Where("lexeme_id NOT IN ?", keep)
		}
		if err := del.Delete(&models.LibraryWord{}).Error; err != nil {
			return err
		}
		return tx.Model(&models.LibraryTopic{}).Where("id = ?", libTopicID).Update("content_hash", hash).Error
	})
}

func (r *LibraryRepository) UpsertLibraryLexeme(ctx context.Context, w entity.SeedWord) (string, error) {
	return upsertLexeme(r.conn(ctx), w)
}

func upsertLexeme(db *gorm.DB, w entity.SeedWord) (string, error) {
	content, err := json.Marshal(w.Card)
	if err != nil {
		return "", err
	}
	candidates := []byte("[]")
	if w.Card.Image != nil {
		candidates, _ = json.Marshal([]any{w.Card.Image})
	}
	var ids []string
	err = db.Raw(`
		INSERT INTO lexemes (key, headword, lemma, sense_hint, meaning_lang, content, status, image_query, image_candidates,
		                     quality_flags, origin)
		VALUES (?, ?, ?, '', 'vi', ?, 'ready', ?, ?, '[]', 'library')
		ON CONFLICT (key) DO UPDATE SET headword = EXCLUDED.headword, lemma = EXCLUDED.lemma, content = EXCLUDED.content,
			status = 'ready', suggestion = '', image_query = EXCLUDED.image_query, image_candidates = EXCLUDED.image_candidates,
			quality_flags = '[]', report_count = 0, claimed_at = NULL, origin = 'library', updated_at = now()
		RETURNING id`, w.Key, w.Card.Headword, w.Card.Lemma, content, w.ImageQuery, candidates).Scan(&ids).Error
	if err != nil {
		return "", err
	}
	if w.PixabayImage && w.Card.Image != nil {
		// Pixabay links expire: register the photo already expired, so its first view fetches a fresh link.
		err = db.Exec(`
			INSERT INTO image_links (provider, provider_id, url, thumb_url, expires_at, query, headword)
			VALUES ('pixabay', ?, '', '', ?, ?, ?)
			ON CONFLICT (provider, provider_id) DO NOTHING`,
			w.Card.Image.ProviderID, time.Unix(0, 0), w.ImageQuery, w.Card.Headword).Error
		if err != nil {
			return "", err
		}
	}
	return ids[0], nil
}

func (r *LibraryRepository) FindLibraryLexeme(ctx context.Context, key string, wordsOnly bool) (string, error) {
	var ids []string
	q := r.conn(ctx).Model(&models.Lexeme{}).Where("key = ? AND origin = ? AND status = 'ready'", key, entity.OriginLibrary)
	if wordsOnly {
		q = q.Where("EXISTS (SELECT 1 FROM library_words w WHERE w.lexeme_id = lexemes.id)")
	}
	err := q.Pluck("id", &ids).Error
	if err != nil || len(ids) == 0 {
		return "", err
	}
	return ids[0], nil
}
