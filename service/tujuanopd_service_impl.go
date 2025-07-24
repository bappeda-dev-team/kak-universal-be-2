package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"ekak_kabupaten_madiun/model/web/tujuanopd"
	"ekak_kabupaten_madiun/repository"
	"fmt"
	"sort"
	"strconv"

	"github.com/google/uuid"
)

type TujuanOpdServiceImpl struct {
	TujuanOpdRepository    repository.TujuanOpdRepository
	OpdRepository          repository.OpdRepository
	PeriodeRepository      repository.PeriodeRepository
	BidangUrusanRepository repository.BidangUrusanRepository
	DB                     *sql.DB
}

func NewTujuanOpdServiceImpl(tujuanOpdRepository repository.TujuanOpdRepository, opdRepository repository.OpdRepository, periodeRepository repository.PeriodeRepository, bidangUrusanRepository repository.BidangUrusanRepository, DB *sql.DB) *TujuanOpdServiceImpl {
	return &TujuanOpdServiceImpl{
		TujuanOpdRepository:    tujuanOpdRepository,
		OpdRepository:          opdRepository,
		PeriodeRepository:      periodeRepository,
		BidangUrusanRepository: bidangUrusanRepository,
		DB:                     DB,
	}
}

func (service *TujuanOpdServiceImpl) Create(ctx context.Context, request tujuanopd.TujuanOpdCreateRequest) (tujuanopd.TujuanOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi periode
	periode, err := service.PeriodeRepository.FindById(ctx, tx, request.PeriodeId)
	if err != nil {
		if err == sql.ErrNoRows {
			return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("periode dengan id %d tidak ditemukan", request.PeriodeId)
		}
		return tujuanopd.TujuanOpdResponse{}, err
	}

	// Convert tahun awal dan akhir ke integer untuk validasi
	tahunAwal, err := strconv.Atoi(periode.TahunAwal)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("format tahun awal periode tidak valid: %s", periode.TahunAwal)
	}

	tahunAkhir, err := strconv.Atoi(periode.TahunAkhir)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("format tahun akhir periode tidak valid: %s", periode.TahunAkhir)
	}

	//validasi bidang urusan
	_, err = service.BidangUrusanRepository.FindByKodeBidangUrusan(ctx, tx, request.KodeBidangUrusan)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	tujuanOpdDomain := domain.TujuanOpd{
		KodeOpd:          request.KodeOpd,
		KodeBidangUrusan: request.KodeBidangUrusan,
		Tujuan:           request.Tujuan,
		PeriodeId: domain.Periode{
			Id:           request.PeriodeId,
			TahunAwal:    periode.TahunAwal,
			TahunAkhir:   periode.TahunAkhir,
			JenisPeriode: periode.JenisPeriode,
		},
		TahunAwal:    periode.TahunAwal,
		TahunAkhir:   periode.TahunAkhir,
		JenisPeriode: periode.JenisPeriode,
	}

	// Convert indikator request to domain
	for _, indikatorReq := range request.Indikator {
		// Generate ID indikator dengan format IND-TJN-XXXXX
		uuidInd := uuid.New().String()[:5]
		indikatorId := fmt.Sprintf("IND-TJN-%s", uuidInd)

		indikatorDomain := domain.Indikator{
			Id:               indikatorId,
			Indikator:        indikatorReq.Indikator,
			RumusPerhitungan: sql.NullString{String: indikatorReq.RumusPerhitungan, Valid: true},
			SumberData:       sql.NullString{String: indikatorReq.SumberData, Valid: true},
		}

		// Map untuk mengecek duplikasi tahun
		tahunMap := make(map[string]bool)

		// Convert target request to domain
		for _, targetReq := range indikatorReq.Target {
			// Validasi format tahun target
			tahunTarget, err := strconv.Atoi(targetReq.Tahun)
			if err != nil {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("format tahun target tidak valid: %s", targetReq.Tahun)
			}

			// Validasi tahun target berada dalam range periode
			if tahunTarget < tahunAwal || tahunTarget > tahunAkhir {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"tahun target %d harus berada dalam rentang periode %d-%d",
					tahunTarget,
					tahunAwal,
					tahunAkhir,
				)
			}

			// Validasi duplikasi tahun
			if tahunMap[targetReq.Tahun] {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"tahun target %s duplikat dalam indikator yang sama",
					targetReq.Tahun,
				)
			}
			tahunMap[targetReq.Tahun] = true

			// Validasi target dan satuan tidak boleh kosong
			if targetReq.Target == "" {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"target untuk tahun %s tidak boleh kosong",
					targetReq.Tahun,
				)
			}
			if targetReq.Satuan == "" {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"satuan untuk tahun %s tidak boleh kosong",
					targetReq.Tahun,
				)
			}

			// Generate ID target dengan format TRG-TJN-XXXXX
			uuidTrg := uuid.New().String()[:5]
			targetId := fmt.Sprintf("TRG-TJN-%s", uuidTrg)

			targetDomain := domain.Target{
				Id:     targetId,
				Target: targetReq.Target,
				Satuan: targetReq.Satuan,
				Tahun:  targetReq.Tahun,
			}
			indikatorDomain.Target = append(indikatorDomain.Target, targetDomain)
		}

		// Validasi minimal harus ada 1 target
		if len(indikatorReq.Target) == 0 {
			return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
				"indikator harus memiliki minimal 1 target dalam rentang periode %d-%d",
				tahunAwal,
				tahunAkhir,
			)
		}

		tujuanOpdDomain.Indikator = append(tujuanOpdDomain.Indikator, indikatorDomain)
	}

	// Panggil repository dan terima hasil domain yang sudah ada ID-nya
	tujuanOpdResult, err := service.TujuanOpdRepository.Create(ctx, tx, tujuanOpdDomain)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	return helper.ToTujuanOpdResponse(tujuanOpdResult), nil
}

