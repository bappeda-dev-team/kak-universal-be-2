package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"ekak_kabupaten_madiun/model/web/pohonkinerja"
	"ekak_kabupaten_madiun/model/web/sasaranopd"
	"ekak_kabupaten_madiun/repository"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SasaranOpdServiceImpl struct {
	sasaranOpdRepository      repository.SasaranOpdRepository
	opdRepository             repository.OpdRepository
	rencanaKinerjaRepository  repository.RencanaKinerjaRepository
	manualIndikatorRepository repository.ManualIKRepository
	pegawaiRepository         repository.PegawaiRepository
	pohonkinerjaRepository    repository.PohonKinerjaRepository
	DB                        *sql.DB
	validate                  *validator.Validate
}

func NewSasaranOpdServiceImpl(
	sasaranOpdRepository repository.SasaranOpdRepository,
	opdRepository repository.OpdRepository,
	rencanaKinerjaRepository repository.RencanaKinerjaRepository,
	manualIndikatorRepository repository.ManualIKRepository,
	pegawaiRepository repository.PegawaiRepository,
	pohonkinerjaRepository repository.PohonKinerjaRepository,
	db *sql.DB,
	validate *validator.Validate,
) *SasaranOpdServiceImpl {
	return &SasaranOpdServiceImpl{
		sasaranOpdRepository:      sasaranOpdRepository,
		opdRepository:             opdRepository,
		rencanaKinerjaRepository:  rencanaKinerjaRepository,
		manualIndikatorRepository: manualIndikatorRepository,
		pegawaiRepository:         pegawaiRepository,
		pohonkinerjaRepository:    pohonkinerjaRepository,
		DB:                        db,
		validate:                  validate,
	}
}

func (service *SasaranOpdServiceImpl) FindAll(ctx context.Context, KodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]sasaranopd.SasaranOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	sasaranOpds, err := service.sasaranOpdRepository.FindAll(ctx, tx, KodeOpd, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		return nil, err
	}

	// Sort sasaranOpds berdasarkan nama_pohon, jika sama berdasarkan id ASC
	sort.Slice(sasaranOpds, func(i, j int) bool {
		if sasaranOpds[i].NamaPohon == sasaranOpds[j].NamaPohon {
			return sasaranOpds[i].Id < sasaranOpds[j].Id
		}
		return sasaranOpds[i].NamaPohon < sasaranOpds[j].NamaPohon
	})

	var responses []sasaranopd.SasaranOpdResponse
	for _, sasaranOpd := range sasaranOpds {
		response := sasaranopd.SasaranOpdResponse{
			IdPohon:    sasaranOpd.IdPohon,
			NamaPohon:  sasaranOpd.NamaPohon,
			JenisPohon: sasaranOpd.JenisPohon,
			LevelPohon: sasaranOpd.LevelPohon,
			TahunPohon: sasaranOpd.TahunPohon,
			Pelaksana:  make([]sasaranopd.PelaksanaOpdResponse, 0),
			SasaranOpd: make([]sasaranopd.SasaranOpdDetailResponse, 0),
		}

		// Convert Pelaksana
		for _, pelaksana := range sasaranOpd.Pelaksana {
			response.Pelaksana = append(response.Pelaksana, sasaranopd.PelaksanaOpdResponse{
				Id:          pelaksana.Id,
				PegawaiId:   pelaksana.PegawaiId,
				Nip:         pelaksana.Nip,
				NamaPegawai: pelaksana.NamaPegawai,
			})
		}

		// Temporary slice untuk sorting sasaran
		tempSasaranResponses := make([]sasaranopd.SasaranOpdDetailResponse, 0)

		// Convert SasaranOpd
		for _, sasaran := range sasaranOpd.SasaranOpd {
			sasaranResponse := sasaranopd.SasaranOpdDetailResponse{
				Id:             strconv.Itoa(sasaran.Id),
				NamaSasaranOpd: sasaran.NamaSasaranOpd,
				TahunAwal:      sasaran.TahunAwal,
				TahunAkhir:     sasaran.TahunAkhir,
				JenisPeriode:   sasaran.JenisPeriode,
				Indikator:      make([]sasaranopd.IndikatorResponse, 0),
			}

			// Convert Indikator
			for _, indikator := range sasaran.Indikator {
				indResponse := sasaranopd.IndikatorResponse{
					Id:               indikator.Id,
					Indikator:        indikator.Indikator,
					RumusPerhitungan: indikator.RumusPerhitungan.String,
					SumberData:       indikator.SumberData.String,
					Target:           make([]sasaranopd.TargetResponse, 0),
				}

				// Convert Target
				for _, target := range indikator.Target {
					indResponse.Target = append(indResponse.Target, sasaranopd.TargetResponse{
						Id:     target.Id,
						Tahun:  target.Tahun,
						Target: target.Target,
						Satuan: target.Satuan,
					})
				}

				sasaranResponse.Indikator = append(sasaranResponse.Indikator, indResponse)
			}

			tempSasaranResponses = append(tempSasaranResponses, sasaranResponse)
		}

		// Sort sasaran berdasarkan nama_sasaran_opd
		sort.Slice(tempSasaranResponses, func(i, j int) bool {
			return tempSasaranResponses[i].NamaSasaranOpd < tempSasaranResponses[j].NamaSasaranOpd
		})

		// Assign sorted sasaran ke response
		response.SasaranOpd = tempSasaranResponses

		responses = append(responses, response)
	}

	return responses, nil
}

