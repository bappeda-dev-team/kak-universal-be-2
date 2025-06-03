package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"ekak_kabupaten_madiun/model/web/opdmaster"
	"ekak_kabupaten_madiun/model/web/pohonkinerja"
	"ekak_kabupaten_madiun/repository"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PohonKinerjaOpdServiceImpl struct {
	pohonKinerjaOpdRepository repository.PohonKinerjaRepository
	opdRepository             repository.OpdRepository
	pegawaiRepository         repository.PegawaiRepository
	tujuanOpdRepository       repository.TujuanOpdRepository
	crosscuttingOpdRepository repository.CrosscuttingOpdRepository
	reviewRepository          repository.ReviewRepository
	DB                        *sql.DB
	Validate                  *validator.Validate
}

func NewPohonKinerjaOpdServiceImpl(pohonKinerjaOpdRepository repository.PohonKinerjaRepository, opdRepository repository.OpdRepository, pegawaiRepository repository.PegawaiRepository, tujuanOpdRepository repository.TujuanOpdRepository, crosscuttingOpdRepository repository.CrosscuttingOpdRepository, reviewRepository repository.ReviewRepository, DB *sql.DB, validate *validator.Validate) *PohonKinerjaOpdServiceImpl {
	return &PohonKinerjaOpdServiceImpl{
		pohonKinerjaOpdRepository: pohonKinerjaOpdRepository,
		opdRepository:             opdRepository,
		pegawaiRepository:         pegawaiRepository,
		tujuanOpdRepository:       tujuanOpdRepository,
		crosscuttingOpdRepository: crosscuttingOpdRepository,
		reviewRepository:          reviewRepository,
		DB:                        DB,
		Validate:                  validate,
	}
}

func (service *PohonKinerjaOpdServiceImpl) Create(ctx context.Context, request pohonkinerja.PohonKinerjaCreateRequest) (pohonkinerja.PohonKinerjaOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi request
	if request.NamaPohon == "" {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("nama program tidak boleh kosong")
	}

	// Validasi kode OPD
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, request.KodeOpd)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("kode opd tidak ditemukan")
	}
	if opd.KodeOpd == "" {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("kode opd tidak valid")
	}

	// Validasi dan persiapan data pelaksana
	var pelaksanaList []domain.PelaksanaPokin
	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse

	for _, pelaksanaReq := range request.PelaksanaId {
		// Generate ID untuk pelaksana_pokin
		pelaksanaId := fmt.Sprintf("PLKS-%s", uuid.New().String()[:8])

		// Validasi setiap pelaksana
		pegawaiPelaksana, err := service.pegawaiRepository.FindById(ctx, tx, pelaksanaReq.PegawaiId)
		if err != nil {
			return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("pelaksana tidak ditemukan")
		}
		if pegawaiPelaksana.Id == "" {
			return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("pelaksana tidak ditemukan")
		}

		// Tambahkan ke list pelaksana
		pelaksanaList = append(pelaksanaList, domain.PelaksanaPokin{
			Id:        pelaksanaId,
			PegawaiId: pelaksanaReq.PegawaiId,
		})

		// Siapkan response pelaksana
		pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
			Id:          pelaksanaId,
			PegawaiId:   pegawaiPelaksana.Id,
			NamaPegawai: pegawaiPelaksana.NamaPegawai,
		})
	}

	// Validasi dan persiapan data indikator dan target
	var indikatorList []domain.Indikator
	var indikatorResponses []pohonkinerja.IndikatorResponse

	for _, indikatorReq := range request.Indikator {
		// Generate ID untuk indikator
		indikatorId := fmt.Sprintf("IND-%s", uuid.New().String()[:8])

		var targetList []domain.Target
		var targetResponses []pohonkinerja.TargetResponse

		// Proses target untuk setiap indikator
		for _, targetReq := range indikatorReq.Target {
			targetId := fmt.Sprintf("TRG-%s", uuid.New().String()[:8])

			target := domain.Target{
				Id:          targetId,
				IndikatorId: indikatorId,
				Target:      targetReq.Target,
				Satuan:      targetReq.Satuan,
				Tahun:       request.Tahun,
			}
			targetList = append(targetList, target)

			targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
				Id:              targetId,
				IndikatorId:     indikatorId,
				TargetIndikator: targetReq.Target,
				SatuanIndikator: targetReq.Satuan,
			})
		}

		indikator := domain.Indikator{
			Id:        indikatorId,
			Indikator: indikatorReq.NamaIndikator,
			Target:    targetList,
		}
		indikatorList = append(indikatorList, indikator)

		indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
			Id:            indikatorId,
			NamaIndikator: indikatorReq.NamaIndikator,
			Target:        targetResponses,
		})
	}

	pohonKinerja := domain.PohonKinerja{
		NamaPohon: request.NamaPohon,
		Parent:    request.Parent,

		JenisPohon: request.JenisPohon,
		LevelPohon: request.LevelPohon,
		KodeOpd:    request.KodeOpd,
		Keterangan: request.Keterangan,
		Tahun:      request.Tahun,
		Status:     request.Status,
		Pelaksana:  pelaksanaList,
		Indikator:  indikatorList,
	}

	result, err := service.pohonKinerjaOpdRepository.Create(ctx, tx, pohonKinerja)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}

	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, result.Id)
	helper.PanicIfError(err)

	response := pohonkinerja.PohonKinerjaOpdResponse{
		Id:          result.Id,
		Parent:      strconv.Itoa(result.Parent),
		NamaPohon:   result.NamaPohon,
		JenisPohon:  result.JenisPohon,
		LevelPohon:  result.LevelPohon,
		KodeOpd:     result.KodeOpd,
		NamaOpd:     opd.NamaOpd,
		Keterangan:  result.Keterangan,
		Tahun:       result.Tahun,
		Status:      result.Status,
		CountReview: countReview,
		Pelaksana:   pelaksanaResponses,
		Indikator:   indikatorResponses,
	}

	return response, nil
}

