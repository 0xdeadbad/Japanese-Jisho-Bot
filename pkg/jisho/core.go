package jisho

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func JishoSearch(payload string) (*Result, error) {
	result := Result{}

	response, err := http.Get(fmt.Sprintf("https://jisho.org/api/v1/search/words?keyword=%s", payload))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
