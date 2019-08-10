package wapi

type DummyURLsProvider []string

func (p DummyURLsProvider) URLs() ([]string, error) {
	return p, nil
}