func (service *PohonKinerjaOpdServiceImpl) Update(ctx context.Context, request pohonkinerja.PohonKinerjaUpdateRequest) (pohonkinerja.PohonKinerjaOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi request
	if request.NamaPohon == "" {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("nama program tidak boleh kosong")
	}

	// Validasi kode OPD
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, request.KodeOpd)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("kode opd tidak ditemukan")
	}
	if opd.KodeOpd == "" {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("kode opd tidak valid")
	}

	// Cek apakah ini adalah pohon kinerja yang di-clone
	cloneFrom, err := service.pohonKinerjaOpdRepository.CheckCloneFrom(ctx, tx, request.Id)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}

	// Jika ini adalah pohon kinerja yang di-clone, tidak boleh diupdate
	if cloneFrom != 0 {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("tidak dapat mengupdate pohon kinerja yang merupakan hasil clone")
	}

	// Dapatkan semua pohon kinerja yang terkait (asli dan clone)
	var pokinsToUpdate []domain.PohonKinerja

	// Tambahkan pohon kinerja yang sedang diupdate
	existingPokin, err := service.pohonKinerjaOpdRepository.FindById(ctx, tx, request.Id)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("data pohon kinerja tidak ditemukan")
	}
	pokinsToUpdate = append(pokinsToUpdate, existingPokin)

	// Cari pohon kinerja yang merupakan clone dari yang sedang diupdate
	clonedPokins, err := service.pohonKinerjaOpdRepository.FindPokinByCloneFrom(ctx, tx, request.Id)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}
	pokinsToUpdate = append(pokinsToUpdate, clonedPokins...)

	// Persiapkan data pelaksana
	var pelaksanaList []domain.PelaksanaPokin
	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse

	for _, pelaksanaReq := range request.PelaksanaId {
		pelaksanaId := fmt.Sprintf("PLKS-%s", uuid.New().String()[:8])
		pegawaiPelaksana, err := service.pegawaiRepository.FindById(ctx, tx, pelaksanaReq.PegawaiId)
		if err != nil {
			return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("pelaksana tidak ditemukan")
		}

		pelaksanaList = append(pelaksanaList, domain.PelaksanaPokin{
			Id:        pelaksanaId,
			PegawaiId: pelaksanaReq.PegawaiId,
		})

		pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
			Id:          pelaksanaId,
			PegawaiId:   pegawaiPelaksana.Id,
			NamaPegawai: pegawaiPelaksana.NamaPegawai,
		})
	}

	// Persiapkan response indikator di luar loop
	var indikatorResponses []pohonkinerja.IndikatorResponse

	// Update untuk setiap pohon kinerja (asli dan clone)
	var updatedPokin domain.PohonKinerja
	for _, pokin := range pokinsToUpdate {
		var indikatorList []domain.Indikator

		for _, indikatorReq := range request.Indikator {
			var indikatorId string
			var cloneFromIndikator string

			if pokin.Id == request.Id {
				// Untuk pohon asli, gunakan ID dari request
				indikatorId = indikatorReq.Id
				if indikatorId == "" {
					indikatorId = fmt.Sprintf("IND-%s", uuid.New().String()[:8])
				}
				cloneFromIndikator = ""
			} else {
				// Untuk pohon clone, cari ID indikator yang sudah ada berdasarkan clone_from
				existingIndikator, err := service.pohonKinerjaOpdRepository.FindIndikatorByCloneFrom(ctx, tx, pokin.Id, indikatorReq.Id)
				if err == nil && existingIndikator.Id != "" {
					// Gunakan ID yang sudah ada jika ditemukan
					indikatorId = existingIndikator.Id
				} else {
					// Buat ID baru jika belum ada
					indikatorId = fmt.Sprintf("IND-%s", uuid.New().String()[:8])
				}
				cloneFromIndikator = indikatorReq.Id
			}

			var targetList []domain.Target
			var targetResponses []pohonkinerja.TargetResponse

			for _, targetReq := range indikatorReq.Target {
				var targetId string
				var cloneFromTarget string

				if pokin.Id == request.Id {
					// Untuk pohon asli, gunakan ID dari request
					targetId = targetReq.Id
					if targetId == "" {
						targetId = fmt.Sprintf("TRG-%s", uuid.New().String()[:8])
					}
					cloneFromTarget = ""
				} else {
					// Untuk pohon clone, cari ID target yang sudah ada berdasarkan clone_from
					existingTarget, err := service.pohonKinerjaOpdRepository.FindTargetByCloneFrom(ctx, tx, indikatorId, targetReq.Id)
					if err == nil && existingTarget.Id != "" {
						// Gunakan ID yang sudah ada jika ditemukan
						targetId = existingTarget.Id
					} else {
						// Buat ID baru jika belum ada
						targetId = fmt.Sprintf("TRG-%s", uuid.New().String()[:8])
					}
					cloneFromTarget = targetReq.Id
				}

				target := domain.Target{
					Id:          targetId,
					IndikatorId: indikatorId,
					Target:      targetReq.Target,
					Satuan:      targetReq.Satuan,
					Tahun:       request.Tahun,
					CloneFrom:   cloneFromTarget,
				}
				targetList = append(targetList, target)

				if pokin.Id == request.Id {
					targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
						Id:              targetId,
						IndikatorId:     indikatorId,
						TargetIndikator: targetReq.Target,
						SatuanIndikator: targetReq.Satuan,
					})
				}
			}

			indikator := domain.Indikator{
				Id:        indikatorId,
				PokinId:   fmt.Sprint(pokin.Id),
				Indikator: indikatorReq.NamaIndikator,
				Tahun:     request.Tahun,
				Target:    targetList,
				CloneFrom: cloneFromIndikator,
			}
			indikatorList = append(indikatorList, indikator)

			if pokin.Id == request.Id {
				indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
					Id:            indikatorId,
					IdPokin:       fmt.Sprint(pokin.Id),
					NamaIndikator: indikatorReq.NamaIndikator,
					Target:        targetResponses,
				})
			}
		}

		pohonKinerjaUpdate := domain.PohonKinerja{
			Id:                     pokin.Id,
			NamaPohon:              request.NamaPohon,
			Parent:                 request.Parent,
			JenisPohon:             request.JenisPohon,
			LevelPohon:             request.LevelPohon,
			KodeOpd:                request.KodeOpd,
			Keterangan:             request.Keterangan,
			Tahun:                  request.Tahun,
			Status:                 pokin.Status,
			CloneFrom:              pokin.CloneFrom,
			Pelaksana:              pelaksanaList,
			Indikator:              indikatorList,
			KeteranganCrosscutting: pokin.KeteranganCrosscutting,
		}

		result, err := service.pohonKinerjaOpdRepository.Update(ctx, tx, pohonKinerjaUpdate)
		if err != nil {
			return pohonkinerja.PohonKinerjaOpdResponse{}, err
		}

		if pokin.Id == request.Id {
			updatedPokin = result
		}
	}

	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, updatedPokin.Id)
	helper.PanicIfError(err)

	return pohonkinerja.PohonKinerjaOpdResponse{
		Id:                     updatedPokin.Id,
		Parent:                 strconv.Itoa(updatedPokin.Parent),
		NamaPohon:              updatedPokin.NamaPohon,
		JenisPohon:             updatedPokin.JenisPohon,
		LevelPohon:             updatedPokin.LevelPohon,
		KodeOpd:                updatedPokin.KodeOpd,
		NamaOpd:                opd.NamaOpd,
		Keterangan:             updatedPokin.Keterangan,
		Tahun:                  updatedPokin.Tahun,
		CountReview:            countReview,
		Status:                 updatedPokin.Status,
		Pelaksana:              pelaksanaResponses,
		Indikator:              indikatorResponses,
		KeteranganCrosscutting: updatedPokin.KeteranganCrosscutting,
	}, nil
}

func (service *PohonKinerjaOpdServiceImpl) Delete(ctx context.Context, id int) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(tx)

	// 1. Cek apakah pohon kinerja dengan ID tersebut ada
	_, err = service.pohonKinerjaOpdRepository.FindById(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("pohon kinerja tidak ditemukan: %v", err)
	}

	// 2. Lakukan penghapusan dengan fungsi baru
	err = service.pohonKinerjaOpdRepository.Delete(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus pohon kinerja: %v", err)
	}

	return nil
}

// func (service *PohonKinerjaOpdServiceImpl) Delete(ctx context.Context, id int) error {
// 	tx, err := service.DB.Begin()
// 	if err != nil {
// 		return fmt.Errorf("gagal memulai transaksi: %v", err)
// 	}
// 	defer helper.CommitOrRollback(tx)

// 	// 1. Cek apakah pohon kinerja dengan ID tersebut ada
// 	_, err = service.pohonKinerjaOpdRepository.FindById(ctx, tx, id)
// 	if err != nil {
// 		return fmt.Errorf("pohon kinerja tidak ditemukan: %v", err)
// 	}

// 	// 2. Dapatkan seluruh hierarki child
// 	childHierarchy, err := service.pohonKinerjaOpdRepository.FindPokinAdminByIdHierarki(ctx, tx, id)
// 	if err != nil {
// 		return fmt.Errorf("gagal mendapatkan hierarki pohon kinerja: %v", err)
// 	}

// 	// 3. Pertama, update status semua node yang memiliki clone_from
// 	for _, node := range childHierarchy {
// 		cloneFrom, err := service.pohonKinerjaOpdRepository.CheckCloneFrom(ctx, tx, node.Id)
// 		if err != nil && err != sql.ErrNoRows {
// 			return fmt.Errorf("gagal memeriksa clone_from untuk node ID %d: %v", node.Id, err)
// 		}

// 		if cloneFrom != 0 {
// 			// Update status node asli menjadi menunggu_disetujui
// 			err = service.pohonKinerjaOpdRepository.UpdatePokinStatusFromApproved(ctx, tx, cloneFrom)
// 			if err != nil {
// 				return fmt.Errorf("gagal mengupdate status node asli ID %d: %v", cloneFrom, err)
// 			}
// 		}
// 	}

// 	// 4. Kemudian, proses penghapusan
// 	for _, node := range childHierarchy {
// 		cloneFrom, err := service.pohonKinerjaOpdRepository.CheckCloneFrom(ctx, tx, node.Id)
// 		if err != nil && err != sql.ErrNoRows {
// 			return fmt.Errorf("gagal memeriksa clone_from untuk node ID %d: %v", node.Id, err)
// 		}

// 		if cloneFrom != 0 {
// 			// Hapus node clone dan seluruh hierarkinya
// 			err = service.pohonKinerjaOpdRepository.DeleteClonedPokinHierarchy(ctx, tx, node.Id)
// 			if err != nil {
// 				return fmt.Errorf("gagal menghapus hierarki clone untuk node ID %d: %v", node.Id, err)
// 			}
// 		} else {
// 			// Jika bukan clone, hapus node beserta indikator dan targetnya
// 			err = service.pohonKinerjaOpdRepository.DeletePokinWithIndikatorAndTarget(ctx, tx, node.Id)
// 			if err != nil {
// 				return fmt.Errorf("gagal menghapus node ID %d beserta indikator dan target: %v", node.Id, err)
// 			}
// 		}
// 	}

// 	return nil
// }

