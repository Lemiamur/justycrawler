package domain

type CrawledData struct {
	URL        string   `bson:"url"`
	Depth      int      `bson:"depth"`
	FoundOn    string   `bson:"found_on"` // URL, на котором была найдена эта страница
	FoundLinks []string `bson:"found_links"`
}
