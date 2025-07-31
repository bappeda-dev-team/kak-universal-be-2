package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type SasaranOpdController interface {
	FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	Update(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindByIdPokin(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindIdPokinSasaran(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	GetByTahun(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
}