func (service *PohonKinerjaOpdServiceImpl) FindById(ctx context.Context, id int) (pohonkinerja.PohonKinerjaOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// 1. Ambil data pohon kinerja
	pokin, err := service.pohonKinerjaOpdRepository.FindById(ctx, tx, id)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}

	// 2. Validasi data pohon kinerja
	if pokin.Id == 0 {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("data tidak ditemukan")
	}

	// 3. Ambil data OPD
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("data opd tidak ditemukan")
	}

	// 4. Ambil data pelaksana
	pelaksanaList, err := service.pohonKinerjaOpdRepository.FindPelaksanaPokin(ctx, tx, fmt.Sprint(pokin.Id))
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, errors.New("gagal mengambil data pelaksana")
	}

	// 5. Proses data pelaksana
	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
	for _, pelaksana := range pelaksanaList {
		pegawaiPelaksana, err := service.pegawaiRepository.FindById(ctx, tx, pelaksana.PegawaiId)
		if err != nil {
			continue // Skip jika pegawai tidak ditemukan
		}

		pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
			Id:          pelaksana.Id,
			PegawaiId:   pegawaiPelaksana.Id,
			Nip:         pegawaiPelaksana.Nip,
			NamaPegawai: pegawaiPelaksana.NamaPegawai,
		})
	}

	// 6. Ambil data indikator dan target
	var indikatorResponses []pohonkinerja.IndikatorResponse
	indikatorList, err := service.pohonKinerjaOpdRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(pokin.Id))
	if err == nil {
		for _, indikator := range indikatorList {
			// Ambil target untuk setiap indikator
			targetList, err := service.pohonKinerjaOpdRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id)
			if err != nil {
				continue
			}

			var targetResponses []pohonkinerja.TargetResponse
			for _, target := range targetList {
				targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
					Id:              target.Id,
					IndikatorId:     target.IndikatorId,
					TargetIndikator: target.Target,
					SatuanIndikator: target.Satuan,
				})
			}

			indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
				Id:            indikator.Id,
				IdPokin:       indikator.PokinId,
				NamaIndikator: indikator.Indikator,
				Target:        targetResponses,
			})
		}
	}

	// 7. Susun response
	response := pohonkinerja.PohonKinerjaOpdResponse{
		Id:         pokin.Id,
		Parent:     strconv.Itoa(pokin.Parent),
		NamaPohon:  pokin.NamaPohon,
		JenisPohon: pokin.JenisPohon,
		LevelPohon: pokin.LevelPohon,
		KodeOpd:    pokin.KodeOpd,
		NamaOpd:    opd.NamaOpd,
		Keterangan: pokin.Keterangan,
		Tahun:      pokin.Tahun,
		Status:     pokin.Status,
		Pelaksana:  pelaksanaResponses,
		Indikator:  indikatorResponses,
	}

	return response, nil
}

//find all lama
// func (service *PohonKinerjaOpdServiceImpl) FindAll(ctx context.Context, kodeOpd, tahun string) (pohonkinerja.PohonKinerjaOpdAllResponse, error) {
// 	tx, err := service.DB.Begin()
// 	if err != nil {
// 		return pohonkinerja.PohonKinerjaOpdAllResponse{}, err
// 	}
// 	defer helper.CommitOrRollback(tx)

// 	// Validasi OPD
// 	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
// 	if err != nil {
// 		return pohonkinerja.PohonKinerjaOpdAllResponse{}, errors.New("kode opd tidak ditemukan")
// 	}

// 	// Inisialisasi response dasar
// 	response := pohonkinerja.PohonKinerjaOpdAllResponse{
// 		KodeOpd:    kodeOpd,
// 		NamaOpd:    opd.NamaOpd,
// 		Tahun:      tahun,
// 		TujuanOpd:  make([]pohonkinerja.TujuanOpdResponse, 0),
// 		Strategics: make([]pohonkinerja.StrategicOpdResponse, 0),
// 	}

// 	// Ambil data tujuan OPD
// 	tujuanOpds, err := service.tujuanOpdRepository.FindTujuanOpdByTahun(ctx, tx, kodeOpd, tahun, "RPJMD")
// 	if err != nil {
// 		log.Printf("Error getting tujuan OPD: %v", err)
// 		// Kembalikan response dengan array kosong jika terjadi error
// 		return response, nil
// 	}

// 	// Konversi tujuan OPD ke format response
// 	for _, tujuan := range tujuanOpds {
// 		indikators, err := service.tujuanOpdRepository.FindIndikatorByTujuanOpdId(ctx, tx, tujuan.Id)
// 		if err != nil {
// 			log.Printf("Error getting indikator for tujuan ID %d: %v", tujuan.Id, err)
// 			continue
// 		}

// 		var indikatorResponses []pohonkinerja.IndikatorTujuanResponse
// 		for _, indikator := range indikators {
// 			indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorTujuanResponse{
// 				Indikator: indikator.Indikator,
// 			})
// 		}

// 		response.TujuanOpd = append(response.TujuanOpd, pohonkinerja.TujuanOpdResponse{
// 			Id:        tujuan.Id,
// 			KodeOpd:   tujuan.KodeOpd,
// 			Tujuan:    tujuan.Tujuan,
// 			Indikator: indikatorResponses,
// 		})
// 	}

// 	// Ambil data pohon kinerja
// 	pokins, err := service.pohonKinerjaOpdRepository.FindAll(ctx, tx, kodeOpd, tahun)
// 	if err != nil {
// 		// Kembalikan response dengan data yang sudah ada jika terjadi error
// 		return response, nil
// 	}

// 	// Jika tidak ada data pohon kinerja, kembalikan response dengan array kosong
// 	if len(pokins) == 0 {
// 		return response, nil
// 	}

// 	// Proses data pohon kinerja seperti sebelumnya
// 	pohonMap := make(map[int]map[int][]domain.PohonKinerja)
// 	pelaksanaMap := make(map[int][]pohonkinerja.PelaksanaOpdResponse)
// 	indikatorMap := make(map[int][]pohonkinerja.IndikatorResponse)

// 	// Kelompokkan data dan ambil data pelaksana & indikator
// 	maxLevel := 0
// 	for _, p := range pokins {
// 		if p.LevelPohon >= 4 {
// 			// Update max level jika ditemukan level yang lebih tinggi
// 			if p.LevelPohon > maxLevel {
// 				maxLevel = p.LevelPohon
// 			}

// 			// Inisialisasi map untuk level jika belum ada
// 			if pohonMap[p.LevelPohon] == nil {
// 				pohonMap[p.LevelPohon] = make(map[int][]domain.PohonKinerja)
// 			}

// 			p.NamaOpd = opd.NamaOpd
// 			pohonMap[p.LevelPohon][p.Parent] = append(
// 				pohonMap[p.LevelPohon][p.Parent],
// 				p,
// 			)

// 			// Ambil data pelaksana
// 			pelaksanaList, err := service.pohonKinerjaOpdRepository.FindPelaksanaPokin(ctx, tx, fmt.Sprint(p.Id))
// 			if err == nil {
// 				var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
// 				for _, pelaksana := range pelaksanaList {
// 					pegawaiPelaksana, err := service.pegawaiRepository.FindById(ctx, tx, pelaksana.PegawaiId)
// 					if err != nil {
// 						continue
// 					}
// 					pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
// 						Id:          pelaksana.Id,
// 						PegawaiId:   pegawaiPelaksana.Id,
// 						NamaPegawai: pegawaiPelaksana.NamaPegawai,
// 					})
// 				}
// 				pelaksanaMap[p.Id] = pelaksanaResponses
// 			}

// 			// Ambil data indikator dan target
// 			indikatorList, err := service.pohonKinerjaOpdRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(p.Id))
// 			if err == nil {
// 				var indikatorResponses []pohonkinerja.IndikatorResponse
// 				for _, indikator := range indikatorList {
// 					// Ambil target untuk setiap indikator
// 					targetList, err := service.pohonKinerjaOpdRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id)
// 					if err != nil {
// 						continue
// 					}

// 					var targetResponses []pohonkinerja.TargetResponse
// 					for _, target := range targetList {
// 						targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
// 							Id:              target.Id,
// 							IndikatorId:     target.IndikatorId,
// 							TargetIndikator: target.Target,
// 							SatuanIndikator: target.Satuan,
// 						})
// 					}

// 					indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
// 						Id:            indikator.Id,
// 						IdPokin:       indikator.PokinId,
// 						NamaIndikator: indikator.Indikator,
// 						Target:        targetResponses,
// 					})
// 				}
// 				indikatorMap[p.Id] = indikatorResponses
// 			}
// 		}
// 	}

// 	// Build response untuk strategic (level 4)
// 	if strategicList := pohonMap[4]; len(strategicList) > 0 {
// 		for _, strategicsByParent := range strategicList {
// 			sort.Slice(strategicsByParent, func(i, j int) bool {
// 				return strategicsByParent[i].Id < strategicsByParent[j].Id
// 			})

// 			for _, strategic := range strategicsByParent {
// 				strategicResp := service.buildStrategicResponse(ctx, tx, pohonMap, strategic, pelaksanaMap, indikatorMap)
// 				response.Strategics = append(response.Strategics, strategicResp)
// 			}
// 		}

// 		// Urutkan strategics berdasarkan Id
// 		sort.Slice(response.Strategics, func(i, j int) bool {
// 			return response.Strategics[i].Id < response.Strategics[j].Id
// 		})
// 	}

