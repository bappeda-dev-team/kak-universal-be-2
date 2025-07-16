package service

import (
	"context"
	"ekak_kabupaten_madiun/model/web/sasaranpemda"
)

type SasaranPemdaService interface {
	Create(ctx context.Context, request sasaranpemda.SasaranPemdaCreateRequest) (sasaranpemda.SasaranPemdaResponse, error)
	Update(ctx context.Context, request sasaranpemda.SasaranPemdaUpdateRequest) (sasaranpemda.SasaranPemdaResponse, error)
	Delete(ctx context.Context, id int) error
	FindById(ctx context.Context, sasaranPemdaId int) (sasaranpemda.SasaranPemdaResponse, error)
	FindAll(ctx context.Context, tahun string) ([]sasaranpemda.SasaranPemdaResponse, error)
	FindByTahun(ctx context.Context, tahun string) ([]sasaranpemda.SasaranPemdaMinimalResponse, error)
	FindAllWithPokin(ctx context.Context, tahunAwal, tahunAkhir, jenisPeriode string) ([]sasaranpemda.TematikResponse, error)
}