func (service *TujuanOpdServiceImpl) Update(ctx context.Context, request tujuanopd.TujuanOpdUpdateRequest) (tujuanopd.TujuanOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi periode
	periode, err := service.PeriodeRepository.FindById(ctx, tx, request.PeriodeId)
	if err != nil {
		if err == sql.ErrNoRows {
			return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("periode dengan id %d tidak ditemukan", request.PeriodeId)
		}
		return tujuanopd.TujuanOpdResponse{}, err
	}

	// Convert tahun awal dan akhir ke integer untuk validasi
	tahunAwal, err := strconv.Atoi(periode.TahunAwal)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("format tahun awal periode tidak valid: %s", periode.TahunAwal)
	}

	tahunAkhir, err := strconv.Atoi(periode.TahunAkhir)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("format tahun akhir periode tidak valid: %s", periode.TahunAkhir)
	}

	// Cek apakah data exists
	_, err = service.TujuanOpdRepository.FindById(ctx, tx, request.Id)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	//validasi bidang urusan
	_, err = service.BidangUrusanRepository.FindByKodeBidangUrusan(ctx, tx, request.KodeBidangUrusan)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	// Update data utama
	tujuanOpd := domain.TujuanOpd{
		Id:               request.Id,
		KodeOpd:          request.KodeOpd,
		KodeBidangUrusan: request.KodeBidangUrusan,
		Tujuan:           request.Tujuan,
		PeriodeId: domain.Periode{
			Id:           request.PeriodeId,
			TahunAwal:    periode.TahunAwal,
			TahunAkhir:   periode.TahunAkhir,
			JenisPeriode: periode.JenisPeriode,
		},
		TahunAwal:    periode.TahunAwal,
		TahunAkhir:   periode.TahunAkhir,
		JenisPeriode: periode.JenisPeriode,
	}

	// Convert indikator request to domain
	for _, indikatorReq := range request.Indikator {
		var indikatorId string

		// Gunakan ID lama jika ada, generate baru jika tidak ada
		if indikatorReq.Id != "" {
			indikatorId = indikatorReq.Id
		} else {
			uuidInd := uuid.New().String()[:5]
			indikatorId = fmt.Sprintf("IND-TJN-%s", uuidInd)
		}

		indikatorDomain := domain.Indikator{
			Id:               indikatorId,
			Indikator:        indikatorReq.Indikator,
			RumusPerhitungan: sql.NullString{String: indikatorReq.RumusPerhitungan, Valid: true},
			SumberData:       sql.NullString{String: indikatorReq.SumberData, Valid: true},
		}

		// Map untuk mengecek duplikasi tahun
		tahunMap := make(map[string]bool)

		// Convert target request to domain
		for _, targetReq := range indikatorReq.Target {
			// Validasi format tahun target
			tahunTarget, err := strconv.Atoi(targetReq.Tahun)
			if err != nil {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf("format tahun target tidak valid: %s", targetReq.Tahun)
			}

			// Validasi tahun target berada dalam range periode
			if tahunTarget < tahunAwal || tahunTarget > tahunAkhir {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"tahun target %d harus berada dalam rentang periode %d-%d",
					tahunTarget,
					tahunAwal,
					tahunAkhir,
				)
			}

			// Validasi duplikasi tahun
			if tahunMap[targetReq.Tahun] {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"tahun target %s duplikat dalam indikator yang sama",
					targetReq.Tahun,
				)
			}
			tahunMap[targetReq.Tahun] = true

			// Validasi target dan satuan tidak boleh kosong
			if targetReq.Target == "" {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"target untuk tahun %s tidak boleh kosong",
					targetReq.Tahun,
				)
			}
			if targetReq.Satuan == "" {
				return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
					"satuan untuk tahun %s tidak boleh kosong",
					targetReq.Tahun,
				)
			}

			var targetId string
			// Gunakan ID lama jika ada, generate baru jika tidak ada
			if targetReq.Id != "" {
				targetId = targetReq.Id
			} else {
				uuidTrg := uuid.New().String()[:5]
				targetId = fmt.Sprintf("TRG-TJN-%s", uuidTrg)
			}

			targetDomain := domain.Target{
				Id:          targetId,
				IndikatorId: indikatorId,
				Target:      targetReq.Target,
				Satuan:      targetReq.Satuan,
				Tahun:       targetReq.Tahun,
			}
			indikatorDomain.Target = append(indikatorDomain.Target, targetDomain)
		}

		// Validasi minimal harus ada 1 target
		if len(indikatorReq.Target) == 0 {
			return tujuanopd.TujuanOpdResponse{}, fmt.Errorf(
				"indikator harus memiliki minimal 1 target dalam rentang periode %d-%d",
				tahunAwal,
				tahunAkhir,
			)
		}

		tujuanOpd.Indikator = append(tujuanOpd.Indikator, indikatorDomain)
	}

	// Update semua data ke database
	err = service.TujuanOpdRepository.Update(ctx, tx, tujuanOpd)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	return helper.ToTujuanOpdResponse(tujuanOpd), nil
}