func (service *SasaranOpdServiceImpl) FindById(ctx context.Context, id int) (*sasaranopd.SasaranOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	sasaranOpd, err := service.sasaranOpdRepository.FindById(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	response := &sasaranopd.SasaranOpdResponse{
		IdPohon:    sasaranOpd.IdPohon,
		NamaPohon:  sasaranOpd.NamaPohon,
		JenisPohon: sasaranOpd.JenisPohon,
		LevelPohon: sasaranOpd.LevelPohon,
		TahunPohon: sasaranOpd.TahunPohon,
		Pelaksana:  make([]sasaranopd.PelaksanaOpdResponse, 0),
		SasaranOpd: make([]sasaranopd.SasaranOpdDetailResponse, 0),
	}

	// Convert Pelaksana
	for _, pelaksana := range sasaranOpd.Pelaksana {
		response.Pelaksana = append(response.Pelaksana, sasaranopd.PelaksanaOpdResponse{
			Id:          pelaksana.Id,
			PegawaiId:   pelaksana.PegawaiId,
			Nip:         pelaksana.Nip,
			NamaPegawai: pelaksana.NamaPegawai,
		})
	}

	// Convert SasaranOpd
	for _, sasaran := range sasaranOpd.SasaranOpd {
		sasaranResponse := sasaranopd.SasaranOpdDetailResponse{
			Id:             strconv.Itoa(sasaran.Id),
			NamaSasaranOpd: sasaran.NamaSasaranOpd,
			TahunAwal:      sasaran.TahunAwal,
			TahunAkhir:     sasaran.TahunAkhir,
			JenisPeriode:   sasaran.JenisPeriode,
			Indikator:      make([]sasaranopd.IndikatorResponse, 0),
		}

		// Convert Indikator
		for _, indikator := range sasaran.Indikator {
			indResponse := sasaranopd.IndikatorResponse{
				Id:               indikator.Id,
				Indikator:        indikator.Indikator,
				RumusPerhitungan: indikator.RumusPerhitungan.String,
				SumberData:       indikator.SumberData.String,
				Target:           make([]sasaranopd.TargetResponse, 0),
			}

			// Convert Target
			for _, target := range indikator.Target {
				indResponse.Target = append(indResponse.Target, sasaranopd.TargetResponse{
					Id:     target.Id,
					Tahun:  target.Tahun,
					Target: target.Target,
					Satuan: target.Satuan,
				})
			}

			sasaranResponse.Indikator = append(sasaranResponse.Indikator, indResponse)
		}

		response.SasaranOpd = append(response.SasaranOpd, sasaranResponse)
	}

	return response, nil
}

func (service *SasaranOpdServiceImpl) Create(ctx context.Context, request sasaranopd.SasaranOpdCreateRequest) (*sasaranopd.SasaranOpdCreateResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	if err := service.validate.Struct(request); err != nil {
		return nil, err
	}

	_, err = service.pohonkinerjaRepository.FindById(ctx, tx, request.IdPohon)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("data pohon kinerja dengan ID %d tidak ditemukan", request.IdPohon)
		}
		return nil, fmt.Errorf("gagal mengambil data pohon kinerja: %v", err)
	}
	// Generate ID yang lebih unik untuk sasaran OPD
	idSasaran := rand.Intn(100000)

	sasaranOpd := domain.SasaranOpdDetail{
		Id:             idSasaran,
		IdPohon:        request.IdPohon,
		NamaSasaranOpd: request.NamaSasaran,
		TahunAwal:      request.TahunAwal,
		TahunAkhir:     request.TahunAkhir,
		JenisPeriode:   request.JenisPeriode,
		Indikator:      make([]domain.Indikator, 0),
	}

	// Proses indikator
	for _, indReq := range request.Indikator {
		indikatorId := fmt.Sprintf("IND-SAS-%d", uuid.New().ID()%100000)

		indikator := domain.Indikator{
			Id:               indikatorId,
			Indikator:        indReq.Indikator,
			RumusPerhitungan: sql.NullString{String: indReq.RumusPerhitungan, Valid: true},
			SumberData:       sql.NullString{String: indReq.SumberData, Valid: true},
			Target:           make([]domain.Target, 0),
		}

		// Proses target
		for _, targetReq := range indReq.Target {
			if targetReq.Target != "" {
				targetId := fmt.Sprintf("TRG-SAS-%d-%s", uuid.New().ID()%100000, targetReq.Tahun)

				target := domain.Target{
					Id:          targetId,
					IndikatorId: indikator.Id,
					Tahun:       targetReq.Tahun,
					Target:      targetReq.Target,
					Satuan:      targetReq.Satuan,
				}
				indikator.Target = append(indikator.Target, target)
			}
		}

		sasaranOpd.Indikator = append(sasaranOpd.Indikator, indikator)
	}

	err = service.sasaranOpdRepository.Create(ctx, tx, sasaranOpd)
	if err != nil {
		return nil, err
	}

	// Buat response dengan indikator dan target
	response := &sasaranopd.SasaranOpdCreateResponse{
		IdPohon:        sasaranOpd.IdPohon,
		NamaSasaranOpd: sasaranOpd.NamaSasaranOpd,
		TahunAwal:      sasaranOpd.TahunAwal,
		TahunAkhir:     sasaranOpd.TahunAkhir,
		JenisPeriode:   sasaranOpd.JenisPeriode,
		Indikator:      make([]sasaranopd.IndikatorDetail, 0),
	}

	// Convert indikator untuk response
	for _, indikator := range sasaranOpd.Indikator {
		indResponse := sasaranopd.IndikatorDetail{
			Id:               indikator.Id,
			Indikator:        indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String,
			SumberData:       indikator.SumberData.String,
			Target:           make([]sasaranopd.TargetDetail, 0),
		}

		// Convert target untuk response
		for _, target := range indikator.Target {
			indResponse.Target = append(indResponse.Target, sasaranopd.TargetDetail{
				Id:     target.Id,
				Tahun:  target.Tahun,
				Target: target.Target,
				Satuan: target.Satuan,
			})
		}

		response.Indikator = append(response.Indikator, indResponse)
	}

	return response, nil
}

