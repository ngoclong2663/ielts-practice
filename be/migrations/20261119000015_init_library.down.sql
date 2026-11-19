DROP TABLE IF EXISTS library_words;
DROP TABLE IF EXISTS library_topics;
DROP TABLE IF EXISTS library_collections;
DELETE FROM lexemes WHERE origin = 'library' AND id NOT IN (SELECT lexeme_id FROM cards WHERE lexeme_id IS NOT NULL);
ALTER TABLE lexemes DROP COLUMN IF EXISTS origin;
