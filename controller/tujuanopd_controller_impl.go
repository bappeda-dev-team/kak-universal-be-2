package controller

import (
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/web"
	"ekak_kabupaten_madiun/model/web/tujuanopd"
	"ekak_kabupaten_madiun/service"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type TujuanOpdControllerImpl struct {
	TujuanOpdService service.TujuanOpdService
}

func NewTujuanOpdControllerImpl(tujuanOpdService service.TujuanOpdService) *TujuanOpdControllerImpl {
	return &TujuanOpdControllerImpl{
		TujuanOpdService: tujuanOpdService,
	}
}

func (controller *TujuanOpdControllerImpl) Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	tujuanOpdCreateRequest := tujuanopd.TujuanOpdCreateRequest{}
	helper.ReadFromRequestBody(request, &tujuanOpdCreateRequest)

	// Panggil service Create
	tujuanOpdResponse, err := controller.TujuanOpdService.Create(request.Context(), tujuanOpdCreateRequest)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusCreated,
		Status: "success create tujuan opd",
		Data:   tujuanOpdResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *TujuanOpdControllerImpl) Update(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// Ambil ID dari params
	tujuanOpdId := params.ByName("tujuanOpdId")
	tujuanOpdIdInt, err := strconv.Atoi(tujuanOpdId)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
	}

	// Baca request body
	tujuanOpdUpdateRequest := tujuanopd.TujuanOpdUpdateRequest{}
	helper.ReadFromRequestBody(request, &tujuanOpdUpdateRequest)

	// Set ID dari params ke request
	tujuanOpdUpdateRequest.Id = tujuanOpdIdInt

	// Panggil service Update
	tujuanOpdResponse, err := controller.TujuanOpdService.Update(request.Context(), tujuanOpdUpdateRequest)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusOK,
		Status: "OK",
		Data:   tujuanOpdResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *TujuanOpdControllerImpl) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	tujuanOpdId := params.ByName("tujuanOpdId")
	tujuanOpdIdInt, err := strconv.Atoi(tujuanOpdId)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	err = controller.TujuanOpdService.Delete(request.Context(), tujuanOpdIdInt)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusOK,
		Status: "OK",
		Data:   "Data berhasil dihapus",
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *TujuanOpdControllerImpl) FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	tujuanOpdId := params.ByName("tujuanOpdId")
	tujuanOpdIdInt, err := strconv.Atoi(tujuanOpdId)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	tujuanOpdResponse, err := controller.TujuanOpdService.FindById(request.Context(), tujuanOpdIdInt)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusOK,
		Status: "success find by id tujuan opd",
		Data:   tujuanOpdResponse,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *TujuanOpdControllerImpl) FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	kodeOpd := params.ByName("kode_opd")
	tahunAwal := params.ByName("tahun_awal")
	tahunAkhir := params.ByName("tahun_akhir")
	jenisPeriode := params.ByName("jenis_periode")

	tujuanOpdResponses, err := controller.TujuanOpdService.FindAll(request.Context(), kodeOpd, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusOK,
		Status: "success find all tujuan opd",
		Data:   tujuanOpdResponses,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *TujuanOpdControllerImpl) GetByTahun(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	tahun := params.ByName("tahun")
	kodeOpd := params.ByName("kode_opd")

	tujuanOpdResponses, err := controller.TujuanOpdService.GetByTahun(request.Context(), tahun, kodeOpd)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusOK,
		Status: "success find all tujuan opd",
		Data:   tujuanOpdResponses,
	}
	helper.WriteToResponseBody(writer, webResponse)
}