func (service *SasaranOpdServiceImpl) Update(ctx context.Context, request sasaranopd.SasaranOpdUpdateRequest) (*sasaranopd.SasaranOpdCreateResponse, error) {
	// Validasi awal tetap sama
	err := service.validate.Struct(request)
	if err != nil {
		return nil, err
	}

	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi periode tetap sama
	tahunAwalInt, err := strconv.Atoi(request.TahunAwal)
	if err != nil {
		return nil, errors.New("format tahun awal tidak valid")
	}
	tahunAkhirInt, err := strconv.Atoi(request.TahunAkhir)
	if err != nil {
		return nil, errors.New("format tahun akhir tidak valid")
	}
	if tahunAkhirInt < tahunAwalInt {
		return nil, errors.New("tahun akhir tidak boleh lebih kecil dari tahun awal")
	}

	// Cek apakah data sasaran OPD ada
	existingSasaran, err := service.sasaranOpdRepository.FindByIdSasaran(ctx, tx, request.IdSasaranOpd)
	if err != nil {
		return nil, errors.New("sasaran opd tidak ditemukan")
	}

	// Persiapkan data indikator
	var indikatorList []domain.Indikator
	var indikatorResponses []sasaranopd.IndikatorDetail

	for _, indikatorReq := range request.Indikator {
		var indikatorId string

		// Cek apakah indikator sudah ada
		if indikatorReq.Id != "" {
			// Gunakan ID yang sudah ada
			indikatorId = indikatorReq.Id
		} else {
			// Generate ID baru untuk indikator baru
			indikatorId = fmt.Sprintf("IND-SAS-%s", uuid.New().String()[:4])
		}

		var targetList []domain.Target
		var targetResponses []sasaranopd.TargetDetail

		for _, targetReq := range indikatorReq.Target {
			var targetId string

			// Cek apakah target sudah ada
			if targetReq.Id != "" {
				// Gunakan ID yang sudah ada
				targetId = targetReq.Id
			} else {
				// Generate ID baru untuk target baru
				targetId = fmt.Sprintf("TRG-SAS-%d-%s-%s", request.IdSasaranOpd, indikatorId, targetReq.Tahun)
			}

			target := domain.Target{
				Id:          targetId,
				IndikatorId: indikatorId,
				Tahun:       targetReq.Tahun,
				Target:      targetReq.Target,
				Satuan:      targetReq.Satuan,
			}
			targetList = append(targetList, target)

			targetResponses = append(targetResponses, sasaranopd.TargetDetail{
				Id:     targetId,
				Tahun:  targetReq.Tahun,
				Target: targetReq.Target,
				Satuan: targetReq.Satuan,
			})
		}

		indikator := domain.Indikator{
			Id:           indikatorId,
			SasaranOpdId: request.IdSasaranOpd,
			Indikator:    indikatorReq.Indikator,
			RumusPerhitungan: sql.NullString{
				String: indikatorReq.RumusPerhitungan,
				Valid:  true,
			},
			SumberData: sql.NullString{
				String: indikatorReq.SumberData,
				Valid:  true,
			},
			Target: targetList,
		}
		indikatorList = append(indikatorList, indikator)

		indikatorResponses = append(indikatorResponses, sasaranopd.IndikatorDetail{
			Id:               indikatorId,
			Indikator:        indikatorReq.Indikator,
			RumusPerhitungan: indikatorReq.RumusPerhitungan,
			SumberData:       indikatorReq.SumberData,
			Target:           targetResponses,
		})
	}

	// Persiapkan data update
	sasaranOpdUpdate := domain.SasaranOpdDetail{
		Id:             request.IdSasaranOpd,
		IdPohon:        existingSasaran.IdPohon,
		NamaSasaranOpd: request.NamaSasaran,
		TahunAwal:      request.TahunAwal,
		TahunAkhir:     request.TahunAkhir,
		JenisPeriode:   request.JenisPeriode,
		Indikator:      indikatorList,
	}

	// Lakukan update
	updatedSasaranOpd, err := service.sasaranOpdRepository.Update(ctx, tx, sasaranOpdUpdate)
	if err != nil {
		return nil, err
	}

	// Buat response
	response := &sasaranopd.SasaranOpdCreateResponse{
		IdPohon:        updatedSasaranOpd.IdPohon,
		NamaSasaranOpd: updatedSasaranOpd.NamaSasaranOpd,
		TahunAwal:      updatedSasaranOpd.TahunAwal,
		TahunAkhir:     updatedSasaranOpd.TahunAkhir,
		JenisPeriode:   updatedSasaranOpd.JenisPeriode,
		Indikator:      indikatorResponses,
	}

	return response, nil
}

