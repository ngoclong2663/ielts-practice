package rules

// Topic is a preset IELTS theme (Epic 4 §4). Labels are translated by the web app (vocab.topics.<id>);
// the description is what the model sees.
type Topic struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

var Topics = []Topic{
	{"health", "health and hospitals: illness, treatment, hospitals, healthcare systems, public health"},
	{"money", "money and finance: banking, saving, spending, the economy, taxes"},
	{"environment", "the environment: pollution, climate change, energy, conservation, waste"},
	{"education", "education: schools, universities, learning, teaching, exams"},
	{"technology", "technology: computers, the internet, social media, artificial intelligence, innovation"},
	{"work", "work and careers: jobs, workplaces, employment, management, skills"},
	{"crime", "crime and law: crime, punishment, police, courts, justice"},
	{"travel", "travel and tourism: holidays, transport, tourism, culture abroad"},
	{"food", "food: diet, nutrition, cooking, eating habits, agriculture"},
	{"media", "media and advertising: news, television, newspapers, advertising, marketing"},
	{"cities", "housing and cities: housing, urbanisation, infrastructure, city life"},
	{"society", "family and society: family, generations, ageing, social issues, community"},
	{"sport", "sport: exercise, fitness, competitions, sports events"},
	{"science", "science: research, discoveries, space, experiments"},
}

func FindTopic(id string) (Topic, bool) {
	for _, t := range Topics {
		if t.ID == id {
			return t, true
		}
	}
	return Topic{}, false
}

// TopicOther tags general language that belongs to no preset topic ("however", "significant").
const TopicOther = "other"

// TopicIDs lists the preset topic ids, then "other": the topics a generated card may be tagged with (Epic 9).
func TopicIDs() []string {
	out := make([]string, 0, len(Topics)+1)
	for _, t := range Topics {
		out = append(out, t.ID)
	}
	return append(out, TopicOther)
}
