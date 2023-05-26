package utils_test

import (
	"testing"

	"github.com/paulori22/encoder-microservice/framework/utils"
	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	json := `{
				"id":"fc47a813-4bca-4f01-bee1-7063b942f36e",
				"file_path":"file.mp4",
				"status":"PENDING"

		     }`
	err := utils.IsJson(json)
	require.Nil(t, err)

	json = `invalid_json`
	err = utils.IsJson(json)
	require.Error(t, err)

}