// 	return response, nil
// }

// findall baru
func (service *PohonKinerjaOpdServiceImpl) FindAll(ctx context.Context, kodeOpd, tahun string) (pohonkinerja.PohonKinerjaOpdAllResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdAllResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdAllResponse{}, errors.New("kode opd tidak ditemukan")
	}

	// Inisialisasi response dasar
	response := pohonkinerja.PohonKinerjaOpdAllResponse{
		KodeOpd:    kodeOpd,
		NamaOpd:    opd.NamaOpd,
		Tahun:      tahun,
		TujuanOpd:  make([]pohonkinerja.TujuanOpdResponse, 0),
		Strategics: make([]pohonkinerja.StrategicOpdResponse, 0),
	}

	// Ambil data tujuan OPD
	tujuanOpds, err := service.tujuanOpdRepository.FindTujuanOpdByTahun(ctx, tx, kodeOpd, tahun, "RPJMD")
	if err != nil {
		log.Printf("Error getting tujuan OPD: %v", err)
		// Kembalikan response dengan array kosong jika terjadi error
		return response, nil
	}

	// Konversi tujuan OPD ke format response
	for _, tujuan := range tujuanOpds {
		indikators, err := service.tujuanOpdRepository.FindIndikatorByTujuanOpdId(ctx, tx, tujuan.Id)
		if err != nil {
			log.Printf("Error getting indikator for tujuan ID %d: %v", tujuan.Id, err)
			continue
		}

		var indikatorResponses []pohonkinerja.IndikatorTujuanResponse
		for _, indikator := range indikators {
			// Ambil target untuk setiap indikator dengan filter tahun
			targets, err := service.tujuanOpdRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id, tahun)
			if err != nil {
				log.Printf("Error getting targets for indikator ID %s: %v", indikator.Id, err)
				continue
			}

			var targetResponses []pohonkinerja.TargetTujuanResponse
			for _, target := range targets {
				targetResponses = append(targetResponses, pohonkinerja.TargetTujuanResponse{
					Tahun:  target.Tahun,
					Target: target.Target,
					Satuan: target.Satuan,
				})
			}

			indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorTujuanResponse{
				Indikator: indikator.Indikator,
				Target:    targetResponses,
			})
		}

		response.TujuanOpd = append(response.TujuanOpd, pohonkinerja.TujuanOpdResponse{
			Id:        tujuan.Id,
			KodeOpd:   tujuan.KodeOpd,
			Tujuan:    tujuan.Tujuan,
			Indikator: indikatorResponses,
		})
	}

	// Ambil data pohon kinerja
	pokins, err := service.pohonKinerjaOpdRepository.FindAll(ctx, tx, kodeOpd, tahun)
	if err != nil {
		// Kembalikan response dengan data yang sudah ada jika terjadi error
		return response, nil
	}

	// Jika tidak ada data pohon kinerja, kembalikan response dengan array kosong
	if len(pokins) == 0 {
		return response, nil
	}

	// Proses data pohon kinerja seperti sebelumnya
	pohonMap := make(map[int]map[int][]domain.PohonKinerja)
	pelaksanaMap := make(map[int][]pohonkinerja.PelaksanaOpdResponse)
	indikatorMap := make(map[int][]pohonkinerja.IndikatorResponse)

	// Kelompokkan data dan ambil data pelaksana & indikator
	maxLevel := 0
	for _, p := range pokins {
		if p.LevelPohon >= 4 {
			// Update max level jika ditemukan level yang lebih tinggi
			if p.LevelPohon > maxLevel {
				maxLevel = p.LevelPohon
			}

			// Inisialisasi map untuk level jika belum ada
			if pohonMap[p.LevelPohon] == nil {
				pohonMap[p.LevelPohon] = make(map[int][]domain.PohonKinerja)
			}

			p.NamaOpd = opd.NamaOpd
			pohonMap[p.LevelPohon][p.Parent] = append(
				pohonMap[p.LevelPohon][p.Parent],
				p,
			)

			// Ambil data pelaksana
			pelaksanaList, err := service.pohonKinerjaOpdRepository.FindPelaksanaPokin(ctx, tx, fmt.Sprint(p.Id))
			if err == nil {
				var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
				for _, pelaksana := range pelaksanaList {
					pegawaiPelaksana, err := service.pegawaiRepository.FindById(ctx, tx, pelaksana.PegawaiId)
					if err != nil {
						continue
					}
					pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
						Id:          pelaksana.Id,
						PegawaiId:   pegawaiPelaksana.Id,
						NamaPegawai: pegawaiPelaksana.NamaPegawai,
					})
				}
				pelaksanaMap[p.Id] = pelaksanaResponses
			}

			// Ambil data indikator dan target
			indikatorList, err := service.pohonKinerjaOpdRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(p.Id))
			if err == nil {
				var indikatorResponses []pohonkinerja.IndikatorResponse
				for _, indikator := range indikatorList {
					// Ambil target untuk setiap indikator
					targetList, err := service.pohonKinerjaOpdRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id)
					if err != nil {
						continue
					}

					var targetResponses []pohonkinerja.TargetResponse
					for _, target := range targetList {
						targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
							Id:              target.Id,
							IndikatorId:     target.IndikatorId,
							TargetIndikator: target.Target,
							SatuanIndikator: target.Satuan,
						})
					}

					indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
						Id:            indikator.Id,
						IdPokin:       indikator.PokinId,
						NamaIndikator: indikator.Indikator,
						Target:        targetResponses,
					})
				}
				indikatorMap[p.Id] = indikatorResponses
			}
		}
	}

	// Build response untuk strategic (level 4)
	if strategicList := pohonMap[4]; len(strategicList) > 0 {
		for _, strategicsByParent := range strategicList {
			// Ubah pengurutan di sini
			sort.Slice(strategicsByParent, func(i, j int) bool {
				// Prioritaskan status "pokin dari pemda"
				if strategicsByParent[i].Status == "pokin dari pemda" && strategicsByParent[j].Status != "pokin dari pemda" {
					return true
				}
				if strategicsByParent[i].Status != "pokin dari pemda" && strategicsByParent[j].Status == "pokin dari pemda" {
					return false
				}
				// Jika status sama, urutkan berdasarkan Id
				return strategicsByParent[i].Id < strategicsByParent[j].Id
			})

			for _, strategic := range strategicsByParent {
				strategicResp := service.buildStrategicResponse(ctx, tx, pohonMap, strategic, pelaksanaMap, indikatorMap)
				response.Strategics = append(response.Strategics, strategicResp)
			}
		}

		// Ubah pengurutan final di sini juga
		sort.Slice(response.Strategics, func(i, j int) bool {
			// Prioritaskan status "pokin dari pemda"
			if response.Strategics[i].Status == "pokin dari pemda" && response.Strategics[j].Status != "pokin dari pemda" {
				return true
			}
			if response.Strategics[i].Status != "pokin dari pemda" && response.Strategics[j].Status == "pokin dari pemda" {
				return false
			}
			// Jika status sama, urutkan berdasarkan Id
			return response.Strategics[i].Id < response.Strategics[j].Id
		})
	}

	return response, nil
}

func (service *PohonKinerjaOpdServiceImpl) FindStrategicNoParent(ctx context.Context, kodeOpd, tahun string) ([]pohonkinerja.StrategicOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi kode OPD
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
	if err != nil {
		return nil, errors.New("kode opd tidak ditemukan")
	}

	// Ambil data strategic dengan level pohon 4
	pokins, err := service.pohonKinerjaOpdRepository.FindStrategicNoParent(ctx, tx, 4, 0, kodeOpd, tahun)
	if err != nil {
		return nil, err
	}

	// Urutkan data berdasarkan ID
	sort.Slice(pokins, func(i, j int) bool {
		return pokins[i].Id < pokins[j].Id
	})

	// Konversi ke response format
	var strategics []pohonkinerja.StrategicOpdResponse
	for _, pokin := range pokins {
		strategic := pohonkinerja.StrategicOpdResponse{
			Id: pokin.Id,
			KodeOpd: opdmaster.OpdResponseForAll{
				KodeOpd: kodeOpd,
				NamaOpd: opd.NamaOpd,
			},
			Strategi:   pokin.NamaPohon,
			Keterangan: pokin.Keterangan,
		}
		strategics = append(strategics, strategic)
	}

	return strategics, nil
}

func (service *PohonKinerjaOpdServiceImpl) DeletePelaksana(ctx context.Context, pelaksanaId string) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)
	return service.pohonKinerjaOpdRepository.DeletePelaksanaPokin(ctx, tx, pelaksanaId)
}

