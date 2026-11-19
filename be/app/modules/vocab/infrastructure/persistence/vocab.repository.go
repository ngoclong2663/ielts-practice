// Package persistence is the PostgreSQL (GORM) adapter of the vocab module.
package persistence

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/internal/mapper"
)

var _ domain.IVocabRepository = &VocabRepository{}

type VocabRepository struct {
	db     *gorm.DB
	mapper *mapper.VocabMapper
}

func NewVocabRepository(db *gorm.DB, m *mapper.VocabMapper) domain.IVocabRepository {
	return &VocabRepository{db: db, mapper: m}
}

func (r *VocabRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func notFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// ---- lexemes ---------------------------------------------------------------

func (r *VocabRepository) FindLexemeByKey(ctx context.Context, key string) (*entity.Lexeme, error) {
	var row models.Lexeme
	err := r.conn(ctx).First(&row, "key = ?", key).Error
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapper.LexemeModelToEntity(&row), nil
}

func (r *VocabRepository) FindLexemeByID(ctx context.Context, id string) (*entity.Lexeme, error) {
	var row models.Lexeme
	err := r.conn(ctx).First(&row, "id = ?", id).Error
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapper.LexemeModelToEntity(&row), nil
}

func (r *VocabRepository) ClaimLexeme(ctx context.Context, key, headword, senseHint string, staleBefore time.Time) (domain.ClaimResult, error) {
	var ids []string
	// Take over failed, reported or abandoned rows; never steal a fresh pending claim, nor a library card
	// (Epic 8: reviewed by hand, so the AI never rewrites it).
	err := r.conn(ctx).Raw(`
		INSERT INTO lexemes (key, headword, sense_hint, meaning_lang, status, claimed_at)
		VALUES (?, ?, ?, ?, 'pending', now())
		ON CONFLICT (key) DO UPDATE SET status = 'pending', claimed_at = now(), updated_at = now()
		WHERE lexemes.origin <> 'library' AND (lexemes.status IN ('failed', 'needs_review')
		   OR (lexemes.status = 'pending' AND (lexemes.claimed_at IS NULL OR lexemes.claimed_at < ?)))
		RETURNING id`, key, headword, senseHint, entity.MeaningLangVI, staleBefore).Scan(&ids).Error
	if err != nil {
		return domain.ClaimResult{}, err
	}
	if len(ids) == 1 {
		lex, err := r.FindLexemeByID(ctx, ids[0])
		return domain.ClaimResult{Lexeme: lex, Owned: true}, err
	}
	lex, err := r.FindLexemeByKey(ctx, key)
	return domain.ClaimResult{Lexeme: lex}, err
}

func (r *VocabRepository) SaveLexeme(ctx context.Context, l *entity.Lexeme) error {
	updates := map[string]any{
		"headword":         l.Headword,
		"lemma":            l.Lemma,
		"status":           l.Status,
		"suggestion":       l.Suggestion,
		"image_query":      l.ImageQuery,
		"image_candidates": mapper.MarshalJSON(l.ImageCandidates, "[]"),
		"quality_flags":    mapper.MarshalJSON(l.QualityFlags, "[]"),
		"topic":            l.Topic,
		"claimed_at":       nil,
		"updated_at":       time.Now(),
	}
	if l.Content != nil {
		updates["content"] = mapper.MarshalJSON(l.Content, "null")
	}
	if l.Status == entity.LexemeReady {
		updates["report_count"] = 0
	}
	return r.conn(ctx).Model(&models.Lexeme{}).Where("id = ?", l.ID).Updates(updates).Error
}

func (r *VocabRepository) SetLexemeStatus(ctx context.Context, id, status string) error {
	return r.conn(ctx).Model(&models.Lexeme{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "claimed_at": nil, "updated_at": time.Now()}).Error
}

func (r *VocabRepository) ReportLexeme(ctx context.Context, lexemeID, userID, reason, note string, threshold int) (bool, error) {
	res := r.conn(ctx).Exec(`
		INSERT INTO lexeme_reports (lexeme_id, user_id, reason, note) VALUES (?, ?, ?, ?)
		ON CONFLICT (lexeme_id, user_id) DO NOTHING`, lexemeID, userID, reason, note)
	if res.Error != nil || res.RowsAffected == 0 {
		return false, res.Error
	}
	err := r.conn(ctx).Exec(`
		UPDATE lexemes SET report_count = report_count + 1,
		       status = CASE WHEN report_count + 1 >= ? AND status = 'ready' AND origin <> 'library'
		                     THEN 'needs_review' ELSE status END,
		       updated_at = now()
		WHERE id = ?`, threshold, lexemeID).Error
	return true, err
}

// ---- jobs ------------------------------------------------------------------

func (r *VocabRepository) CreateJob(ctx context.Context, job *entity.Job) (*entity.Job, error) {
	row := &models.GenerationJob{
		UserID:       job.UserID,
		TargetDeckID: job.TargetDeckID,
		Status:       job.Status,
		Total:        len(job.Items),
		Done:         job.Done,
	}
	if err := r.conn(ctx).Omit("ID").Create(row).Error; err != nil {
		return nil, err
	}
	for i := range job.Items {
		it := &job.Items[i]
		model := &models.GenerationJobItem{
			JobID:           row.ID,
			UserID:          job.UserID,
			Position:        it.Position,
			InputText:       it.InputText,
			SenseHint:       it.SenseHint,
			LexemeID:        it.LexemeID,
			Status:          it.Status,
			Charged:         it.Charged,
			Cached:          it.Cached,
			Content:         mapper.MarshalJSON(it.Content, "null"),
			ImageCandidates: mapper.MarshalJSON(it.ImageCandidates, "[]"),
			ErrorCode:       it.ErrorCode,
			Suggestion:      it.Suggestion,
			NextAttemptAt:   time.Now(),
		}
		if it.Content == nil {
			model.Content = nil
		}
		if err := r.conn(ctx).Omit("ID").Create(model).Error; err != nil {
			return nil, err
		}
	}
	if err := r.RefreshJobProgress(ctx, row.ID, time.Now()); err != nil {
		return nil, err
	}
	return r.FindJob(ctx, job.UserID, row.ID)
}

func (r *VocabRepository) FindJob(ctx context.Context, userID, jobID string) (*entity.Job, error) {
	var row models.GenerationJob
	err := r.conn(ctx).First(&row, "id = ? AND user_id = ?", jobID, userID).Error
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job := r.mapper.JobModelToEntity(&row)
	var items []models.GenerationJobItem
	if err := r.conn(ctx).Where("job_id = ?", jobID).Order("position ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	for i := range items {
		job.Items = append(job.Items, *r.mapper.ItemModelToEntity(&items[i]))
	}
	return job, nil
}

func (r *VocabRepository) ListJobs(ctx context.Context, userID string, limit int) ([]entity.Job, error) {
	var rows []models.GenerationJob
	if err := r.conn(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Job, 0, len(rows))
	ids := make([]string, 0, len(rows))
	for i := range rows {
		out = append(out, *r.mapper.JobModelToEntity(&rows[i]))
		ids = append(ids, rows[i].ID)
	}
	if len(ids) == 0 {
		return out, nil
	}
	// Only the first words of each job, for the "Recent generations" list.
	var items []models.GenerationJobItem
	err := r.conn(ctx).Select("job_id, position, input_text, status, removed").
		Where("job_id IN ? AND position < 6", ids).Order("position ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	byJob := map[string]int{}
	for i := range out {
		byJob[out[i].ID] = i
	}
	for i := range items {
		j := byJob[items[i].JobID]
		out[j].Items = append(out[j].Items, *r.mapper.ItemModelToEntity(&items[i]))
	}
	return out, nil
}

func (r *VocabRepository) RefreshJobProgress(ctx context.Context, jobID string, now time.Time) error {
	return r.conn(ctx).Exec(`
		UPDATE generation_jobs j SET
		  total = s.total,
		  done = s.done,
		  status = CASE
		    WHEN s.done < s.total THEN CASE WHEN s.started = 0 THEN 'queued' ELSE 'running' END
		    WHEN s.ready = s.total THEN 'done'
		    WHEN s.ready = 0 THEN 'failed'
		    ELSE 'partial' END,
		  finished_at = CASE WHEN s.done >= s.total THEN coalesce(j.finished_at, ?) ELSE NULL END,
		  updated_at = now()
		FROM (
		  SELECT count(*) AS total,
		         count(*) FILTER (WHERE status IN ('ready', 'failed', 'invalid')) AS done,
		         count(*) FILTER (WHERE status = 'ready') AS ready,
		         count(*) FILTER (WHERE status <> 'queued' OR attempts > 0) AS started
		  FROM generation_job_items WHERE job_id = ? AND NOT removed
		) s
		WHERE j.id = ?`, now, jobID, jobID).Error
}

func (r *VocabRepository) AddJobUsage(ctx context.Context, jobID string, u entity.Usage) error {
	return r.conn(ctx).Exec(`
		UPDATE generation_jobs SET input_tokens = input_tokens + ?, output_tokens = output_tokens + ?,
		       cache_read_tokens = cache_read_tokens + ? WHERE id = ?`,
		u.InputTokens, u.OutputTokens, u.CacheReadTokens, jobID).Error
}

func (r *VocabRepository) SetJobCommitted(ctx context.Context, jobID, deckID string) error {
	return r.conn(ctx).Model(&models.GenerationJob{}).Where("id = ?", jobID).
		Updates(map[string]any{"committed_deck_id": deckID, "updated_at": time.Now()}).Error
}

// ---- items -----------------------------------------------------------------

func (r *VocabRepository) FindItem(ctx context.Context, userID, jobID, itemID string) (*entity.JobItem, error) {
	var row models.GenerationJobItem
	err := r.conn(ctx).First(&row, "id = ? AND job_id = ? AND user_id = ?", itemID, jobID, userID).Error
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapper.ItemModelToEntity(&row), nil
}

func (r *VocabRepository) UpdateItem(ctx context.Context, it *entity.JobItem) error {
	content := mapper.MarshalJSON(it.Content, "null")
	if it.Content == nil {
		content = nil
	}
	return r.conn(ctx).Model(&models.GenerationJobItem{}).Where("id = ?", it.ID).Updates(map[string]any{
		"input_text":       it.InputText,
		"sense_hint":       it.SenseHint,
		"lexeme_id":        it.LexemeID,
		"status":           it.Status,
		"regenerate":       it.Regenerate,
		"charged":          it.Charged,
		"cached":           it.Cached,
		"content":          content,
		"image_candidates": mapper.MarshalJSON(it.ImageCandidates, "[]"),
		"error_code":       it.ErrorCode,
		"suggestion":       it.Suggestion,
		"attempts":         it.Attempts,
		"next_attempt_at":  it.NextAttemptAt,
		"removed":          it.Removed,
		"updated_at":       time.Now(),
	}).Error
}

func (r *VocabRepository) ClaimQueuedItems(ctx context.Context, limit int, now time.Time) ([]entity.JobItem, error) {
	var rows []models.GenerationJobItem
	err := r.conn(ctx).Raw(`
		UPDATE generation_job_items SET status = 'content', attempts = attempts + 1, updated_at = now()
		WHERE id IN (
		  SELECT id FROM generation_job_items
		  WHERE status = 'queued' AND NOT removed AND next_attempt_at <= ?
		  ORDER BY created_at, position
		  LIMIT ? FOR UPDATE SKIP LOCKED
		)
		RETURNING *`, now, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]entity.JobItem, 0, len(rows))
	for i := range rows {
		out = append(out, *r.mapper.ItemModelToEntity(&rows[i]))
	}
	return out, nil
}

func (r *VocabRepository) RequeueStuckItems(ctx context.Context, before time.Time) (int64, error) {
	res := r.conn(ctx).Exec(`
		UPDATE generation_job_items SET status = 'queued', next_attempt_at = now(), updated_at = now()
		WHERE status IN ('content', 'image') AND updated_at < ?`, before)
	return res.RowsAffected, res.Error
}

func (r *VocabRepository) ChargedWordsSince(ctx context.Context, userID string, since time.Time) (int, error) {
	var n int64
	err := r.conn(ctx).Model(&models.GenerationJobItem{}).
		Where("user_id = ? AND charged AND created_at >= ?", userID, since).Count(&n).Error
	return int(n), err
}

// ---- image links -------------------------------------------------------------

func (r *VocabRepository) UpsertImageLink(ctx context.Context, l *entity.ImageLink) error {
	row := &models.ImageLink{
		Provider:    l.Provider,
		ProviderID:  l.ProviderID,
		URL:         l.URL,
		ThumbURL:    l.ThumbURL,
		ExpiresAt:   l.ExpiresAt,
		Query:       l.Query,
		Headword:    l.Headword,
		Replacement: mapper.MarshalJSON(l.Replacement, "null"),
		Gone:        l.Gone,
	}
	if l.Replacement == nil {
		row.Replacement = nil
	}
	return r.conn(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "provider"}, {Name: "provider_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"url": row.URL, "thumb_url": row.ThumbURL, "expires_at": row.ExpiresAt,
			"replacement": row.Replacement, "gone": row.Gone, "updated_at": time.Now(),
			// Keep the first query/headword we saw; they drive replacement searches.
			"query":    gorm.Expr("CASE WHEN image_links.query = '' THEN EXCLUDED.query ELSE image_links.query END"),
			"headword": gorm.Expr("CASE WHEN image_links.headword = '' THEN EXCLUDED.headword ELSE image_links.headword END"),
		}),
	}).Create(row).Error
}

