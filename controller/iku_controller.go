package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type IkuController interface {
	FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindAllIkuOpd(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	GetByTahun(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}