func (service *TujuanOpdServiceImpl) Delete(ctx context.Context, tujuanOpdId int) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	_, err = service.TujuanOpdRepository.FindById(ctx, tx, tujuanOpdId)
	if err != nil {
		return err
	}

	return service.TujuanOpdRepository.Delete(ctx, tx, tujuanOpdId)
}

func (service *TujuanOpdServiceImpl) FindById(ctx context.Context, tujuanOpdId int) (tujuanopd.TujuanOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	tujuanOpd, err := service.TujuanOpdRepository.FindById(ctx, tx, tujuanOpdId)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	// Ambil data OPD
	opd, err := service.OpdRepository.FindByKodeOpd(ctx, tx, tujuanOpd.KodeOpd)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	// Ambil data bidang urusan
	bidangUrusan, err := service.BidangUrusanRepository.FindByKodeBidangUrusan(ctx, tx, tujuanOpd.KodeBidangUrusan)
	if err != nil {
		return tujuanopd.TujuanOpdResponse{}, err
	}

	response := tujuanopd.TujuanOpdResponse{
		Id:               tujuanOpd.Id,
		KodeBidangUrusan: tujuanOpd.KodeBidangUrusan,
		NamaBidangUrusan: bidangUrusan.NamaBidangUrusan,
		KodeOpd:          tujuanOpd.KodeOpd,
		NamaOpd:          opd.NamaOpd,
		Tujuan:           tujuanOpd.Tujuan,
		TahunAwal:        tujuanOpd.TahunAwal,
		TahunAkhir:       tujuanOpd.TahunAkhir,
		JenisPeriode:     tujuanOpd.JenisPeriode,
		Indikator:        make([]tujuanopd.IndikatorResponse, 0),
	}

	for _, indikator := range tujuanOpd.Indikator {
		indikatorResponse := tujuanopd.IndikatorResponse{
			Id:               indikator.Id,
			IdTujuanOpd:      tujuanOpd.Id,
			NamaIndikator:    indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String,
			SumberData:       indikator.SumberData.String,
			Target:           make([]tujuanopd.TargetResponse, 0),
		}

		tahunAwalInt, _ := strconv.Atoi(tujuanOpd.TahunAwal)
		tahunAkhirInt, _ := strconv.Atoi(tujuanOpd.TahunAkhir)

		// Buat map untuk target yang ada
		targetMap := make(map[string]domain.Target)
		for _, t := range indikator.Target {
			if t.Id != "" {
				targetMap[t.Tahun] = t
			}
		}

		// Generate target untuk setiap tahun dalam range
		for year := tahunAwalInt; year <= tahunAkhirInt; year++ {
			tahunStr := strconv.Itoa(year)
			if target, exists := targetMap[tahunStr]; exists {
				targetResponse := tujuanopd.TargetResponse{
					Id:              target.Id,
					IndikatorId:     indikator.Id,
					Tahun:           tahunStr,
					TargetIndikator: target.Target,
					SatuanIndikator: target.Satuan,
				}
				indikatorResponse.Target = append(indikatorResponse.Target, targetResponse)
			} else {
				targetResponse := tujuanopd.TargetResponse{
					Id:              "",
					IndikatorId:     indikator.Id,
					Tahun:           tahunStr,
					TargetIndikator: "",
					SatuanIndikator: "",
				}
				indikatorResponse.Target = append(indikatorResponse.Target, targetResponse)
			}
		}

		response.Indikator = append(response.Indikator, indikatorResponse)
	}

	return response, nil
}

