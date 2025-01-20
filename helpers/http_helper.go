package helpers

import "net/http"

func HumanlikeGet(url string) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	return client.Do(req)
}
