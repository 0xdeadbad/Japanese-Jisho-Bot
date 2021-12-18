package jisho

type Source struct {
	Language string `json:"language"`
	Word     string `json:"word"`
}

type Link struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

type Sense struct {
	EnglishDefinition []string `json:"english_definitions"`
	PartsOfSpeech     []string `json:"parts_of_speech"`
	Links             []Link   `json:"links"`
	Tags              []string `json:"tags"`
	Restrictions      []string `json:"restrictions"`
	SeeAlso           []string `json:"see_also"`
	Antonyms          []string `json:"antonyms"`
	Source            []Source `json:"source"`
	Info              []string `json:"info"`
}

type WordReading struct {
	Word    string `json:"word"`
	Reading string `json:"reading"`
}

type Data struct {
	Slug     string        `json:"slug"`
	IsCommon bool          `json:"is_common"`
	Tags     []string      `json:"tags"`
	Jlpt     []string      `json:"jlpt"`
	Japanese []WordReading `json:"japanese"`
	Senses   []Sense       `json:"senses"`
}

type Meta struct {
	Status uint16 `json:"status"`
}

type Result struct {
	Meta Meta   `json:"meta"`
	Data []Data `json:"data"`
}