// Tambahkan fungsi helper untuk membangun OperationalN response
func (service *PohonKinerjaOpdServiceImpl) buildOperationalNResponse(ctx context.Context, tx *sql.Tx, pohonMap map[int]map[int][]domain.PohonKinerja, operationalN domain.PohonKinerja, pelaksanaMap map[int][]pohonkinerja.PelaksanaOpdResponse, indikatorMap map[int][]pohonkinerja.IndikatorResponse) pohonkinerja.OperationalNOpdResponse {
	var keteranganCrosscutting *string
	if operationalN.KeteranganCrosscutting != nil && *operationalN.KeteranganCrosscutting != "" {
		keteranganCrosscutting = operationalN.KeteranganCrosscutting
	}
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, operationalN.KodeOpd)
	if err == nil {
		operationalN.NamaOpd = opd.NamaOpd
	}
	//review
	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, operationalN.Id)
	helper.PanicIfError(err)

	reviews, err := service.reviewRepository.FindByPohonKinerja(ctx, tx, operationalN.Id)
	var reviewResponses []pohonkinerja.ReviewResponse
	if err == nil {
		for _, review := range reviews {
			pegawai, err := service.pegawaiRepository.FindByNip(ctx, tx, reviews[0].CreatedBy)
			if err != nil {
				return pohonkinerja.OperationalNOpdResponse{}
			}
			reviewResponses = append(reviewResponses, pohonkinerja.ReviewResponse{
				Id:             review.Id,
				IdPohonKinerja: review.IdPohonKinerja,
				Review:         review.Review,
				Keterangan:     review.Keterangan,
				CreatedBy:      pegawai.NamaPegawai,
			})
		}
	}
	operationalNResp := pohonkinerja.OperationalNOpdResponse{
		Id:                     operationalN.Id,
		Parent:                 operationalN.Parent,
		Strategi:               operationalN.NamaPohon,
		JenisPohon:             operationalN.JenisPohon,
		LevelPohon:             operationalN.LevelPohon,
		Keterangan:             operationalN.Keterangan,
		KeteranganCrosscutting: keteranganCrosscutting,
		Status:                 operationalN.Status,
		IsActive:               operationalN.IsActive,
		KodeOpd: opdmaster.OpdResponseForAll{
			KodeOpd: operationalN.KodeOpd,
			NamaOpd: operationalN.NamaOpd,
		},
		Pelaksana:   pelaksanaMap[operationalN.Id],
		Indikator:   indikatorMap[operationalN.Id],
		Review:      reviewResponses,
		CountReview: countReview,
	}

	// Build child nodes secara rekursif
	nextLevel := operationalN.LevelPohon + 1
	if nextOperationalNList := pohonMap[nextLevel][operationalN.Id]; len(nextOperationalNList) > 0 {
		var childs []pohonkinerja.OperationalNOpdResponse
		sort.Slice(nextOperationalNList, func(i, j int) bool {
			return nextOperationalNList[i].Id < nextOperationalNList[j].Id
		})

		for _, nextOpN := range nextOperationalNList {
			childResp := service.buildOperationalNResponse(ctx, tx, pohonMap, nextOpN, pelaksanaMap, indikatorMap)
			childs = append(childs, childResp)
		}
		operationalNResp.Childs = childs
	}

	return operationalNResp
}

// Helper functions untuk membangun response
func (service *PohonKinerjaOpdServiceImpl) buildStrategicResponse(ctx context.Context, tx *sql.Tx, pohonMap map[int]map[int][]domain.PohonKinerja, strategic domain.PohonKinerja, pelaksanaMap map[int][]pohonkinerja.PelaksanaOpdResponse, indikatorMap map[int][]pohonkinerja.IndikatorResponse) pohonkinerja.StrategicOpdResponse {
	var keteranganCrosscutting *string
	if strategic.KeteranganCrosscutting != nil && *strategic.KeteranganCrosscutting != "" {
		keteranganCrosscutting = strategic.KeteranganCrosscutting
	}

	//review
	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, strategic.Id)
	helper.PanicIfError(err)

	reviews, err := service.reviewRepository.FindByPohonKinerja(ctx, tx, strategic.Id)
	var reviewResponses []pohonkinerja.ReviewResponse
	if err == nil {
		for _, review := range reviews {
			pegawai, err := service.pegawaiRepository.FindByNip(ctx, tx, reviews[0].CreatedBy)
			if err != nil {
				return pohonkinerja.StrategicOpdResponse{}
			}
			reviewResponses = append(reviewResponses, pohonkinerja.ReviewResponse{
				Id:             review.Id,
				IdPohonKinerja: review.IdPohonKinerja,
				Review:         review.Review,
				Keterangan:     review.Keterangan,
				CreatedBy:      pegawai.NamaPegawai,
			})
		}
	}

	strategicResp := pohonkinerja.StrategicOpdResponse{
		Id:                     strategic.Id,
		Parent:                 nil,
		Strategi:               strategic.NamaPohon,
		JenisPohon:             strategic.JenisPohon,
		LevelPohon:             strategic.LevelPohon,
		Keterangan:             strategic.Keterangan,
		KeteranganCrosscutting: keteranganCrosscutting,
		Status:                 strategic.Status,
		IsActive:               strategic.IsActive,
		KodeOpd: opdmaster.OpdResponseForAll{
			KodeOpd: strategic.KodeOpd,
			NamaOpd: strategic.NamaOpd,
		},
		Pelaksana:   pelaksanaMap[strategic.Id],
		Indikator:   indikatorMap[strategic.Id],
		Review:      reviewResponses,
		CountReview: countReview,
	}

	// Build tactical (level 5)
	if tacticalList := pohonMap[5][strategic.Id]; len(tacticalList) > 0 {
		var tacticals []pohonkinerja.TacticalOpdResponse
		sort.Slice(tacticalList, func(i, j int) bool {
			if tacticalList[i].Status == "pokin dari pemda" && tacticalList[j].Status != "pokin dari pemda" {
				return true
			}
			if tacticalList[i].Status != "pokin dari pemda" && tacticalList[j].Status == "pokin dari pemda" {
				return false
			}
			return tacticalList[i].Id < tacticalList[j].Id
		})

		for _, tactical := range tacticalList {
			tacticalResp := service.buildTacticalResponse(ctx, tx, pohonMap, tactical, pelaksanaMap, indikatorMap)
			tacticals = append(tacticals, tacticalResp)
		}
		strategicResp.Tacticals = tacticals
	}

	return strategicResp
}

func (service *PohonKinerjaOpdServiceImpl) buildTacticalResponse(ctx context.Context, tx *sql.Tx, pohonMap map[int]map[int][]domain.PohonKinerja, tactical domain.PohonKinerja, pelaksanaMap map[int][]pohonkinerja.PelaksanaOpdResponse, indikatorMap map[int][]pohonkinerja.IndikatorResponse) pohonkinerja.TacticalOpdResponse {
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, tactical.KodeOpd)
	if err == nil {
		tactical.NamaOpd = opd.NamaOpd
	}
	var keteranganCrosscutting *string
	if tactical.KeteranganCrosscutting != nil && *tactical.KeteranganCrosscutting != "" {
		keteranganCrosscutting = tactical.KeteranganCrosscutting
	}
	//review
	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, tactical.Id)
	helper.PanicIfError(err)
	reviews, err := service.reviewRepository.FindByPohonKinerja(ctx, tx, tactical.Id)
	var reviewResponses []pohonkinerja.ReviewResponse
	if err == nil {
		for _, review := range reviews {
			pegawai, err := service.pegawaiRepository.FindByNip(ctx, tx, reviews[0].CreatedBy)
			if err != nil {
				return pohonkinerja.TacticalOpdResponse{}
			}
			reviewResponses = append(reviewResponses, pohonkinerja.ReviewResponse{
				Id:             review.Id,
				IdPohonKinerja: review.IdPohonKinerja,
				Review:         review.Review,
				Keterangan:     review.Keterangan,
				CreatedBy:      pegawai.NamaPegawai,
			})
		}
	}
	tacticalResp := pohonkinerja.TacticalOpdResponse{
		Id:                     tactical.Id,
		Parent:                 tactical.Parent,
		Strategi:               tactical.NamaPohon,
		JenisPohon:             tactical.JenisPohon,
		LevelPohon:             tactical.LevelPohon,
		Keterangan:             tactical.Keterangan,
		KeteranganCrosscutting: keteranganCrosscutting,
		Status:                 tactical.Status,
		IsActive:               tactical.IsActive,
		KodeOpd: opdmaster.OpdResponseForAll{
			KodeOpd: tactical.KodeOpd,
			NamaOpd: tactical.NamaOpd,
		},
		Pelaksana:   pelaksanaMap[tactical.Id],
		Indikator:   indikatorMap[tactical.Id],
		Review:      reviewResponses,
		CountReview: countReview,
	}

	// Build operational (level 6)
	if operationalList := pohonMap[6][tactical.Id]; len(operationalList) > 0 {
		var operationals []pohonkinerja.OperationalOpdResponse
		sort.Slice(operationalList, func(i, j int) bool {
			if operationalList[i].Status == "pokin dari pemda" && operationalList[j].Status != "pokin dari pemda" {
				return true
			}
			if operationalList[i].Status != "pokin dari pemda" && operationalList[j].Status == "pokin dari pemda" {
				return false
			}
			return operationalList[i].Id < operationalList[j].Id
		})

		for _, operational := range operationalList {
			operationalResp := service.buildOperationalResponse(ctx, tx, pohonMap, operational, pelaksanaMap, indikatorMap)
			operationals = append(operationals, operationalResp)
		}
		tacticalResp.Operationals = operationals
	}

	return tacticalResp
}