func (service *SasaranOpdServiceImpl) Delete(ctx context.Context, id string) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	err = service.sasaranOpdRepository.Delete(ctx, tx, id)
	if err != nil {
		return err
	}

	return nil
}

func (service *SasaranOpdServiceImpl) FindByIdPokin(ctx context.Context, idPokin int, tahun string) (*sasaranopd.SasaranOpdResponse, error) {
	// Start transaction
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}

	// Ensure transaction is handled properly
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	sasaranOpd, err := service.sasaranOpdRepository.FindByIdPokin(ctx, tx, idPokin, tahun)
	if err != nil {
		return nil, fmt.Errorf("error finding sasaran opd: %v", err)
	}

	// Commit transaction before converting response
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	response := &sasaranopd.SasaranOpdResponse{
		IdPohon:    sasaranOpd.IdPohon,
		NamaPohon:  sasaranOpd.NamaPohon,
		JenisPohon: sasaranOpd.JenisPohon,
		LevelPohon: sasaranOpd.LevelPohon,
		TahunPohon: sasaranOpd.TahunPohon,
		Pelaksana:  make([]sasaranopd.PelaksanaOpdResponse, 0),
		SasaranOpd: make([]sasaranopd.SasaranOpdDetailResponse, 0),
	}

	// Convert response outside of transaction
	for _, pelaksana := range sasaranOpd.Pelaksana {
		response.Pelaksana = append(response.Pelaksana, sasaranopd.PelaksanaOpdResponse{
			Id:          pelaksana.Id,
			PegawaiId:   pelaksana.PegawaiId,
			Nip:         pelaksana.Nip,
			NamaPegawai: pelaksana.NamaPegawai,
		})
	}

	for _, sasaran := range sasaranOpd.SasaranOpd {
		sasaranResponse := sasaranopd.SasaranOpdDetailResponse{
			Id:             strconv.Itoa(sasaran.Id),
			NamaSasaranOpd: sasaran.NamaSasaranOpd,
			TahunAwal:      sasaran.TahunAwal,
			TahunAkhir:     sasaran.TahunAkhir,
			JenisPeriode:   sasaran.JenisPeriode,
			Indikator:      make([]sasaranopd.IndikatorResponse, 0),
		}

		// Convert dan urutkan indikator
		for _, indikator := range sasaran.Indikator {
			indResponse := sasaranopd.IndikatorResponse{
				Id:               indikator.Id,
				Indikator:        indikator.Indikator,
				RumusPerhitungan: indikator.RumusPerhitungan.String,
				SumberData:       indikator.SumberData.String,
				Target:           make([]sasaranopd.TargetResponse, 0),
			}

			// Convert dan urutkan target
			for _, target := range indikator.Target {
				indResponse.Target = append(indResponse.Target, sasaranopd.TargetResponse{
					Id:     target.Id,
					Tahun:  target.Tahun,
					Target: target.Target,
					Satuan: target.Satuan,
				})
			}

			// Urutkan target berdasarkan tahun
			sort.Slice(indResponse.Target, func(i, j int) bool {
				return indResponse.Target[i].Tahun < indResponse.Target[j].Tahun
			})

			sasaranResponse.Indikator = append(sasaranResponse.Indikator, indResponse)
		}

		// Urutkan indikator berdasarkan nama
		sort.Slice(sasaranResponse.Indikator, func(i, j int) bool {
			return sasaranResponse.Indikator[i].Indikator < sasaranResponse.Indikator[j].Indikator
		})

		response.SasaranOpd = append(response.SasaranOpd, sasaranResponse)
	}

	return response, nil
}

