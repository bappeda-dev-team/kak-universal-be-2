package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"ekak_kabupaten_madiun/model/web/sasaranpemda"
	"ekak_kabupaten_madiun/repository"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type SasaranPemdaServiceImpl struct {
	SasaranPemdaRepository repository.SasaranPemdaRepository
	PeriodeRepository      repository.PeriodeRepository
	PohonKinerjaRepository repository.PohonKinerjaRepository
	TujuanPemdaRepository  repository.TujuanPemdaRepository
	DB                     *sql.DB
}

func NewSasaranPemdaServiceImpl(sasaranPemdaRepository repository.SasaranPemdaRepository, periodeRepository repository.PeriodeRepository, pohonKinerjaRepository repository.PohonKinerjaRepository, tujuanPemdaRepository repository.TujuanPemdaRepository, DB *sql.DB) *SasaranPemdaServiceImpl {
	return &SasaranPemdaServiceImpl{
		SasaranPemdaRepository: sasaranPemdaRepository,
		PeriodeRepository:      periodeRepository,
		PohonKinerjaRepository: pohonKinerjaRepository,
		TujuanPemdaRepository:  tujuanPemdaRepository,
		DB:                     DB,
	}
}

func (service *SasaranPemdaServiceImpl) generateRandomId(ctx context.Context, tx *sql.Tx) int {
	rand.Seed(time.Now().UnixNano())
	for {
		// Generate random number between 10000-99999
		id := rand.Intn(90000) + 10000
		if !service.SasaranPemdaRepository.IsIdExists(ctx, tx, id) {
			return id
		}
	}
}

func generateIndikatorIdSasaran() string {
	currentYear := time.Now().Format("2006")
	uuid := uuid.New().String()[:5] // Mengambil 5 karakter pertama dari UUID
	return fmt.Sprintf("IND-SAS-PMD-%s-%s", currentYear, uuid)
}

func generateTargetIdSasaran() string {
	currentYear := time.Now().Format("2006")
	uuid := uuid.New().String()[:5]
	return fmt.Sprintf("TRG-SAS-PMD-%s-%s", currentYear, uuid)
}