func (service *PohonKinerjaOpdServiceImpl) buildOperationalResponse(ctx context.Context, tx *sql.Tx, pohonMap map[int]map[int][]domain.PohonKinerja, operational domain.PohonKinerja, pelaksanaMap map[int][]pohonkinerja.PelaksanaOpdResponse, indikatorMap map[int][]pohonkinerja.IndikatorResponse) pohonkinerja.OperationalOpdResponse {
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, operational.KodeOpd)
	if err == nil {
		operational.NamaOpd = opd.NamaOpd
	}
	var keteranganCrosscutting *string
	if operational.KeteranganCrosscutting != nil && *operational.KeteranganCrosscutting != "" {
		keteranganCrosscutting = operational.KeteranganCrosscutting
	}
	//review
	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, operational.Id)
	helper.PanicIfError(err)

	reviews, err := service.reviewRepository.FindByPohonKinerja(ctx, tx, operational.Id)
	var reviewResponses []pohonkinerja.ReviewResponse
	if err == nil {
		for _, review := range reviews {
			pegawai, err := service.pegawaiRepository.FindByNip(ctx, tx, reviews[0].CreatedBy)
			if err != nil {
				return pohonkinerja.OperationalOpdResponse{}
			}
			reviewResponses = append(reviewResponses, pohonkinerja.ReviewResponse{
				Id:             review.Id,
				IdPohonKinerja: review.IdPohonKinerja,
				Review:         review.Review,
				Keterangan:     review.Keterangan,
				CreatedBy:      pegawai.NamaPegawai,
			})
		}
	}
	operationalResp := pohonkinerja.OperationalOpdResponse{
		Id:                     operational.Id,
		Parent:                 operational.Parent,
		Strategi:               operational.NamaPohon,
		JenisPohon:             operational.JenisPohon,
		LevelPohon:             operational.LevelPohon,
		Keterangan:             operational.Keterangan,
		KeteranganCrosscutting: keteranganCrosscutting,
		Status:                 operational.Status,
		IsActive:               operational.IsActive,
		KodeOpd: opdmaster.OpdResponseForAll{
			KodeOpd: operational.KodeOpd,
			NamaOpd: operational.NamaOpd,
		},
		Pelaksana:   pelaksanaMap[operational.Id],
		Indikator:   indikatorMap[operational.Id],
		Review:      reviewResponses,
		CountReview: countReview,
	}

	// Build operational-n untuk level > 6
	nextLevel := operational.LevelPohon + 1
	if operationalNList := pohonMap[nextLevel][operational.Id]; len(operationalNList) > 0 {
		var childs []pohonkinerja.OperationalNOpdResponse
		// Ubah pengurutan untuk operational-n
		sort.Slice(operationalNList, func(i, j int) bool {
			// Prioritaskan status "pokin dari pemda"
			if operationalNList[i].Status == "pokin dari pemda" && operationalNList[j].Status != "pokin dari pemda" {
				return true
			}
			if operationalNList[i].Status != "pokin dari pemda" && operationalNList[j].Status == "pokin dari pemda" {
				return false
			}
			return operationalNList[i].Id < operationalNList[j].Id
		})

		for _, opN := range operationalNList {
			childResp := service.buildOperationalNResponse(ctx, tx, pohonMap, opN, pelaksanaMap, indikatorMap)
			childs = append(childs, childResp)
		}
		operationalResp.Childs = childs
	}

	return operationalResp
}

func (service *PohonKinerjaOpdServiceImpl) FindPokinByPelaksana(ctx context.Context, pegawaiId string, tahun string) ([]pohonkinerja.PohonKinerjaOpdResponse, error) {
	log.Printf("Memulai proses FindPokinByPelaksana untuk pegawai ID: %s", pegawaiId)

	tx, err := service.DB.Begin()
	if err != nil {
		log.Printf("Gagal memulai transaksi: %v", err)
		return nil, fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(tx)

	// Validasi pegawai
	pegawai, err := service.pegawaiRepository.FindById(ctx, tx, pegawaiId)
	if err != nil {
		log.Printf("Pegawai tidak ditemukan: %v", err)
		return nil, fmt.Errorf("pegawai tidak ditemukan: %v", err)
	}

	// Ambil data pohon kinerja berdasarkan pegawai
	pokinList, err := service.pohonKinerjaOpdRepository.FindPokinByPelaksana(ctx, tx, pegawaiId, tahun)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Tidak ada pohon kinerja untuk pegawai ID: %s", pegawaiId)
			return []pohonkinerja.PohonKinerjaOpdResponse{}, nil
		}
		log.Printf("Gagal mengambil data pohon kinerja: %v", err)
		return nil, fmt.Errorf("gagal mengambil data pohon kinerja: %v", err)
	}

	var responses []pohonkinerja.PohonKinerjaOpdResponse
	for _, pokin := range pokinList {
		// Ambil data OPD
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
		if err != nil {
			log.Printf("Gagal mengambil data OPD: %v", err)
			return nil, fmt.Errorf("gagal mengambil data OPD: %v", err)
		}

		indikators, err := service.pohonKinerjaOpdRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(pokin.Id))
		var indikatorResponses []pohonkinerja.IndikatorResponse
		if err == nil {
			for _, indikator := range indikators {
				var targetResponses []pohonkinerja.TargetResponse
				for _, target := range indikator.Target {
					targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
						Id:              target.Id,
						IndikatorId:     target.IndikatorId,
						TargetIndikator: target.Target,
						SatuanIndikator: target.Satuan,
					})
				}

				indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
					Id:            indikator.Id,
					IdPokin:       indikator.PokinId,
					NamaIndikator: indikator.Indikator,
					Target:        targetResponses,
				})
			}
		}

		// Buat response pelaksana hanya untuk pegawai yang bersangkutan
		pelaksanaResponse := pohonkinerja.PelaksanaOpdResponse{
			Id:             pokin.Pelaksana[0].Id, // Mengambil ID pelaksana pertama karena sudah difilter di repository
			PohonKinerjaId: fmt.Sprint(pokin.Id),
			PegawaiId:      pegawaiId,
			NamaPegawai:    pegawai.NamaPegawai,
		}

		responses = append(responses, pohonkinerja.PohonKinerjaOpdResponse{
			Id:         pokin.Id,
			Parent:     fmt.Sprint(pokin.Parent),
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    opd.KodeOpd,
			NamaOpd:    opd.NamaOpd,
			Keterangan: pokin.Keterangan,
			Tahun:      pokin.Tahun,
			Indikator:  indikatorResponses,
			Pelaksana:  []pohonkinerja.PelaksanaOpdResponse{pelaksanaResponse}, // Hanya menampilkan pelaksana yang sesuai
		})
	}

	log.Printf("Berhasil mengambil %d pohon kinerja untuk pegawai %s", len(responses), pegawai.NamaPegawai)
	return responses, nil
}

// func (service *PohonKinerjaOpdServiceImpl) buildCrosscuttingResponse(ctx context.Context, tx *sql.Tx, pokinId int, pelaksanaMap map[int][]pohonkinerja.PelaksanaOpdResponse, indikatorMap map[int][]pohonkinerja.IndikatorResponse) []pohonkinerja.CrosscuttingOpdResponse {
// 	// Ambil data crosscutting berdasarkan crosscutting_from
// 	crosscuttings, err := service.crosscuttingOpdRepository.FindAllCrosscutting(ctx, tx, pokinId)
// 	if err != nil {
// 		log.Printf("Error getting crosscutting data: %v", err)
// 		return nil
// 	}

// 	var crosscuttingResponses []pohonkinerja.CrosscuttingOpdResponse
// 	for _, crosscutting := range crosscuttings {
// 		// Filter status crosscutting yang akan ditampilkan
// 		if crosscutting.Status != "crosscutting_disetujui" &&
// 			crosscutting.Status != "crosscutting_disetujui_existing" {
// 			continue
// 		}

// 		// Ambil data OPD
// 		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, crosscutting.KodeOpd)
// 		if err != nil {
// 			log.Printf("Gagal mengambil data OPD: %v", err)
// 			continue
// 		}

