package api_utils

import (
	dataaccess "app/internal/data_access"
	"encoding/json"
	"net/http"
	"strings"
)

type ApiResponseService struct {
	ResponseWriter http.ResponseWriter
}

func NewApiResponseService(w http.ResponseWriter) *ApiResponseService {
	return &ApiResponseService{
		ResponseWriter: w,
	}
}

func (s *ApiResponseService) Response(statusCode int, d *dataaccess.ApiResponseDao) {
	s.ResponseWriter.Header().Set("Content-Type", "application/json")
	s.ResponseWriter.WriteHeader(statusCode)
	json.NewEncoder(s.ResponseWriter).Encode(d)
}

func (s *ApiResponseService) Abort(statusCode int, title, message string) {
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(statusCode)
	// json.NewEncoder(w).Encode(dataaccess.ApiResponseDao{
	// 	Code:   statusCode,
	// 	Status: title,
	// 	Payload: struct {
	// 		Message string `json:"message"`
	// 	}{
	// 		Message: message,
	// 	},
	// })

	d := &dataaccess.ApiResponseDao{
		Code:   statusCode,
		Status: title,
		Payload: struct {
			Message string
		}{
			Message: message,
		},
	}

	s.Response(statusCode, d)
}

func (s *ApiResponseService) IsRequestApi(r *http.Request) bool {
	pathList := strings.Split(r.URL.Path, "/")
	if len(pathList) > 1 && strings.ToLower(pathList[1]) == "api" {
		return true
	}
	return false
}
