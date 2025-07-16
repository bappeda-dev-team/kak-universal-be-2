package controller

import (
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/web"
	"ekak_kabupaten_madiun/model/web/sasaranpemda"
	"ekak_kabupaten_madiun/service"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type SasaranPemdaControllerImpl struct {
	sasaranPemdaService service.SasaranPemdaService
}

func NewSasaranPemdaControllerImpl(sasaranPemdaService service.SasaranPemdaService) *SasaranPemdaControllerImpl {
	return &SasaranPemdaControllerImpl{
		sasaranPemdaService: sasaranPemdaService,
	}
}

func (controller *SasaranPemdaControllerImpl) Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// Decode request body
	sasaranPemdaCreateRequest := sasaranpemda.SasaranPemdaCreateRequest{}
	helper.ReadFromRequestBody(request, &sasaranPemdaCreateRequest)

	// Panggil service create
	sasaranPemdaResponse, err := controller.sasaranPemdaService.Create(request.Context(), sasaranPemdaCreateRequest)
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
		Status: "success create sasaran pemda",
		Data:   sasaranPemdaResponse,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranPemdaControllerImpl) Update(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	sasaranPemdaUpdateRequest := sasaranpemda.SasaranPemdaUpdateRequest{}
	helper.ReadFromRequestBody(request, &sasaranPemdaUpdateRequest)

	id := params.ByName("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   "Invalid ID format",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}
	sasaranPemdaUpdateRequest.Id = idInt

	// Panggil service update
	sasaranPemdaResponse, err := controller.sasaranPemdaService.Update(request.Context(), sasaranPemdaUpdateRequest)
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
		Status: "success update sasaran pemda",
		Data:   sasaranPemdaResponse,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranPemdaControllerImpl) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	sasaranPemdaId := params.ByName("id")
	id, err := strconv.Atoi(sasaranPemdaId)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   "Invalid ID format",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	err = controller.sasaranPemdaService.Delete(request.Context(), id)
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
		Status: "success delete sasaran pemda",
		Data:   nil,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranPemdaControllerImpl) FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	sasaranPemdaId := params.ByName("id")

	id, err := strconv.Atoi(sasaranPemdaId)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   "Invalid ID format",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	sasaranPemdaResponse, err := controller.sasaranPemdaService.FindById(request.Context(), id)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "NOT FOUND",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusOK,
		Status: "success get sasaran pemda by id",
		Data:   sasaranPemdaResponse,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranPemdaControllerImpl) FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	tahun := params.ByName("tahun")
	sasaranPemdaResponses, err := controller.sasaranPemdaService.FindAll(request.Context(), tahun)
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
		Status: "success get sasaran pemda",
		Data:   sasaranPemdaResponses,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranPemdaControllerImpl) FindByTahun(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	tahun := params.ByName("tahun")
	sasaranPemdaResponses, err := controller.sasaranPemdaService.FindByTahun(r.Context(), tahun)
	if err != nil {
		webResp := web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(w, webResp)
		return
	}

	webResp := web.WebResponse{
		Code:   http.StatusOK,
		Status: "success",
		Data:   sasaranPemdaResponses,
	}
	helper.WriteToResponseBody(w, webResp)
}

func (controller *SasaranPemdaControllerImpl) FindAllWithPokin(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	tahunAwal := params.ByName("tahun_awal")
	tahunAkhir := params.ByName("tahun_akhir")
	jenisPeriode := params.ByName("jenis_periode")

	sasaranPemdaResponses, err := controller.sasaranPemdaService.FindAllWithPokin(request.Context(), tahunAwal, tahunAkhir, jenisPeriode)
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
		Status: "success get sasaran pemda with pokin",
		Data:   sasaranPemdaResponses,
	}
	helper.WriteToResponseBody(writer, webResponse)
}