func (service *SasaranPemdaServiceImpl) Create(ctx context.Context, request sasaranpemda.SasaranPemdaCreateRequest) (sasaranpemda.SasaranPemdaResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi pohon kinerja
	pokinData, err := service.PohonKinerjaRepository.FindById(ctx, tx, request.SubtemaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("pohon kinerja tidak ditemukan: %v", err)
	}

	// Validasi level pohon kinerja (1-3)
	if pokinData.LevelPohon < 1 || pokinData.LevelPohon > 3 {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("level pohon kinerja harus berada di antara 1-3, level saat ini: %d", pokinData.LevelPohon)
	}

	// Validasi tujuan pemda exists
	exists := service.TujuanPemdaRepository.IsIdExists(ctx, tx, request.TujuanPemdaId)
	if !exists {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("tujuan pemda dengan id %d tidak ditemukan", request.TujuanPemdaId)
	}

	// Validasi periode
	periode, err := service.PeriodeRepository.FindById(ctx, tx, request.PeriodeId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("periode tidak ditemukan: %v", err)
	}

	// Validasi tahun target untuk setiap indikator
	tahunAwal, _ := strconv.Atoi(periode.TahunAwal)
	tahunAkhir, _ := strconv.Atoi(periode.TahunAkhir)

	// Persiapkan slice indikator untuk domain
	var indikators []domain.Indikator

	for _, indikatorReq := range request.Indikator {
		tahunMap := make(map[string]bool)
		var targets []domain.Target

		for _, targetReq := range indikatorReq.Target {
			targetTahun, _ := strconv.Atoi(targetReq.Tahun)

			// Validasi rentang tahun
			if targetTahun < tahunAwal || targetTahun > tahunAkhir {
				return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf(
					"tahun target %d harus berada dalam rentang periode %d-%d",
					targetTahun, tahunAwal, tahunAkhir,
				)
			}

			// Validasi duplikasi tahun
			if tahunMap[targetReq.Tahun] {
				return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf(
					"duplikasi tahun %s pada indikator %s",
					targetReq.Tahun, indikatorReq.Indikator,
				)
			}
			tahunMap[targetReq.Tahun] = true

			// Tambahkan target
			target := domain.Target{
				Id:     generateTargetIdSasaran(),
				Target: targetReq.Target,
				Satuan: targetReq.Satuan,
				Tahun:  targetReq.Tahun,
			}
			targets = append(targets, target)
		}

		indikator := domain.Indikator{
			Id:        generateIndikatorIdSasaran(),
			Indikator: indikatorReq.Indikator,
			RumusPerhitungan: sql.NullString{
				String: indikatorReq.RumusPerhitungan,
				Valid:  true,
			},
			SumberData: sql.NullString{
				String: indikatorReq.SumberData,
				Valid:  true,
			},
			Target: targets,
		}
		indikators = append(indikators, indikator)
	}

	sasaranPemda := domain.SasaranPemda{
		Id:            service.generateRandomId(ctx, tx),
		SubtemaId:     request.SubtemaId,
		TujuanPemdaId: request.TujuanPemdaId,
		SasaranPemda:  request.SasaranPemda,
		PeriodeId:     request.PeriodeId,
		TahunAwal:     periode.TahunAwal,
		TahunAkhir:    periode.TahunAkhir,
		JenisPeriode:  periode.JenisPeriode,
		Indikator:     indikators,
	}

	result, err := service.SasaranPemdaRepository.Create(ctx, tx, sasaranPemda)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("gagal membuat sasaran pemda: %v", err)
	}

	// Konversi ke response secara manual
	var indikatorResponses []sasaranpemda.IndikatorResponse
	for _, indikator := range result.Indikator {
		var targetResponses []sasaranpemda.TargetResponse
		// Sort target berdasarkan tahun
		sort.Slice(indikator.Target, func(i, j int) bool {
			return indikator.Target[i].Tahun < indikator.Target[j].Tahun
		})

		for _, target := range indikator.Target {
			targetResponses = append(targetResponses, sasaranpemda.TargetResponse{
				Id:     target.Id,
				Target: target.Target,
				Satuan: target.Satuan,
				Tahun:  target.Tahun,
			})
		}

		indikatorResponses = append(indikatorResponses, sasaranpemda.IndikatorResponse{
			Id:               indikator.Id,
			Indikator:        indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String,
			SumberData:       indikator.SumberData.String,
			Target:           targetResponses,
		})
	}

	// Sort indikator berdasarkan ID
	sort.Slice(indikatorResponses, func(i, j int) bool {
		return indikatorResponses[i].Id < indikatorResponses[j].Id
	})

	return sasaranpemda.SasaranPemdaResponse{
		Id:            result.Id,
		TujuanPemdaId: result.TujuanPemdaId,
		SubtemaId:     result.SubtemaId,
		NamaSubtema:   pokinData.NamaPohon,
		SasaranPemda:  result.SasaranPemda,
		Periode: sasaranpemda.PeriodeResponse{
			Id:           periode.Id,
			TahunAwal:    periode.TahunAwal,
			TahunAkhir:   periode.TahunAkhir,
			JenisPeriode: periode.JenisPeriode,
		},
		Indikator: indikatorResponses,
	}, nil
}
func (service *SasaranPemdaServiceImpl) Update(ctx context.Context, request sasaranpemda.SasaranPemdaUpdateRequest) (sasaranpemda.SasaranPemdaResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Convert request.Id string ke int
	sasaranPemdaId := request.Id

	// Validasi tujuan pemda exists
	_, err = service.TujuanPemdaRepository.FindById(ctx, tx, request.TujuanPemdaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("tujuan pemda tidak ditemukan: %v", err)
	}

	if !service.TujuanPemdaRepository.IsIdExists(ctx, tx, request.TujuanPemdaId) {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("tujuan pemda dengan id %d tidak ditemukan", request.TujuanPemdaId)
	}
	// Validasi sasaran pemda pemda exists
	sasaranPemda, err := service.SasaranPemdaRepository.FindById(ctx, tx, sasaranPemdaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, err
	}
	//validasi pohon kinerja
	pokinData, err := service.PohonKinerjaRepository.FindById(ctx, tx, request.SubtemaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("pohon kinerja tidak ditemukan: %v", err)
	}

	// Validasi level pohon kinerja (1-3)
	if pokinData.LevelPohon < 1 || pokinData.LevelPohon > 3 {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("level pohon kinerja harus berada di antara 1-3, level saat ini: %d", pokinData.LevelPohon)
	}

	periode, err := service.PeriodeRepository.FindById(ctx, tx, request.PeriodeId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("periode tidak ditemukan: %v", err)
	}

	// Validasi tahun target untuk setiap indikator
	tahunAwal, _ := strconv.Atoi(periode.TahunAwal)
	tahunAkhir, _ := strconv.Atoi(periode.TahunAkhir)

	for _, indikatorReq := range request.Indikator {
		tahunMap := make(map[string]bool)
		for _, targetReq := range indikatorReq.Target {
			targetTahun, _ := strconv.Atoi(targetReq.Tahun)

			// Validasi rentang tahun
			if targetTahun < tahunAwal || targetTahun > tahunAkhir {
				return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf(
					"tahun target %d harus berada dalam rentang periode %d-%d",
					targetTahun, tahunAwal, tahunAkhir,
				)
			}

			// Validasi duplikasi tahun
			if tahunMap[targetReq.Tahun] {
				return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf(
					"duplikasi tahun %s pada indikator %s",
					targetReq.Tahun, indikatorReq.Indikator,
				)
			}
			tahunMap[targetReq.Tahun] = true
		}
	}

	// Update data sasaran pemda
	sasaranPemda.TujuanPemdaId = request.TujuanPemdaId
	sasaranPemda.SasaranPemda = request.SasaranPemda
	sasaranPemda.PeriodeId = request.PeriodeId
	sasaranPemda.TahunAwal = periode.TahunAwal
	sasaranPemda.TahunAkhir = periode.TahunAkhir
	sasaranPemda.JenisPeriode = periode.JenisPeriode

	// Proses indikator
	var indikators []domain.Indikator
	for _, indikatorReq := range request.Indikator {
		var targets []domain.Target

		// Proses target untuk setiap indikator
		for _, targetReq := range indikatorReq.Target {
			targetId := targetReq.Id
			if targetId == "" || targetId == "-" {
				// Generate ID baru jika ID kosong atau "-"
				targetId = generateTargetIdSasaran()
			}

			target := domain.Target{
				Id:          targetId,
				IndikatorId: indikatorReq.Id,
				Target:      targetReq.Target,
				Satuan:      targetReq.Satuan,
				Tahun:       targetReq.Tahun,
			}
			targets = append(targets, target)
		}

		// Buat atau gunakan ID indikator yang ada
		indikatorId := indikatorReq.Id
		if indikatorId == "" {
			indikatorId = generateIndikatorIdSasaran()
		}

		indikator := domain.Indikator{
			Id:            indikatorId,
			TujuanPemdaId: sasaranPemda.TujuanPemdaId,
			Indikator:     indikatorReq.Indikator,
			RumusPerhitungan: sql.NullString{
				String: indikatorReq.RumusPerhitungan,
				Valid:  true,
			},
			SumberData: sql.NullString{
				String: indikatorReq.SumberData,
				Valid:  true,
			},
			Target: targets,
		}
		indikators = append(indikators, indikator)
	}

	sasaranPemda.Indikator = indikators

	// Simpan semua perubahan
	result, err := service.SasaranPemdaRepository.Update(ctx, tx, sasaranPemda)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, err
	}

	// Ambil data periode untuk response
	periode, err = service.PeriodeRepository.FindById(ctx, tx, sasaranPemda.PeriodeId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("gagal mengambil data periode: %v", err)
	}

	return sasaranpemda.SasaranPemdaResponse{
		Id:            result.Id,
		TujuanPemdaId: result.TujuanPemdaId,
		SubtemaId:     result.SubtemaId,
		NamaSubtema:   pokinData.NamaPohon,
		SasaranPemda:  result.SasaranPemda,
		Periode: sasaranpemda.PeriodeResponse{
			Id:           periode.Id,
			TahunAwal:    periode.TahunAwal,
			TahunAkhir:   periode.TahunAkhir,
			JenisPeriode: periode.JenisPeriode,
		},
		Indikator: convertToIndikatorUpdateResponses(indikators),
	}, nil
}

