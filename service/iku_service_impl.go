package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/web/iku"
	"ekak_kabupaten_madiun/repository"
	"sort"
	"strconv"
)

type IkuServiceImpl struct {
	IkuRepository repository.IkuRepository
	DB            *sql.DB
	TujuanPemdaRepository repository.TujuanPemdaRepository
	SasaranPemdaRepository repository.SasaranPemdaRepository
}

func NewIkuServiceImpl(ikuRepository repository.IkuRepository, db *sql.DB) *IkuServiceImpl {
	return &IkuServiceImpl{
		IkuRepository: ikuRepository,
		DB:            db,
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
	tujuanPemda, err := service.TujuanPemdaRepository.FindAll(ctx, tx, tahun, "rpjmd")
	if err != nil {
		return []iku.IkuResponse{}, err
	}
	sasaranPemda, err := service.SasaranPemdaRepository.FindAll(ctx, tx, tahun)
	if err != nil {
		return []iku.IkuResponse{}, err
	}
}
