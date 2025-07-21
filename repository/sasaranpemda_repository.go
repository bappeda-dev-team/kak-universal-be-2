package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
)

type SasaranPemdaRepository interface {
	Create(ctx context.Context, tx *sql.Tx, sasaranPemda domain.SasaranPemda) (domain.SasaranPemda, error)
	CreateIndikator(ctx context.Context, tx *sql.Tx, indikator domain.Indikator) (domain.Indikator, error)
	CreateTarget(ctx context.Context, tx *sql.Tx, target domain.Target) error
	Update(ctx context.Context, tx *sql.Tx, sasaranPemda domain.SasaranPemda) (domain.SasaranPemda, error)
	Delete(ctx context.Context, tx *sql.Tx, sasaranPemdaId int) error
	FindById(ctx context.Context, tx *sql.Tx, sasaranPemdaId int) (domain.SasaranPemda, error)
	FindAll(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.SasaranPemda, error)
	DeleteIndikator(ctx context.Context, tx *sql.Tx, sasaranPemdaId int) error
	IsIdExists(ctx context.Context, tx *sql.Tx, id int) bool
	UpdatePeriode(ctx context.Context, tx *sql.Tx, sasaranPemda domain.SasaranPemda) (domain.SasaranPemda, error)
	FindAllWithPokin(ctx context.Context, tx *sql.Tx, tahunAwal, tahunAkhir, jenisPeriode string) ([]domain.PohonKinerjaWithSasaran, error)
	IsSubtemaIdExists(ctx context.Context, tx *sql.Tx, subtemaId int) bool
	GetIndikatorSasaranByTahun(ctx context.Context, tx *sql.Tx, sasaranPemdaId int, tahun string) ([]domain.Indikator, error)
	GetAllIndikatorSasaranPemda(ctx context.Context, tx *sql.Tx) ([]domain.Indikator, error)
	GetAllIndikatorSasaranPemdaByTahun(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.Indikator, error)
}
