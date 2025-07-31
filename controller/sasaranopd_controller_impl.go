package controller

import (
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/web"
	"ekak_kabupaten_madiun/model/web/sasaranopd"
	"ekak_kabupaten_madiun/service"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type SasaranOpdControllerImpl struct {
	SasaranOpdService service.SasaranOpdService
}

func NewSasaranOpdControllerImpl(SasaranOpdService service.SasaranOpdService) *SasaranOpdControllerImpl {
	return &SasaranOpdControllerImpl{
		SasaranOpdService: SasaranOpdService,
	}
}

func (controller *SasaranOpdControllerImpl) FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	KodeOpd := params.ByName("kode_opd")
	tahunAwal := params.ByName("tahun_awal")
	tahunAkhir := params.ByName("tahun_akhir")
	jenisPeriode := params.ByName("jenis_periode")

	sasaranOpdResponse, err := controller.SasaranOpdService.FindAll(request.Context(), KodeOpd, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD_REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
	} else {
		webResponse := web.WebResponse{
			Code:   200,
			Status: "get all sasaran opd",
			Data:   sasaranOpdResponse,
		}
		helper.WriteToResponseBody(writer, webResponse)
	}
}

func (controller *SasaranOpdControllerImpl) FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	idInt, err := strconv.Atoi(id)
	helper.PanicIfError(err)

	sasaranOpdResponse, err := controller.SasaranOpdService.FindById(request.Context(), idInt)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD_REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
	} else {
		webResponse := web.WebResponse{
			Code:   200,
			Status: "get sasaran opd by id",
			Data:   sasaranOpdResponse,
		}
		helper.WriteToResponseBody(writer, webResponse)
	}
}

func (controller *SasaranOpdControllerImpl) Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	sasaranOpdCreateRequest := sasaranopd.SasaranOpdCreateRequest{}
	helper.ReadFromRequestBody(request, &sasaranOpdCreateRequest)

	sasaranOpdCreateResponse, err := controller.SasaranOpdService.Create(request.Context(), sasaranOpdCreateRequest)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "failed create sasaran opd",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   http.StatusCreated,
		Status: "success create sasaran opd",
		Data:   sasaranOpdCreateResponse,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranOpdControllerImpl) Update(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	sasaranOpdUpdateRequest := sasaranopd.SasaranOpdUpdateRequest{}
	helper.ReadFromRequestBody(request, &sasaranOpdUpdateRequest)

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "Bad Request",
			Data:   "ID harus berupa angka",
		}
		writer.WriteHeader(http.StatusBadRequest)
		helper.WriteToResponseBody(writer, webResponse)
		return
	}
	sasaranOpdUpdateRequest.IdSasaranOpd = id

	// Panggil service Update
	sasaranOpdResponse, err := controller.SasaranOpdService.Update(request.Context(), sasaranOpdUpdateRequest)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	// Kirim response
	webResponse := web.WebResponse{
		Code:   200,
		Status: "Success Update Sasaran Opd",
		Data:   sasaranOpdResponse,
	}

	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranOpdControllerImpl) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	id := params.ByName("id")

	err := controller.SasaranOpdService.Delete(request.Context(), id)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "failed delete sasaran opd",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
	} else {
		webResponse := web.WebResponse{
			Code:   200,
			Status: "success delete sasaran opd",
			Data:   id,
		}
		helper.WriteToResponseBody(writer, webResponse)
	}
}

func (controller *SasaranOpdControllerImpl) FindByIdPokin(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	idPokinStr := params.ByName("id_pokin")
	tahun := params.ByName("tahun")

	// Validasi parameter tidak boleh kosong
	if idPokinStr == "" {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD REQUEST",
			Data:   "ID pohon kinerja tidak boleh kosong",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	if tahun == "" {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD REQUEST",
			Data:   "Tahun tidak boleh kosong",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	// Validasi format tahun
	if len(tahun) != 4 {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD REQUEST",
			Data:   "Format tahun harus 4 digit (YYYY)",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	idPokin, err := strconv.Atoi(idPokinStr)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD REQUEST",
			Data:   "ID pohon kinerja harus berupa angka",
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	sasaranOpdResponse, err := controller.SasaranOpdService.FindByIdPokin(request.Context(), idPokin, tahun)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
		return
	}

	webResponse := web.WebResponse{
		Code:   200,
		Status: "get sasaran opd by id pokin and tahun",
		Data:   sasaranOpdResponse,
	}
	helper.WriteToResponseBody(writer, webResponse)
}

func (controller *SasaranOpdControllerImpl) FindIdPokinSasaran(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	idPokinStr := params.ByName("id")
	idPokin, err := strconv.Atoi(idPokinStr)
	helper.PanicIfError(err)

	sasaranOpdResponse, err := controller.SasaranOpdService.FindIdPokinSasaran(request.Context(), idPokin)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD_REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
	} else {
		webResponse := web.WebResponse{
			Code:   200,
			Status: "get all sasaran opd by id rencana kinerja",
			Data:   sasaranOpdResponse,
		}
		helper.WriteToResponseBody(writer, webResponse)
	}
}

func (controller *SasaranOpdControllerImpl) GetByTahun(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	kodeOpd := params.ByName("kode_opd")
	tahun := params.ByName("tahun")
	// idPokin, err := strconv.Atoi(idPokinStr)
	// helper.PanicIfError(err)

	sasaranOpdResponse, err := controller.SasaranOpdService.GetByTahun(request.Context(), tahun, kodeOpd)
	if err != nil {
		webResponse := web.WebResponse{
			Code:   400,
			Status: "BAD_REQUEST",
			Data:   err.Error(),
		}
		helper.WriteToResponseBody(writer, webResponse)
	} else {
		webResponse := web.WebResponse{
			Code:   200,
			Status: "get all sasaran opd by id rencana kinerja",
			Data:   sasaranOpdResponse,
		}
		helper.WriteToResponseBody(writer, webResponse)
	}
}
