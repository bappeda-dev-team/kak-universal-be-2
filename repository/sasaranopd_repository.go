package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
)

type SasaranOpdRepository interface {
	FindAll(ctx context.Context, tx *sql.Tx, KodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.SasaranOpd, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (*domain.SasaranOpd, error)
	FindIdPokinSasaran(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error)
	FindByIdSasaran(ctx context.Context, tx *sql.Tx, id int) (*domain.SasaranOpdDetail, error)
	Create(ctx context.Context, tx *sql.Tx, domain domain.SasaranOpdDetail) error
	Update(ctx context.Context, tx *sql.Tx, sasaranOpd domain.SasaranOpdDetail) (domain.SasaranOpdDetail, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	FindByIdPokin(ctx context.Context, tx *sql.Tx, idPokin int, tahun string) (*domain.SasaranOpd, error)
	GetByTahun(ctx context.Context, tx *sql.Tx, tahun string, kodeOpd string)([]domain.SasaranOpdTahunan, error)
	GetIndikatorSasaranOpdByTahun(ctx context.Context, tx *sql.Tx, tahun string, kodeOpd string ) ([]domain.Indikator, error)
}
