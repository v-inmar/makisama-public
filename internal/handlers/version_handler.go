package handlers

import (
	dataaccess "app/internal/data_access"
	api_utils "app/internal/utilities/api"
	"app/internal/version"
	"net/http"
)

func (repo *HandlerRepo) GetApiVersion(w http.ResponseWriter, r *http.Request) {
	apiService := api_utils.NewApiResponseService(w)

	payload := dataaccess.ReadVersionResponseDao{
		Version:   version.Version,
		Commit:    version.CommitHex,
		BuildTime: version.BuildTime,
	}

	data := &dataaccess.ApiResponseDao{
		Code:    http.StatusOK,
		Status:  "OK",
		Payload: payload,
	}

	apiService.Response(http.StatusOK, data)
}
