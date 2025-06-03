package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
)

type PohonKinerjaRepository interface {
	//pokin opd
	Create(ctx context.Context, tx *sql.Tx, pohonKinerja domain.PohonKinerja) (domain.PohonKinerja, error)
	Update(ctx context.Context, tx *sql.Tx, pohonKinerja domain.PohonKinerja) (domain.PohonKinerja, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	FindById(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error)
	FindAll(ctx context.Context, tx *sql.Tx, kodeOpd, tahun string) ([]domain.PohonKinerja, error)
	FindStrategicNoParent(ctx context.Context, tx *sql.Tx, levelPohon, parent int, kodeOpd, tahun string) ([]domain.PohonKinerja, error)
	FindPelaksanaPokin(ctx context.Context, tx *sql.Tx, pohonKinerjaId string) ([]domain.PelaksanaPokin, error)
	DeletePelaksanaPokin(ctx context.Context, tx *sql.Tx, pelaksanaId string) error
	UpdatePokinStatusFromApproved(ctx context.Context, tx *sql.Tx, id int) error
	UpdateParent(ctx context.Context, tx *sql.Tx, pohonKinerja domain.PohonKinerja) (domain.PohonKinerja, error)
	FindidPokinWithAllTema(ctx context.Context, tx *sql.Tx, id int) ([]domain.PohonKinerja, error)
	CheckAsalPokin(ctx context.Context, tx *sql.Tx, id int) (int, error)
	DeletePokinWithIndikatorAndTarget(ctx context.Context, tx *sql.Tx, id int) error
	//admin pokin
	CreatePokinAdmin(ctx context.Context, tx *sql.Tx, pokinAdmin domain.PohonKinerja) (domain.PohonKinerja, error)
	UpdatePokinAdmin(ctx context.Context, tx *sql.Tx, pokinAdmin domain.PohonKinerja) (domain.PohonKinerja, error)
	DeletePokinAdmin(ctx context.Context, tx *sql.Tx, id int) error
	FindPokinAdminById(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error)
	FindPokinAdminAll(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.PohonKinerja, error)
	FindPokinAdminByIdHierarki(ctx context.Context, tx *sql.Tx, idPokin int) ([]domain.PohonKinerja, error)
	FindIndikatorByPokinId(ctx context.Context, tx *sql.Tx, pokinId string) ([]domain.Indikator, error)
	FindTargetByIndikatorId(ctx context.Context, tx *sql.Tx, indikatorId string) ([]domain.Target, error)
	FindPokinToClone(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error)
	ValidateParentLevel(ctx context.Context, tx *sql.Tx, parentId int, levelPohon int) error
	FindIndikatorToClone(ctx context.Context, tx *sql.Tx, pokinId int) ([]domain.Indikator, error)
	FindTargetToClone(ctx context.Context, tx *sql.Tx, indikatorId string) ([]domain.Target, error)
	InsertClonedPokin(ctx context.Context, tx *sql.Tx, pokin domain.PohonKinerja) (int64, error)
	InsertClonedIndikator(ctx context.Context, tx *sql.Tx, indikatorId string, pokinId int64, indikator domain.Indikator) error
	InsertClonedTarget(ctx context.Context, tx *sql.Tx, targetId string, indikatorId string, target domain.Target) error
	UpdatePokinStatus(ctx context.Context, tx *sql.Tx, id int, status string) error
	CheckPokinStatus(ctx context.Context, tx *sql.Tx, id int) (string, error)
	InsertClonedPokinWithStatus(ctx context.Context, tx *sql.Tx, pokin domain.PohonKinerja) (int64, error)
	UpdatePokinStatusTolak(ctx context.Context, tx *sql.Tx, id int, status string) error
	CheckCloneFrom(ctx context.Context, tx *sql.Tx, id int) (int, error)
	FindPokinByCloneFrom(ctx context.Context, tx *sql.Tx, cloneFromId int) ([]domain.PohonKinerja, error)
	FindIndikatorByCloneFrom(ctx context.Context, tx *sql.Tx, pokinId int, cloneFromId string) (domain.Indikator, error)
	FindTargetByCloneFrom(ctx context.Context, tx *sql.Tx, indikatorId string, cloneFromId string) (domain.Target, error)
	DeleteClonedPokinHierarchy(ctx context.Context, tx *sql.Tx, id int) error
	FindChildPokins(ctx context.Context, tx *sql.Tx, parentId int64) ([]domain.PohonKinerja, error)
	InsertClonedPelaksana(ctx context.Context, tx *sql.Tx, newId string, pokinId int64, pelaksana domain.PelaksanaPokin) error
	FindListOpdAllTematik(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.PohonKinerja, error)
	ValidateParentLevelTarikStrategiOpd(ctx context.Context, tx *sql.Tx, parentId int, childLevel int) error

	ValidatePokinId(ctx context.Context, tx *sql.Tx, pokinId int) error
	ValidatePokinLevel(ctx context.Context, tx *sql.Tx, pokinId int, expectedLevel int, purpose string) error

	//find pokin for dropdown
	FindPokinByJenisPohon(ctx context.Context, tx *sql.Tx, jenisPohon string, levelPohon int, tahun string, kodeOpd string, status string) ([]domain.PohonKinerja, error)
	FindPokinByPelaksana(ctx context.Context, tx *sql.Tx, pelaksanaId string, tahun string) ([]domain.PohonKinerja, error)
	FindPokinByStatus(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string, status string) ([]domain.PohonKinerja, error)
	FindPokinByCrosscuttingStatus(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string) ([]domain.PohonKinerja, error)

	//pokin for tujuan and sasaran pemda
	FindPokinWithPeriode(ctx context.Context, tx *sql.Tx, pokinId int, jenisPeriode string) (domain.PohonKinerja, domain.Periode, error)

	//tematik aktif/nonaktif
	UpdateTematikStatus(ctx context.Context, tx *sql.Tx, id int, isActive bool) error
	GetChildrenAndClones(ctx context.Context, tx *sql.Tx, parentId int, isActivating bool) ([]int, error)

	//clone pokin opd
	ClonePokinOpd(ctx context.Context, tx *sql.Tx, kodeOpd string, sourceTahun string, targetTahun string) error
	IsExistsByTahun(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string) bool

	//count pokin pemda in opd
	CountPokinPemdaByLevel(ctx context.Context, tx *sql.Tx, kodeOpd, tahun string) (map[int]int, error)
}
