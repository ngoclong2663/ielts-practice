-- Epic 9: the card generator tags each card with its IELTS topic (an Epic 4 topic id, or 'other' for general
-- language). Library candidates use it to preselect the topic; '' for cards made before this.
ALTER TABLE lexemes ADD COLUMN topic VARCHAR(32) NOT NULL DEFAULT '';