func (service *TujuanOpdServiceImpl) FindAll(ctx context.Context, kodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]tujuanopd.TujuanOpdwithBidangUrusanResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi tahun
	if len(tahunAwal) != 4 || len(tahunAkhir) != 4 {
		return nil, fmt.Errorf("format tahun tidak valid")
	}
	if _, err := strconv.Atoi(tahunAwal); err != nil {
		return nil, fmt.Errorf("tahun awal harus berupa angka")
	}
	if _, err := strconv.Atoi(tahunAkhir); err != nil {
		return nil, fmt.Errorf("tahun akhir harus berupa angka")
	}

	// Ambil data OPD
	opd, err := service.OpdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
	if err != nil {
		return nil, err
	}

	// Ambil semua tujuan OPD
	tujuanOpds, err := service.TujuanOpdRepository.FindAll(ctx, tx, kodeOpd, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		if err == sql.ErrNoRows {
			return make([]tujuanopd.TujuanOpdwithBidangUrusanResponse, 0), nil
		}
		return nil, err
	}

	// Buat map untuk mengelompokkan response berdasarkan kode_bidang_urusan
	responseMap := make(map[string]*tujuanopd.TujuanOpdwithBidangUrusanResponse)

	for _, tujuan := range tujuanOpds {
		// Ambil data bidang urusan
		bidangUrusan, err := service.BidangUrusanRepository.FindByKodeBidangUrusan(ctx, tx, tujuan.KodeBidangUrusan)
		if err != nil {
			return nil, err
		}

		tujuanResponse := tujuanopd.TujuanOpdResponse{
			Id:           tujuan.Id,
			Tujuan:       tujuan.Tujuan,
			TahunAwal:    tujuan.TahunAwal,
			TahunAkhir:   tujuan.TahunAkhir,
			JenisPeriode: tujuan.JenisPeriode,
			Indikator:    make([]tujuanopd.IndikatorResponse, 0),
		}

		// Proses indikator dan target seperti sebelumnya
		for _, indikator := range tujuan.Indikator {
			indikatorResponse := tujuanopd.IndikatorResponse{
				Id:               indikator.Id,
				IdTujuanOpd:      tujuan.Id,
				NamaIndikator:    indikator.Indikator,
				RumusPerhitungan: indikator.RumusPerhitungan.String,
				SumberData:       indikator.SumberData.String,
				Target:           make([]tujuanopd.TargetResponse, 0),
			}

			tahunAwalInt, _ := strconv.Atoi(tujuan.TahunAwal)
			tahunAkhirInt, _ := strconv.Atoi(tujuan.TahunAkhir)

			// Buat map untuk target yang ada
			targetMap := make(map[string]domain.Target)
			for _, t := range indikator.Target {
				if t.Id != "" {
					targetMap[t.Tahun] = t
				}
			}

			// Generate target untuk setiap tahun dalam range
			for year := tahunAwalInt; year <= tahunAkhirInt; year++ {
				tahunStr := strconv.Itoa(year)
				if target, exists := targetMap[tahunStr]; exists {
					targetResponse := tujuanopd.TargetResponse{
						Id:              target.Id,
						IndikatorId:     indikator.Id,
						Tahun:           tahunStr,
						TargetIndikator: target.Target,
						SatuanIndikator: target.Satuan,
					}
					indikatorResponse.Target = append(indikatorResponse.Target, targetResponse)
				} else {
					targetResponse := tujuanopd.TargetResponse{
						Id:              "",
						IndikatorId:     indikator.Id,
						Tahun:           tahunStr,
						TargetIndikator: "",
						SatuanIndikator: "",
					}
					indikatorResponse.Target = append(indikatorResponse.Target, targetResponse)
				}
			}

			tujuanResponse.Indikator = append(tujuanResponse.Indikator, indikatorResponse)
		}

		// Cek apakah sudah ada entry untuk kode_bidang_urusan ini
		mapKey := tujuan.KodeBidangUrusan
		if mapKey == "" {
			mapKey = "000" // Gunakan key default untuk bidang urusan kosong
		}

		if existing, exists := responseMap[mapKey]; exists {
			// Jika sudah ada, tambahkan tujuan ke array tujuan yang ada
			existing.TujuanOpd = append(existing.TujuanOpd, tujuanResponse)
		} else {
			// Jika belum ada, buat entry baru
			kodeUrusan := ""
			if len(bidangUrusan.KodeBidangUrusan) > 0 {
				kodeUrusan = bidangUrusan.KodeBidangUrusan[:1]
			}

			responseMap[mapKey] = &tujuanopd.TujuanOpdwithBidangUrusanResponse{
				Urusan:           bidangUrusan.NamaUrusan,
				KodeUrusan:       kodeUrusan,
				KodeBidangUrusan: bidangUrusan.KodeBidangUrusan,
				NamaBidangUrusan: bidangUrusan.NamaBidangUrusan,
				KodeOpd:          tujuan.KodeOpd,
				NamaOpd:          opd.NamaOpd,
				TujuanOpd:        []tujuanopd.TujuanOpdResponse{tujuanResponse},
			}
		}
	}

	// Convert map to slice
	var responses []tujuanopd.TujuanOpdwithBidangUrusanResponse
	for _, response := range responseMap {
		responses = append(responses, *response)
	}

	// Sort responses berdasarkan kode_bidang_urusan
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].KodeBidangUrusan < responses[j].KodeBidangUrusan
	})

	if len(responses) == 0 {
		responses = make([]tujuanopd.TujuanOpdwithBidangUrusanResponse, 0)
	}

	return responses, nil
}