// 		// Ambil data indikator untuk crosscutting
// 		indikatorList, err := service.pohonKinerjaOpdRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(crosscutting.Id))
// 		if err == nil {
// 			var indikatorResponses []pohonkinerja.IndikatorResponse
// 			for _, indikator := range indikatorList {
// 				// Ambil target untuk setiap indikator
// 				targetList, err := service.pohonKinerjaOpdRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id)
// 				if err != nil {
// 					continue
// 				}

// 				var targetResponses []pohonkinerja.TargetResponse
// 				for _, target := range targetList {
// 					targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
// 						Id:              target.Id,
// 						IndikatorId:     target.IndikatorId,
// 						TargetIndikator: target.Target,
// 						SatuanIndikator: target.Satuan,
// 					})
// 				}

// 				indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
// 					Id:            indikator.Id,
// 					IdPokin:       fmt.Sprint(crosscutting.Id),
// 					NamaIndikator: indikator.Indikator,
// 					Target:        targetResponses,
// 				})
// 			}
// 			indikatorMap[crosscutting.Id] = indikatorResponses
// 		}

// 		// Ambil data pelaksana untuk crosscutting
// 		pelaksanaList, err := service.pohonKinerjaOpdRepository.FindPelaksanaPokin(ctx, tx, fmt.Sprint(crosscutting.Id))
// 		if err == nil {
// 			var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
// 			for _, pelaksana := range pelaksanaList {
// 				pegawai, err := service.pegawaiRepository.FindById(ctx, tx, pelaksana.PegawaiId)
// 				if err != nil {
// 					continue
// 				}
// 				pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
// 					Id:             pelaksana.Id,
// 					PohonKinerjaId: fmt.Sprint(crosscutting.Id),
// 					PegawaiId:      pelaksana.PegawaiId,
// 					NamaPegawai:    pegawai.NamaPegawai,
// 				})
// 			}
// 			pelaksanaMap[crosscutting.Id] = pelaksanaResponses
// 		}

// 		// Jika status disetujui_existing, ambil data pohon kinerja yang di-crosscut
// 		var namaPohon string
// 		var jenisPohon string
// 		var levelPohon int

// 		if crosscutting.Status == "crosscutting_disetujui_existing" && crosscutting.CrosscuttingTo != 0 {
// 			pokinExisting, err := service.pohonKinerjaOpdRepository.FindById(ctx, tx, crosscutting.CrosscuttingTo)
// 			if err == nil {
// 				namaPohon = pokinExisting.NamaPohon
// 				jenisPohon = pokinExisting.JenisPohon
// 				levelPohon = pokinExisting.LevelPohon
// 			}
// 		} else {
// 			namaPohon = crosscutting.NamaPohon
// 			jenisPohon = crosscutting.JenisPohon
// 			levelPohon = crosscutting.LevelPohon
// 		}

// 		crosscuttingResp := pohonkinerja.CrosscuttingOpdResponse{
// 			Id:         crosscutting.Id,
// 			Parent:     pokinId,
// 			NamaPohon:  namaPohon,
// 			JenisPohon: jenisPohon,
// 			LevelPohon: levelPohon,
// 			Keterangan: crosscutting.Keterangan,
// 			Status:     crosscutting.Status,
// 			KodeOpd:    crosscutting.KodeOpd,
// 			NamaOpd:    opd.NamaOpd,
// 			Tahun:      crosscutting.Tahun,
// 			Pelaksana:  pelaksanaMap[crosscutting.Id],
// 			Indikator:  indikatorMap[crosscutting.Id],
// 		}

// 		crosscuttingResponses = append(crosscuttingResponses, crosscuttingResp)
// 	}

// 	return crosscuttingResponses
// }

func (service *PohonKinerjaOpdServiceImpl) DeletePokinPemdaInOpd(ctx context.Context, id int) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(tx)

	// 1. Cek apakah pohon kinerja dengan ID tersebut ada
	_, err = service.pohonKinerjaOpdRepository.FindById(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("pohon kinerja tidak ditemukan: %v", err)
	}

	// 2. Cek apakah ini adalah pohon kinerja yang di-clone dan dapatkan ID aslinya
	cloneFrom, err := service.pohonKinerjaOpdRepository.CheckCloneFrom(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("gagal memeriksa clone_from: %v", err)
	}

	if cloneFrom == 0 {
		return fmt.Errorf("pohon kinerja ini bukan merupakan hasil clone dari pemda")
	}

	// 3. Hapus data clone saat ini
	err = service.pohonKinerjaOpdRepository.Delete(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus pohon kinerja clone: %v", err)
	}

	// 4. Update status pohon kinerja asli (yang di-clone) menjadi "ditolak"
	err = service.pohonKinerjaOpdRepository.UpdatePokinStatusFromApproved(ctx, tx, cloneFrom)
	if err != nil {
		return fmt.Errorf("gagal mengupdate status pohon kinerja asli: %v", err)
	}

	// 5. Dapatkan dan update status semua child dari pohon kinerja asli
	originalHierarchy, err := service.pohonKinerjaOpdRepository.FindPokinAdminByIdHierarki(ctx, tx, cloneFrom)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan hierarki pohon kinerja asli: %v", err)
	}

	// 6. Update status untuk semua child dari pohon kinerja asli
	for _, originalPokin := range originalHierarchy {
		if originalPokin.Id != cloneFrom { // Skip pohon kinerja utama karena sudah diupdate
			err = service.pohonKinerjaOpdRepository.UpdatePokinStatusFromApproved(ctx, tx, originalPokin.Id)
			if err != nil {
				log.Printf("Warning: gagal mengupdate status child pohon kinerja asli ID %d: %v", originalPokin.Id, err)
				continue // Lanjutkan ke child berikutnya meskipun ada error
			}
		}
	}

	return nil
}

func (service *PohonKinerjaOpdServiceImpl) UpdateParent(ctx context.Context, pohonKinerja pohonkinerja.PohonKinerjaUpdateRequest) (pohonkinerja.PohonKinerjaOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(tx)

	pokin := domain.PohonKinerja{
		Id:     pohonKinerja.Id,
		Parent: pohonKinerja.Parent,
	}

	pokin, err = service.pohonKinerjaOpdRepository.UpdateParent(ctx, tx, pokin)
	if err != nil {
		return pohonkinerja.PohonKinerjaOpdResponse{}, err
	}

	return pohonkinerja.PohonKinerjaOpdResponse{
		Id:     pokin.Id,
		Parent: fmt.Sprint(pokin.Parent),
	}, nil
}

