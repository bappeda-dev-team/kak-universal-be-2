package service

import (
	"context"
	"ekak_kabupaten_madiun/model/web/pohonkinerja"
	"ekak_kabupaten_madiun/model/web/sasaranopd"
)

type SasaranOpdService interface {
	FindAll(ctx context.Context, KodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]sasaranopd.SasaranOpdResponse, error)
	FindById(ctx context.Context, id int) (*sasaranopd.SasaranOpdResponse, error)
	Create(ctx context.Context, request sasaranopd.SasaranOpdCreateRequest) (*sasaranopd.SasaranOpdCreateResponse, error)
	Update(ctx context.Context, request sasaranopd.SasaranOpdUpdateRequest) (*sasaranopd.SasaranOpdCreateResponse, error)
	Delete(ctx context.Context, id string) error
	FindByIdPokin(ctx context.Context, idPokin int, tahun string) (*sasaranopd.SasaranOpdResponse, error)
	FindIdPokinSasaran(ctx context.Context, id int) (pohonkinerja.PohonKinerjaOpdResponse, error)
	GetByTahun(ctx context.Context, tahun string, kodeOpd string) ([]sasaranopd.SasaranOpdTahunanResponse, error)
}