func (service *TujuanOpdServiceImpl) GetByTahun(ctx context.Context, tahun string, kodeOpd string) ([]tujuanopd.TujuanOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return []tujuanopd.TujuanOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	tujuanOpdList, err := service.TujuanOpdRepository.GetByTahun(ctx, tx, tahun, kodeOpd)
	if err != nil {
		return []tujuanopd.TujuanOpdResponse{}, err
	}

	// untuk indikator
	tujuanMap := make(map[int]*tujuanopd.TujuanOpdResponse)

	result := make([]tujuanopd.TujuanOpdResponse, 0, len(tujuanOpdList))
	for _, tujuanOpd := range tujuanOpdList {
		resp := tujuanopd.TujuanOpdResponse{
			Id:               tujuanOpd.Id,
			KodeOpd:          tujuanOpd.KodeOpd,
			NamaOpd:          tujuanOpd.NamaOpd,
			KodeBidangUrusan: tujuanOpd.KodeBidangUrusan,
			Tujuan:           tujuanOpd.Tujuan,
			TahunAwal:        tujuanOpd.TahunAwal,
			TahunAkhir:       tujuanOpd.TahunAkhir,
			JenisPeriode:     tujuanOpd.JenisPeriode,
			Indikator:        []tujuanopd.IndikatorResponse{},
		}
		result = append(result, resp)
		tujuanMap[tujuanOpd.Id] = &result[len(result)-1]
	}

	// indikator
	indikatorList, err := service.TujuanOpdRepository.GetIndikatorTujuanOpdByTahun(ctx, tx, tahun, kodeOpd)
	if err != nil {
		return result, nil
	}

	for _, indikator := range indikatorList {
		if tujuanOpd, ok := tujuanMap[indikator.TujuanOpdId]; ok {
			targetResponse := make([]tujuanopd.TargetResponse, 0, len(indikator.Target))
			for _, target := range indikator.Target {
				targetResponse = append(targetResponse, tujuanopd.TargetResponse{
					Id:              target.Id,
					IndikatorId:     target.IndikatorId,
					TargetIndikator: target.Target,
					SatuanIndikator: target.Satuan,
					Tahun:           target.Tahun,
				})
			}

			tujuanOpd.Indikator = append(tujuanOpd.Indikator, tujuanopd.IndikatorResponse{
				Id:               indikator.Id,
				IdTujuanOpd:      indikator.TujuanOpdId,
				NamaIndikator:    indikator.Indikator,
				SumberData:       nullStringToString(indikator.SumberData),
				RumusPerhitungan: nullStringToString(indikator.RumusPerhitungan),
				Target:           targetResponse,
			})
		}
	}
	return result, nil
}
