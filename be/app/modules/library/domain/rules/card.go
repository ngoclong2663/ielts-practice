package rules

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"ielts-practice/app/modules/library/domain/entity"
	vocabentity "ielts-practice/app/modules/vocab/domain/entity"
	vocabrules "ielts-practice/app/modules/vocab/domain/rules"
)

var pixabayPathRe = regexp.MustCompile(`^/api/v1/images/pixabay/([0-9]{1,20})$`)

// CheckCard validates a reviewed card (library word or seeded writing vocabulary, Epic 8): complete, and
// passing the same checks as an AI card (no links, markup or instruction-like text; about its headword). It
// returns the cleaned card and image query, and whether the photo is a Pixabay one served through our endpoint.
func CheckCard(c entity.CardContent, imageQuery string, visualizable bool) (entity.CardContent, string, bool, error) {
	c.Headword, c.Lemma = strings.TrimSpace(c.Headword), strings.TrimSpace(c.Lemma)
	if c.Lemma == "" {
		c.Lemma = c.Headword
	}
	switch {
	case c.Headword == "" || c.MeaningVI == "":
		return c, "", false, errors.New("headword and meaning_vi are required")
	case c.IPA == nil || (c.IPA.UK == "" && c.IPA.US == ""):
		return c, "", false, errors.New("ipa is required")
	case !slices.Contains(entity.Levels, c.CEFR):
		return c, "", false, fmt.Errorf("cefr %q", c.CEFR)
	case len(c.WordForms) == 0 || len(c.Collocations) < minCollocations(c.Headword):
		return c, "", false, fmt.Errorf("needs word_forms and at least %d collocations", minCollocations(c.Headword))
	}
	for _, f := range c.WordForms {
		if f.Example == nil || f.Example.EN == "" || f.Example.VI == "" {
			return c, "", false, fmt.Errorf("word form %q needs an example with vi", f.Form)
		}
	}
	pixabay := false
	if img := c.Image; img != nil {
		switch img.Provider {
		case "unsplash":
			if !strings.HasPrefix(img.URL, "https://images.unsplash.com/") {
				return c, "", false, fmt.Errorf("unsplash image url %q", img.URL)
			}
		case "pixabay":
			m := pixabayPathRe.FindStringSubmatch(img.URL)
			if m == nil || m[1] != img.ProviderID {
				return c, "", false, fmt.Errorf("pixabay image must be served from /api/v1/images/pixabay/<id>, got %q", img.URL)
			}
			pixabay = true
		default:
			return c, "", false, fmt.Errorf("image provider %q", img.Provider)
		}
	}
	res := vocabentity.WordResult{Input: c.Headword, IsValid: true, Content: c, ImageQuery: imageQuery, Visualizable: visualizable}
	vocabrules.SanitizeResult(&res)
	if !vocabrules.Shareable(res.QualityFlags) {
		return c, "", false, fmt.Errorf("card failed the content checks: %v", res.QualityFlags)
	}
	return res.Content, res.ImageQuery, pixabay, nil
}

// minCollocations: a single word needs three; a phrase ("account for", "labour market") one.
func minCollocations(headword string) int {
	if strings.Contains(headword, " ") {
		return 1
	}
	return 3
}
