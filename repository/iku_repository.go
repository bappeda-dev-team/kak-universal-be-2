package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
)

type IkuRepository interface {
	FindAll(ctx context.Context, tx *sql.Tx, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.Indikator, error)
	FindAllIkuOpd(ctx context.Context, tx *sql.Tx, kodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.Indikator, error)
}
