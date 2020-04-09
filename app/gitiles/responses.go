package gitiles

type LogResponse struct {
	Log      []Commit `json:"log"`
	Previous string   `json:"previous"`
	Next     string   `json:"next"`
}

type RevisionResponse struct {
	Commit
	TreeDiff []Diff `json:"tree_diff"`
}

type Commit struct {
	SHA       string   `json:"commit"`
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

type Diff struct {
	Type    string `json:"type"`
	OldID   string `json:"old_id"`
	OldMode int    `json:"old_mode"`
	OldPath string `json:"old_path"`
	NewID   string `json:"new_id"`
	NewMode int    `json:"new_mode"`
	NewPath string `json:"new_path"`
}