func (r *VocabRepository) FindImageLink(ctx context.Context, provider, providerID string) (*entity.ImageLink, error) {
	var row models.ImageLink
	err := r.conn(ctx).First(&row, "provider = ? AND provider_id = ?", provider, providerID).Error
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapper.ImageLinkModelToEntity(&row), nil
}

func (r *VocabRepository) ReplaceImageEverywhere(ctx context.Context, provider, providerID string, replacement *entity.CardImage) error {
	match := `content->'image'->>'provider' = ? AND content->'image'->>'provider_id' = ?`
	setImage := gorm.Expr("content - 'image'")
	if replacement != nil {
		setImage = gorm.Expr("jsonb_set(content, '{image}', ?::jsonb)", string(mapper.MarshalJSON(replacement, "null")))
	}
	dropCandidate := gorm.Expr(`(SELECT coalesce(jsonb_agg(e), '[]'::jsonb) FROM jsonb_array_elements(image_candidates) e
		WHERE NOT (e->>'provider' = ? AND e->>'provider_id' = ?))`, provider, providerID)
	candidateMatch := `image_candidates @> ?::jsonb`
	candidateJSON := string(mapper.MarshalJSON([]map[string]string{{"provider": provider, "provider_id": providerID}}, "[]"))

	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range []string{"lexemes", "generation_job_items", "cards"} {
			if err := tx.Table(table).Where(match, provider, providerID).
				Updates(map[string]any{"content": setImage, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}
		for _, table := range []string{"lexemes", "generation_job_items"} {
			if err := tx.Table(table).Where(candidateMatch, candidateJSON).
				Update("image_candidates", dropCandidate).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ---- words by topic (Epic 4) ------------------------------------------------

func (r *VocabRepository) KnownWords(ctx context.Context, userID string) ([]string, error) {
	var out []string
	err := r.conn(ctx).Raw("SELECT lemma FROM user_known_words WHERE user_id = ? ORDER BY created_at DESC", userID).Scan(&out).Error
	return out, err
}

func (r *VocabRepository) AddKnownWord(ctx context.Context, userID, lemma string) error {
	return r.conn(ctx).Exec(`INSERT INTO user_known_words (user_id, lemma) VALUES (?, ?)
		ON CONFLICT (user_id, lemma) DO NOTHING`, userID, lemma).Error
}

func (r *VocabRepository) RemoveKnownWord(ctx context.Context, userID, lemma string) error {
	return r.conn(ctx).Exec("DELETE FROM user_known_words WHERE user_id = ? AND lemma = ?", userID, lemma).Error
}

func (r *VocabRepository) CreateSuggestion(ctx context.Context, s *entity.Suggestion) (*entity.Suggestion, error) {
	row := &models.TopicSuggestion{
		UserID: s.UserID, DeckID: s.DeckID, TopicID: s.TopicID, TopicText: s.TopicText,
		Requested: s.Requested, Words: mapper.MarshalJSON(s.Words, "[]"), SkippedExisting: s.SkippedExisting,
	}
	if err := r.conn(ctx).Omit("ID").Create(row).Error; err != nil {
		return nil, err
	}
	return r.mapper.SuggestionModelToEntity(row), nil
}

func (r *VocabRepository) FindSuggestion(ctx context.Context, userID, id string) (*entity.Suggestion, error) {
	var row models.TopicSuggestion
	err := r.conn(ctx).First(&row, "id = ? AND user_id = ?", id, userID).Error
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapper.SuggestionModelToEntity(&row), nil
}

func (r *VocabRepository) UpdateSuggestionWords(ctx context.Context, s *entity.Suggestion) error {
	return r.conn(ctx).Model(&models.TopicSuggestion{}).Where("id = ?", s.ID).Updates(map[string]any{
		"words": mapper.MarshalJSON(s.Words, "[]"), "skipped_existing": s.SkippedExisting, "updated_at": time.Now(),
	}).Error
}

func (r *VocabRepository) LogSuggestionRequest(ctx context.Context, userID, suggestionID string, wordsAdded int, u entity.Usage) error {
	return r.conn(ctx).Exec(`INSERT INTO topic_suggestion_requests
		(suggestion_id, user_id, words_added, input_tokens, output_tokens) VALUES (?, ?, ?, ?, ?)`,
		suggestionID, userID, wordsAdded, u.InputTokens, u.OutputTokens).Error
}

func (r *VocabRepository) SuggestionRequestsSince(ctx context.Context, userID string, since time.Time) (int, error) {
	var n int64
	err := r.conn(ctx).Table("topic_suggestion_requests").Where("user_id = ? AND created_at >= ?", userID, since).Count(&n).Error
	return int(n), err
}