func (service *SasaranPemdaServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	return service.SasaranPemdaRepository.Delete(ctx, tx, id)
}

func (service *SasaranPemdaServiceImpl) FindById(ctx context.Context, sasaranPemdaId int) (sasaranpemda.SasaranPemdaResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil data sasaran pemda
	sasaranPemda, err := service.SasaranPemdaRepository.FindById(ctx, tx, sasaranPemdaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, err
	}

	// Ambil data pohon kinerja untuk nama subtema
	pokinData, err := service.PohonKinerjaRepository.FindById(ctx, tx, sasaranPemda.SubtemaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("gagal mengambil data pohon kinerja: %v", err)
	}

	// Ambil data tujuan pemda
	tujuanPemda, err := service.TujuanPemdaRepository.FindById(ctx, tx, sasaranPemda.TujuanPemdaId)
	if err != nil {
		return sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("gagal mengambil data tujuan pemda: %v", err)
	}

	// Konversi indikator ke response
	var indikatorResponses []sasaranpemda.IndikatorResponse
	for _, indikator := range sasaranPemda.Indikator {
		var targetResponses []sasaranpemda.TargetResponse
		for _, target := range indikator.Target {
			targetResponses = append(targetResponses, sasaranpemda.TargetResponse{
				Id:     target.Id,
				Target: target.Target,
				Satuan: target.Satuan,
				Tahun:  target.Tahun,
			})
		}

		indikatorResponses = append(indikatorResponses, sasaranpemda.IndikatorResponse{
			Id:               indikator.Id,
			Indikator:        indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String,
			SumberData:       indikator.SumberData.String,
			Target:           targetResponses,
		})
	}

	return sasaranpemda.SasaranPemdaResponse{
		Id:            sasaranPemda.Id,
		TujuanPemdaId: sasaranPemda.TujuanPemdaId,
		TujuanPemda:   tujuanPemda.TujuanPemda,
		SubtemaId:     sasaranPemda.SubtemaId,
		NamaSubtema:   pokinData.NamaPohon,
		SasaranPemda:  sasaranPemda.SasaranPemda,
		JenisPohon:    sasaranPemda.JenisPohon,
		Periode: sasaranpemda.PeriodeResponse{
			Id:           sasaranPemda.PeriodeId,
			TahunAwal:    sasaranPemda.Periode.TahunAwal,
			TahunAkhir:   sasaranPemda.Periode.TahunAkhir,
			JenisPeriode: sasaranPemda.Periode.JenisPeriode,
		},
		Indikator: indikatorResponses,
	}, nil
}