func (service *PohonKinerjaOpdServiceImpl) FindidPokinWithAllTema(ctx context.Context, id int) (pohonkinerja.PohonKinerjaAdminResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	pokins, err := service.pohonKinerjaOpdRepository.FindidPokinWithAllTema(ctx, tx, id)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponse{}, err
	}

	// Temukan node target dan level parentnya
	var targetPokin domain.PohonKinerja
	var parentLevel int
	var strategicId int // Untuk menyimpan ID dari node Strategic

	for _, pokin := range pokins {
		if pokin.Id == id {
			targetPokin = pokin
			// Jika level 4, cari level parentnya
			if pokin.LevelPohon == 4 {
				for _, p := range pokins {
					if p.Id == pokin.Parent {
						parentLevel = p.LevelPohon
						break
					}
				}
				strategicId = pokin.Id
			} else if pokin.LevelPohon == 5 || pokin.LevelPohon == 6 {
				// Untuk level 5 dan 6, cari node Strategic (level 4) di ancestors
				for _, p := range pokins {
					if p.LevelPohon == 4 {
						strategicId = p.Id
						// Cari level parent dari Strategic
						for _, grandParent := range pokins {
							if grandParent.Id == p.Parent {
								parentLevel = grandParent.LevelPohon
								break
							}
						}
						break
					}
				}
			}
			break
		}
	}

	// Validasi level
	if targetPokin.LevelPohon < 4 || targetPokin.LevelPohon > 6 {
		return pohonkinerja.PohonKinerjaAdminResponse{}, fmt.Errorf("ID harus merujuk ke level Strategic (4), Tactical (5), atau Operational (6)")
	}

	// Helper functions (sama seperti sebelumnya)
	createIndikatorResponse := func(pokin domain.PohonKinerja) []pohonkinerja.IndikatorResponse {
		var indikators []pohonkinerja.IndikatorResponse
		for _, ind := range pokin.Indikator {
			var targets []pohonkinerja.TargetResponse
			for _, t := range ind.Target {
				targets = append(targets, pohonkinerja.TargetResponse{
					Id:              t.Id,
					IndikatorId:     t.IndikatorId,
					TargetIndikator: t.Target,
					SatuanIndikator: t.Satuan,
				})
			}
			indikators = append(indikators, pohonkinerja.IndikatorResponse{
				Id:            ind.Id,
				IdPokin:       ind.PokinId,
				NamaIndikator: ind.Indikator,
				Target:        targets,
			})
		}
		return indikators
	}

	// Buat responses untuk setiap level yang diperlukan
	var strategicResp pohonkinerja.StrategicResponse
	var tacticalResp *pohonkinerja.TacticalResponse
	var operationalResp *pohonkinerja.OperationalResponse

	// Bangun response berdasarkan level target
	for _, pokin := range pokins {
		if pokin.Id == strategicId {
			// Buat Strategic Response
			strategicResp = pohonkinerja.StrategicResponse{
				Id:         pokin.Id,
				Parent:     pokin.Parent,
				Strategi:   pokin.NamaPohon,
				JenisPohon: pokin.JenisPohon,
				LevelPohon: pokin.LevelPohon,
				Keterangan: pokin.Keterangan,
				Status:     pokin.Status,
				Indikators: createIndikatorResponse(pokin),
				Childs:     []interface{}{},
			}
		} else if pokin.LevelPohon == 5 && targetPokin.LevelPohon >= 5 {
			// Buat Tactical Response
			tacticalResp = &pohonkinerja.TacticalResponse{
				Id:         pokin.Id,
				Parent:     pokin.Parent,
				Strategi:   pokin.NamaPohon,
				JenisPohon: pokin.JenisPohon,
				LevelPohon: pokin.LevelPohon,
				Keterangan: &pokin.Keterangan,
				Status:     pokin.Status,
				Indikators: createIndikatorResponse(pokin),
				Childs:     []interface{}{},
			}
			strategicResp.Childs = append(strategicResp.Childs, tacticalResp)
		} else if pokin.LevelPohon == 6 && targetPokin.LevelPohon == 6 {
			// Buat Operational Response
			operationalResp = &pohonkinerja.OperationalResponse{
				Id:         pokin.Id,
				Parent:     pokin.Parent,
				Strategi:   pokin.NamaPohon,
				JenisPohon: pokin.JenisPohon,
				LevelPohon: pokin.LevelPohon,
				Keterangan: &pokin.Keterangan,
				Status:     pokin.Status,
				Indikators: createIndikatorResponse(pokin),
				Childs:     []interface{}{},
			}
			if tacticalResp != nil {
				tacticalResp.Childs = append(tacticalResp.Childs, operationalResp)
			}
		}
	}

	// Bangun hierarki dari Tematik ke bawah
	var tematikResp pohonkinerja.TematikResponse
	for _, pokin := range pokins {
		if pokin.LevelPohon == 0 { // Tematik
			parentInt := pokin.Parent
			tematikResp = pohonkinerja.TematikResponse{
				Id:         pokin.Id,
				Parent:     &parentInt,
				Tema:       pokin.NamaPohon,
				JenisPohon: pokin.JenisPohon,
				LevelPohon: pokin.LevelPohon,
				Keterangan: pokin.Keterangan,
				Indikators: createIndikatorResponse(pokin),
				Child:      []interface{}{},
			}

			if parentLevel == 0 {
				tematikResp.Child = append(tematikResp.Child, strategicResp)
			}
		} else if pokin.LevelPohon <= parentLevel {
			switch pokin.LevelPohon {
			case 1: // Subtematik
				subtematikResp := pohonkinerja.SubtematikResponse{
					Id:         pokin.Id,
					Parent:     pokin.Parent,
					Tema:       pokin.NamaPohon,
					JenisPohon: pokin.JenisPohon,
					LevelPohon: pokin.LevelPohon,
					Keterangan: pokin.Keterangan,
					Indikators: createIndikatorResponse(pokin),
					Child:      []interface{}{},
				}
				if parentLevel == 1 {
					subtematikResp.Child = append(subtematikResp.Child, strategicResp)
				}
				tematikResp.Child = append(tematikResp.Child, subtematikResp)

			case 2: // SubSubTematik
				subsubtematikResp := pohonkinerja.SubSubTematikResponse{
					Id:         pokin.Id,
					Parent:     pokin.Parent,
					Tema:       pokin.NamaPohon,
					JenisPohon: pokin.JenisPohon,
					LevelPohon: pokin.LevelPohon,
					Keterangan: pokin.Keterangan,
					Indikators: createIndikatorResponse(pokin),
					Child:      []interface{}{},
				}
				if parentLevel == 2 {
					subsubtematikResp.Child = append(subsubtematikResp.Child, strategicResp)
				}
				for i := range tematikResp.Child {
					if sub, ok := tematikResp.Child[i].(pohonkinerja.SubtematikResponse); ok && sub.Id == pokin.Parent {
						sub.Child = append(sub.Child, subsubtematikResp)
						tematikResp.Child[i] = sub
					}
				}

			case 3: // SuperSubTematik
				supersubtematikResp := pohonkinerja.SuperSubTematikResponse{
					Id:         pokin.Id,
					Parent:     pokin.Parent,
					Tema:       pokin.NamaPohon,
					JenisPohon: pokin.JenisPohon,
					LevelPohon: pokin.LevelPohon,
					Keterangan: pokin.Keterangan,
					Indikators: createIndikatorResponse(pokin),
					Childs:     []interface{}{},
				}
				if parentLevel == 3 {
					supersubtematikResp.Childs = append(supersubtematikResp.Childs, strategicResp)
				}
				// Tambahkan ke parent yang sesuai
				for i := range tematikResp.Child {
					if sub, ok := tematikResp.Child[i].(pohonkinerja.SubtematikResponse); ok {
						for j := range sub.Child {
							if subsub, ok := sub.Child[j].(pohonkinerja.SubSubTematikResponse); ok && subsub.Id == pokin.Parent {
								subsub.Child = append(subsub.Child, supersubtematikResp)
								sub.Child[j] = subsub
								tematikResp.Child[i] = sub
							}
						}
					}
				}
			}
		}
	}

	response := pohonkinerja.PohonKinerjaAdminResponse{
		Tahun:   targetPokin.Tahun,
		Tematik: []pohonkinerja.TematikResponse{tematikResp},
	}

	return response, nil
}
func (service *PohonKinerjaOpdServiceImpl) CloneByKodeOpdAndTahun(ctx context.Context, request pohonkinerja.PohonKinerjaCloneRequest) error {
	// Validasi request
	err := service.Validate.Struct(request)
	if err != nil {
		return err
	}

	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	// Lakukan cloning
	err = service.pohonKinerjaOpdRepository.ClonePokinOpd(ctx, tx, request.KodeOpd, request.TahunSumber, request.TahunTujuan)
	if err != nil {
		return fmt.Errorf("gagal melakukan cloning: %v", err)
	}

	return nil
}

func (service *PohonKinerjaOpdServiceImpl) CheckPokinExistsByTahun(ctx context.Context, kodeOpd string, tahun string) (bool, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return false, fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer tx.Rollback()

	exists := service.pohonKinerjaOpdRepository.IsExistsByTahun(ctx, tx, kodeOpd, tahun)

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("gagal melakukan commit transaksi: %v", err)
	}

	return exists, nil
}

func (service *PohonKinerjaOpdServiceImpl) CountPokinPemda(ctx context.Context, kodeOpd, tahun string) (pohonkinerja.CountPokinPemdaResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.CountPokinPemdaResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD
	opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
	if err != nil {
		return pohonkinerja.CountPokinPemdaResponse{}, errors.New("kode opd tidak ditemukan")
	}

	// Hitung jumlah pokin pemda per level
	countByLevel, err := service.pohonKinerjaOpdRepository.CountPokinPemdaByLevel(ctx, tx, kodeOpd, tahun)
	if err != nil {
		return pohonkinerja.CountPokinPemdaResponse{}, err
	}

	// Siapkan response
	response := pohonkinerja.CountPokinPemdaResponse{
		KodeOpd: kodeOpd,
		NamaOpd: opd.NamaOpd,
		Tahun:   tahun,
	}

	// Definisikan level yang harus selalu ada dalam response
	requiredLevels := []struct {
		Level      int
		JenisPohon string
	}{
		{4, "Strategic"},
		{5, "Tactical"},
		{6, "Operational"},
	}

	var details []pohonkinerja.LevelDetail
	totalPemda := 0

	// Tambahkan semua level yang required
	for _, req := range requiredLevels {
		count := countByLevel[req.Level] // Akan return 0 jika tidak ada data
		details = append(details, pohonkinerja.LevelDetail{
			Level:       req.Level,
			JenisPohon:  req.JenisPohon,
			JumlahPemda: count,
		})
		totalPemda += count
	}

	// Tambahkan level operational-n (> 6) jika ada
	for level, count := range countByLevel {
		if level > 6 {
			details = append(details, pohonkinerja.LevelDetail{
				Level:       level,
				JenisPohon:  fmt.Sprintf("Operational-%d", level-6),
				JumlahPemda: count,
			})
			totalPemda += count
		}
	}

	// Urutkan detail berdasarkan level
	sort.Slice(details, func(i, j int) bool {
		return details[i].Level < details[j].Level
	})

	response.TotalPemda = totalPemda
	response.DetailLevel = details

	return response, nil
}
