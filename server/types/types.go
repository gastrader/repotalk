package types

type CrawlRequest struct {
	GithubURL string `json:"githubUrl"`
}

type CrawlResponse struct {
	Message string `json:"message"`
	URL     string `json:"url"`
	Username string `json:"username"`
	Reponame string `json:"reponame"`
	ThreadID string `json:"threadID"`
	FileID string `json:"fileID"`
	Response string `json:"response"`
}

type QueryResponse struct {
	Message string `json:"message"`
	Username string `json:"username"`
	Reponame string `json:"reponame"`
	ThreadID string `json:"threadID"`
	Response string `json:"response"`
}

type AsstConfig struct {
	Name        string
	Model       string
	FileBundles []FileBundle
}

type FileBundle struct {
	SrcDir     string  
	SrcGlobs   []string 
	BundleName string  
	DstExt     string   
}

type AsstID string

type ThreadID string

type ThreadRequest struct {
	ThreadID string `json:"tid"`
	Question  string `json:"question"`
	GithubUser string `json:"githubUser"`
	RepoName string `json:"repoName"`
}