package entity

// Value objects: Notification, Field (no identity; sent through notify/ports/in.IAdminNotifier).

// Field is one labelled line, e.g. {"From", "Learner One (learner@example.com)"}.
type Field struct {
	Label string
	Value string
}

// Notification is a channel-neutral message for the app owner. Adapters format it
// (Telegram: HTML with the user text escaped). All text is plain; adapters escape it.
type Notification struct {
	Emoji  string // optional, shown before the title
	Title  string
	Fields []Field
	Body   string // free text from a user; adapters may cut it
	Link   string // optional URL shown last as "Open: <link>"
}