func (service *SasaranOpdServiceImpl) FindIdPokinSasaran(ctx context.Context, id int) (pohonkinerja.PohonKinerjaOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil data pohon kinerja dengan sasaran
	pokin, err := service.sasaranOpdRepository.FindIdPokinSasaran(ctx, tx, id)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}

	// Ambil data OPD jika ada
	var namaOpd string
	if pokin.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
		if err == nil {
			namaOpd = opd.NamaOpd
		}
	}

	// Konversi indikator dan target ke response
	var indikatorResponses []pohonkinerja.IndikatorResponse
	for _, ind := range pokin.Indikator {
		var targetResponses []pohonkinerja.TargetResponse

		// Cari target yang ada di database
		var existingTarget *domain.Target
		for _, t := range ind.Target {
			if t.Id != fmt.Sprintf("TRG-%s-%s", ind.Id, pokin.Tahun) {
				existingTarget = &t
				break
			}
		}

		// Jika ada target di database, gunakan itu
		if existingTarget != nil {
			targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
				Id:              existingTarget.Id,
				IndikatorId:     existingTarget.IndikatorId,
				TargetIndikator: existingTarget.Target,
				SatuanIndikator: existingTarget.Satuan,
				TahunSasaran:    pokin.Tahun,
			})
		} else {
			// Jika tidak ada target di database, gunakan target default
			targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
				Id:              fmt.Sprintf("TRG-%s-%s", ind.Id, pokin.Tahun),
				IndikatorId:     ind.Id,
				TargetIndikator: "-",
				SatuanIndikator: "-",
				TahunSasaran:    pokin.Tahun,
			})
		}

		indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
			Id:            ind.Id,
			IdPokin:       fmt.Sprint(pokin.Id),
			NamaIndikator: ind.Indikator,
			Target:        targetResponses,
		})
	}

	response := pohonkinerja.PohonKinerjaOpdResponse{
		Id:         pokin.Id,
		Parent:     strconv.Itoa(pokin.Parent),
		NamaPohon:  pokin.NamaPohon,
		JenisPohon: pokin.JenisPohon,
		LevelPohon: pokin.LevelPohon,
		KodeOpd:    pokin.KodeOpd,
		NamaOpd:    namaOpd,
		Keterangan: pokin.Keterangan,
		Tahun:      pokin.Tahun,
		Status:     pokin.Status,

		Indikator: indikatorResponses,
	}

	return response, nil
}

