package service

import (
	"context"
	"ekak_kabupaten_madiun/model/web/tujuanopd"
)

type TujuanOpdService interface {
	Create(ctx context.Context, request tujuanopd.TujuanOpdCreateRequest) (tujuanopd.TujuanOpdResponse, error)
	Update(ctx context.Context, request tujuanopd.TujuanOpdUpdateRequest) (tujuanopd.TujuanOpdResponse, error)
	Delete(ctx context.Context, tujuanOpdId int) error
	FindById(ctx context.Context, tujuanOpdId int) (tujuanopd.TujuanOpdResponse, error)
	FindAll(ctx context.Context, kodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]tujuanopd.TujuanOpdwithBidangUrusanResponse, error)
	GetByTahun(ctx context.Context, tahun string, kodeOpd string) ([]tujuanopd.TujuanOpdResponse, error)
}
