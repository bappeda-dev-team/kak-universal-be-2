package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"ekak_kabupaten_madiun/model/web/iku"
	"ekak_kabupaten_madiun/repository"
	"sort"
	"strconv"
)

type IkuServiceImpl struct {
	IkuRepository          repository.IkuRepository
	DB                     *sql.DB
	TujuanPemdaRepository  repository.TujuanPemdaRepository
	SasaranPemdaRepository repository.SasaranPemdaRepository
}

func NewIkuServiceImpl(ikuRepository repository.IkuRepository, db *sql.DB,
	tujuanRepo repository.TujuanPemdaRepository,
	sasaranRepo repository.SasaranPemdaRepository) *IkuServiceImpl {
	return &IkuServiceImpl{
		IkuRepository:          ikuRepository,
		DB:                     db,
		TujuanPemdaRepository:  tujuanRepo,
		SasaranPemdaRepository: sasaranRepo,
	}
}

func (service *IkuServiceImpl) FindAll(ctx context.Context, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]iku.IkuResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil data dari repository dengan parameter baru
	indikatorTargets, err := service.IkuRepository.FindAll(ctx, tx, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		return nil, err
	}

	// Transform ke response
	var responses []iku.IkuResponse
	for _, item := range indikatorTargets {
		var targetResponses []iku.TargetResponse
		for _, target := range item.Target {
			targetResponses = append(targetResponses, iku.TargetResponse{
				Target: target.Target,
				Satuan: target.Satuan,
				Tahun:  target.Tahun,
			})
		}

		responses = append(responses, iku.IkuResponse{
			IndikatorId:      item.Id,
			Sumber:           item.Sumber,
			IsActive:         item.IsActive,
			Indikator:        item.Indikator,
			RumusPerhitungan: item.RumusPerhitungan.String,
			SumberData:       item.SumberData.String,
			CreatedAt:        item.CreatedAt,
			TahunAwal:        item.TahunAwal,
			TahunAkhir:       item.TahunAkhir,
			JenisPeriode:     item.JenisPeriode,
			Target:           targetResponses,
		})
	}

	return responses, nil
}

func (service *IkuServiceImpl) FindAllIkuOpd(ctx context.Context, kodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]iku.IkuOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	indikators, err := service.IkuRepository.FindAllIkuOpd(ctx, tx, kodeOpd, tahunAwal, tahunAkhir, jenisPeriode)
	if err != nil {
		return nil, err
	}

	var responses []iku.IkuOpdResponse
	for _, item := range indikators {
		var targetResponses []iku.TargetOpdResponse

		// Pastikan target terurut berdasarkan tahun
		sort.Slice(item.Target, func(i, j int) bool {
			tahunI, _ := strconv.Atoi(item.Target[i].Tahun)
			tahunJ, _ := strconv.Atoi(item.Target[j].Tahun)
			return tahunI < tahunJ
		})

		// Konversi semua target, termasuk yang kosong
		for _, target := range item.Target {
			targetResponses = append(targetResponses, iku.TargetOpdResponse{
				Target: target.Target,
				Satuan: target.Satuan,
				Tahun:  target.Tahun,
			})
		}

		responses = append(responses, iku.IkuOpdResponse{
			IndikatorId:      item.Id,
			AsalIku:          item.AsalIku,
			Indikator:        item.Indikator,
			RumusPerhitungan: item.RumusPerhitungan.String,
			SumberData:       item.SumberData.String,
			CreatedAt:        item.CreatedAt,
			TahunAwal:        item.TahunAwal,
			TahunAkhir:       item.TahunAkhir,
			JenisPeriode:     item.JenisPeriode,
			Target:           targetResponses,
		})
	}

	// Urutkan responses berdasarkan CreatedAt
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].CreatedAt.Before(responses[j].CreatedAt)
	})

	if len(responses) == 0 {
		responses = make([]iku.IkuOpdResponse, 0)
	}

	return responses, nil
}

func (service *IkuServiceImpl) GetByTahun(ctx context.Context, tahun string) ([]iku.IkuResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil semua TujuanPemda
	indikatorTujuans, err := service.TujuanPemdaRepository.GetAllIndikatorTujuanPemdaByTahun(ctx, tx, tahun)
	if err != nil {
		return nil, err
	}

	// Ambil semua SasaranPemda
	indikatorSasarans, err := service.SasaranPemdaRepository.GetAllIndikatorSasaranPemdaByTahun(ctx, tx, tahun)
	if err != nil {
		return nil, err
	}

	var hasil []iku.IkuResponse

	// Proses indikator dari TujuanPemda
	for _, indikator := range indikatorTujuans {
		hasil = append(hasil, iku.IkuResponse{
			IndikatorId:      indikator.Id,
			Sumber:           "TujuanPemda",
			IsActive:         indikator.IsActive,
			Indikator:        indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String, // sql.NullString
			SumberData:       indikator.SumberData.String,       // sql.NullString
			CreatedAt:        indikator.CreatedAt,
			TahunAwal:        indikator.TahunAwal,
			TahunAkhir:       indikator.TahunAkhir,
			JenisPeriode:     indikator.JenisPeriode,
			Target:           toTargetResponse(indikator.Target),
		})
	}

	// Proses indikator dari SasaranPemda
	for _, indikator := range indikatorSasarans {
		hasil = append(hasil, iku.IkuResponse{
			IndikatorId:      indikator.Id,
			Sumber:           "SasaranPemda",
			IsActive:         indikator.IsActive,
			Indikator:        indikator.Indikator,
			RumusPerhitungan: indikator.RumusPerhitungan.String,
			SumberData:       indikator.SumberData.String,
			CreatedAt:        indikator.CreatedAt,
			TahunAwal:        indikator.TahunAwal,
			TahunAkhir:       indikator.TahunAkhir,
			JenisPeriode:     indikator.JenisPeriode,
			Target:           toTargetResponse(indikator.Target),
		})
	}

	return hasil, nil
}

func toTargetResponse(targets []domain.Target) []iku.TargetResponse {
	var result []iku.TargetResponse
	for _, t := range targets {
		result = append(result, iku.TargetResponse{
			Id:     t.Id,
			Tahun:  t.Tahun,
			Target: t.Target,
			Satuan: t.Satuan,
		})
	}
	return result
}