func (service *SasaranOpdServiceImpl) GetByTahun(ctx context.Context, tahun string, kodeOpd string) ([]sasaranopd.SasaranOpdTahunanResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return []sasaranopd.SasaranOpdTahunanResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	sasaranOpdList, err := service.sasaranOpdRepository.GetByTahun(ctx, tx, tahun, kodeOpd)
	if err != nil {
		return []sasaranopd.SasaranOpdTahunanResponse{}, err
	}

	// untuk indikator
	sasaranMap := make(map[int]*sasaranopd.SasaranOpdTahunanResponse)

	result := make([]sasaranopd.SasaranOpdTahunanResponse, 0, len(sasaranOpdList))
	for _, sasaranOpd := range sasaranOpdList {
		sasaranopdId := strconv.Itoa(sasaranOpd.Id)
		resp := sasaranopd.SasaranOpdTahunanResponse{
			Id:           sasaranopdId,
			IdPohon:      sasaranOpd.IdPohon,
			KodeOpd:      sasaranOpd.KodeOpd,
			NamaOpd:      sasaranOpd.NamaOpd,
			SasaranOpd:   sasaranOpd.SasaranOpd,
			TahunAwal:    sasaranOpd.TahunAwal,
			TahunAkhir:   sasaranOpd.TahunAkhir,
			JenisPeriode: sasaranOpd.JenisPeriode,
			JenisPohon:   sasaranOpd.JenisPohon,
			PohonAktif:   sasaranOpd.PohonAktif,
			Indikator:    []sasaranopd.IndikatorResponse{},
		}
		result = append(result, resp)
		sasaranMap[sasaranOpd.Id] = &result[len(result)-1]
	}

	// indikator
	indikatorList, err := service.sasaranOpdRepository.GetIndikatorSasaranOpdByTahun(ctx, tx, tahun, kodeOpd)
	if err != nil {
		return result, nil
	}

	for _, indikator := range indikatorList {
		if sasaranOpd, ok := sasaranMap[indikator.SasaranOpdId]; ok {
			targetResponse := make([]sasaranopd.TargetResponse, 0, len(indikator.Target))
			for _, target := range indikator.Target {
				targetResponse = append(targetResponse, sasaranopd.TargetResponse{
					Id:          target.Id,
					IndikatorId: target.IndikatorId,
					Target:      target.Target,
					Satuan:      target.Satuan,
					Tahun:       target.Tahun,
				})
			}

			sasaranOpd.Indikator = append(sasaranOpd.Indikator, sasaranopd.IndikatorResponse{
				Id:               indikator.Id,
				SasaranOpdId:     strconv.Itoa(indikator.SasaranOpdId),
				Indikator:        indikator.Indikator,
				SumberData:       nullStringToString(indikator.SumberData),
				RumusPerhitungan: nullStringToString(indikator.RumusPerhitungan),
				Target:           targetResponse,
			})
		}
	}
	return result, nil
}