func (service *SasaranPemdaServiceImpl) FindAll(ctx context.Context, tahun string) ([]sasaranpemda.SasaranPemdaResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return []sasaranpemda.SasaranPemdaResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	sasaranPemdaList, err := service.SasaranPemdaRepository.FindAll(ctx, tx, tahun)
	if err != nil {
		return []sasaranpemda.SasaranPemdaResponse{}, err
	}

	sasaranPemdaResponses := make([]sasaranpemda.SasaranPemdaResponse, 0, len(sasaranPemdaList))
	for _, sasaranPemda := range sasaranPemdaList {
		pokinData, err := service.PohonKinerjaRepository.FindById(ctx, tx, sasaranPemda.SubtemaId)
		if err != nil {
			return []sasaranpemda.SasaranPemdaResponse{}, fmt.Errorf("gagal mengambil data pohon kinerja: %v", err)
		}

		sasaranPemdaResponses = append(sasaranPemdaResponses, sasaranpemda.SasaranPemdaResponse{
			Id:           sasaranPemda.Id,
			SubtemaId:    sasaranPemda.SubtemaId,
			NamaSubtema:  pokinData.NamaPohon,
			SasaranPemda: sasaranPemda.SasaranPemda,
		})
	}

	return sasaranPemdaResponses, nil
}

