package gitiles

type LogResponse struct {
	Log      []Commit `json:"log"`
	Previous string   `json:"previous"`
	Next     string   `json:"next"`
}

type Commit struct {
	Commit    string   `json:"commit"`
	Tree      string   `json:"tree"`
	Parents   []string `json:"parents"`
	Author    Ident    `json:"author"`
	Committer Ident    `json:"committer"`
	Message   string   `json:"message"`
}

type Ident struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Time  string `json:"time"`
}
