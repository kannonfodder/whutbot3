package api

type FileToSend struct{
	Name string
	URL  string
}

type MediaSearcher interface {
	Search(tags []string) (file FileToSend, err error)
}