func (service *SasaranPemdaServiceImpl) FindByTahun(ctx context.Context, tahun string) ([]sasaranpemda.SasaranPemdaMinimalResponse, error) {
	// cek koneksi db jika error berhenti disini
	tx, err := service.DB.Begin()
	if err != nil {
		return []sasaranpemda.SasaranPemdaMinimalResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// ambil data dari repo sasaran pemda
	sasaranPemdaList, err := service.SasaranPemdaRepository.FindAll(ctx, tx, tahun)
	if err != nil {
		return []sasaranpemda.SasaranPemdaMinimalResponse{}, err
	}

	// buat response
	sasaranPemdaResponses := make([]sasaranpemda.SasaranPemdaMinimalResponse, 0, len(sasaranPemdaList))
	for _, sasaranPemda := range sasaranPemdaList {
		// TODO refactor agar lebih efisien, this is ugly
		indikatorList, err := service.SasaranPemdaRepository.GetIndikatorSasaranByTahun(ctx, tx, sasaranPemda.Id, tahun)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] terjadi kesalahan mengambil indikator sasaran_pemda_id %d: %v", sasaranPemda.Id, err)
		}

		// susun response indikator target
		indikatorResponse := make([]sasaranpemda.IndikatorResponse, 0, len(indikatorList))
		for _, indikator := range indikatorList {
			targetResponse := make([]sasaranpemda.TargetResponse, 0, len(indikator.Target))
			for _, target := range indikator.Target {
				targetResponse = append(targetResponse, sasaranpemda.TargetResponse{
					Id: target.Id,
					Target: target.Target,
					Satuan: target.Satuan,
					Tahun: target.Tahun,
				})
			}

			indikatorResponse = append(indikatorResponse, sasaranpemda.IndikatorResponse{
				Id: indikator.Id,
				Indikator: indikator.Indikator,
				SumberData: nullStringToString(indikator.SumberData),
				RumusPerhitungan: nullStringToString(indikator.RumusPerhitungan),
				Target: targetResponse,
			})
		}

		sasaranPemdaResponses = append(sasaranPemdaResponses, sasaranpemda.SasaranPemdaMinimalResponse{
			Id: sasaranPemda.Id,
			SasaranPemda: sasaranPemda.SasaranPemda,
			TahunAwal: sasaranPemda.TahunAwal,
			TahunAkhir: sasaranPemda.TahunAkhir,
			Indikator: indikatorResponse,
		})
	}

	return sasaranPemdaResponses, nil
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func (service *SasaranPemdaServiceImpl) FindAllWithPokin(ctx context.Context, tahunAwal, tahunAkhir, jenisPeriode string) ([]sasaranpemda.TematikResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi periode
	query := "SELECT COUNT(*) FROM tb_periode WHERE tahun_awal = ? AND tahun_akhir = ? AND jenis_periode = ?"
	var count int
	err = tx.QueryRowContext(ctx, query, tahunAwal, tahunAkhir, jenisPeriode).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error validating periode: %v", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("periode dengan tahun_awal %s, tahun_akhir %s, dan jenis_periode %s tidak ditemukan",
			tahunAwal, tahunAkhir, jenisPeriode)
	}

	// Ambil data dari repository
	pokinData, err := service.SasaranPemdaRepository.FindAllWithPokin(ctx, tx, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		return nil, err
	}

	// Konversi dari domain ke response
	var result []sasaranpemda.TematikResponse
	for _, tematik := range pokinData {
		tematikResponse := sasaranpemda.TematikResponse{
			TematikId:   tematik.TematikId,
			NamaTematik: tematik.NamaTematik,
			Subtematik:  []sasaranpemda.SubtematikResponse{},
		}

		// Konversi subtematik (hanya level 1-3)
		for _, subtematik := range tematik.Subtematik {
			if subtematik.LevelPohon < 1 || subtematik.LevelPohon > 3 {
				continue
			}

			subtematikResponse := sasaranpemda.SubtematikResponse{
				SubtematikId:   subtematik.SubtematikId,
				NamaSubtematik: subtematik.NamaSubtematik,
				JenisPohon:     subtematik.JenisPohon,
				LevelPohon:     subtematik.LevelPohon,
				IsActive:       subtematik.IsActive,
				SasaranPemda:   []sasaranpemda.SasaranPemdaWithpokinResponse{},
			}

			// Konversi sasaran pemda
			for _, sasaran := range subtematik.SasaranPemdaList {
				sasaranResponse := sasaranpemda.SasaranPemdaWithpokinResponse{
					IdSasaranPemda: sasaran.Id,
					SasaranPemda:   sasaran.SasaranPemda,
					Periode: sasaranpemda.PeriodeResponse{
						TahunAwal:    tahunAwal,
						TahunAkhir:   tahunAkhir,
						JenisPeriode: jenisPeriode,
					},
					Indikator: []sasaranpemda.IndikatorSubtematikResponse{},
				}

				// Konversi indikator
				for _, indikator := range sasaran.Indikator {
					indikatorResponse := sasaranpemda.IndikatorSubtematikResponse{
						Id:               indikator.Id,
						Indikator:        indikator.Indikator,
						RumusPerhitungan: indikator.RumusPerhitungan.String,
						SumberData:       indikator.SumberData.String,
						Target:           []sasaranpemda.TargetResponse{},
					}

					// Buat target untuk setiap tahun dalam range
					tahunAwalInt, _ := strconv.Atoi(tahunAwal)
					tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)

					// Map untuk menyimpan target yang sudah ada
					existingTargets := make(map[string]domain.Target)
					for _, t := range indikator.Target {
						existingTargets[t.Tahun] = domain.Target{
							Id:     t.Id,
							Target: t.Target,
							Satuan: t.Satuan,
							Tahun:  t.Tahun,
						}
					}

					// Buat atau gunakan target untuk setiap tahun
					for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
						tahunStr := strconv.Itoa(tahun)
						var targetResponse sasaranpemda.TargetResponse

						if target, exists := existingTargets[tahunStr]; exists {
							targetResponse = sasaranpemda.TargetResponse{
								Id:     target.Id,
								Target: target.Target,
								Satuan: target.Satuan,
								Tahun:  tahunStr,
							}
						} else {
							targetResponse = sasaranpemda.TargetResponse{
								Id:     "-",
								Target: "",
								Satuan: "",
								Tahun:  tahunStr,
							}
						}
						indikatorResponse.Target = append(indikatorResponse.Target, targetResponse)
					}

					// Sort target berdasarkan tahun
					sort.Slice(indikatorResponse.Target, func(i, j int) bool {
						tahunI, _ := strconv.Atoi(indikatorResponse.Target[i].Tahun)
						tahunJ, _ := strconv.Atoi(indikatorResponse.Target[j].Tahun)
						return tahunI < tahunJ
					})

					sasaranResponse.Indikator = append(sasaranResponse.Indikator, indikatorResponse)
				}

				subtematikResponse.SasaranPemda = append(subtematikResponse.SasaranPemda, sasaranResponse)
			}

			tematikResponse.Subtematik = append(tematikResponse.Subtematik, subtematikResponse)
		}

		// Tambahkan tematik jika memiliki subtematik
		if len(tematikResponse.Subtematik) > 0 {
			result = append(result, tematikResponse)
		}
	}

	// Sort berdasarkan ID di setiap level
	sort.Slice(result, func(i, j int) bool {
		return result[i].TematikId < result[j].TematikId
	})

	for i := range result {
		sort.Slice(result[i].Subtematik, func(x, y int) bool {
			return result[i].Subtematik[x].SubtematikId < result[i].Subtematik[y].SubtematikId
		})

		for j := range result[i].Subtematik {
			sort.Slice(result[i].Subtematik[j].SasaranPemda, func(x, y int) bool {
				return result[i].Subtematik[j].SasaranPemda[x].IdSasaranPemda <
					result[i].Subtematik[j].SasaranPemda[y].IdSasaranPemda
			})

			for k := range result[i].Subtematik[j].SasaranPemda {
				sort.Slice(result[i].Subtematik[j].SasaranPemda[k].Indikator, func(x, y int) bool {
					return result[i].Subtematik[j].SasaranPemda[k].Indikator[x].Id <
						result[i].Subtematik[j].SasaranPemda[k].Indikator[y].Id
				})

				// Pastikan target terurut berdasarkan tahun
				for l := range result[i].Subtematik[j].SasaranPemda[k].Indikator {
					sort.Slice(result[i].Subtematik[j].SasaranPemda[k].Indikator[l].Target, func(x, y int) bool {
						tahunX, _ := strconv.Atoi(result[i].Subtematik[j].SasaranPemda[k].Indikator[l].Target[x].Tahun)
						tahunY, _ := strconv.Atoi(result[i].Subtematik[j].SasaranPemda[k].Indikator[l].Target[y].Tahun)
						return tahunX < tahunY
					})
				}
			}
		}
	}

	return result, nil
}

func convertToIndikatorUpdateResponses(indikators []domain.Indikator) []sasaranpemda.IndikatorResponse {
	if len(indikators) == 0 {
		return nil
	}

	responses := make([]sasaranpemda.IndikatorResponse, len(indikators))
	for i, indikator := range indikators {
		targetResponses := make([]sasaranpemda.TargetResponse, len(indikator.Target))
		for j, target := range indikator.Target {
			targetResponses[j] = sasaranpemda.TargetResponse{
				Id:     target.Id,
				Target: target.Target,
				Satuan: target.Satuan,
				Tahun:  target.Tahun,
			}
		}

		// Sort target berdasarkan tahun
		sort.Slice(targetResponses, func(i, j int) bool {
			return targetResponses[i].Tahun < targetResponses[j].Tahun
		})

		responses[i] = sasaranpemda.IndikatorResponse{
			Id:               indikator.Id,
			Indikator:        indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String,
			SumberData:       indikator.SumberData.String,
			Target:           targetResponses,
		}
	}

	// Sort indikator berdasarkan ID
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].Id < responses[j].Id
	})

	return responses
}
