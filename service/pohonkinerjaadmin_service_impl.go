package service

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"ekak_kabupaten_madiun/model/web/opdmaster"
	"ekak_kabupaten_madiun/model/web/pohonkinerja"
	"ekak_kabupaten_madiun/repository"

	"log"

	"fmt"

	"sort"

	"errors"

	"github.com/google/uuid"
)

type PohonKinerjaAdminServiceImpl struct {
	pohonKinerjaRepository repository.PohonKinerjaRepository
	opdRepository          repository.OpdRepository
	pegawaiRepository      repository.PegawaiRepository
	reviewRepository       repository.ReviewRepository
	DB                     *sql.DB
}

func NewPohonKinerjaAdminServiceImpl(pohonKinerjaRepository repository.PohonKinerjaRepository, opdRepository repository.OpdRepository, DB *sql.DB, pegawaiRepository repository.PegawaiRepository, reviewRepository repository.ReviewRepository) *PohonKinerjaAdminServiceImpl {
	return &PohonKinerjaAdminServiceImpl{
		pohonKinerjaRepository: pohonKinerjaRepository,
		opdRepository:          opdRepository,
		pegawaiRepository:      pegawaiRepository,
		DB:                     DB,
		reviewRepository:       reviewRepository,
	}
}

func (service *PohonKinerjaAdminServiceImpl) Create(ctx context.Context, request pohonkinerja.PohonKinerjaAdminCreateRequest) (pohonkinerja.PohonKinerjaAdminResponseData, error) {
	log.Printf("Memulai proses pembuatan PohonKinerja untuk tahun: %s", request.Tahun)

	tx, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error memulai transaksi: %v", err)
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Persiapkan data pelaksana
	var pelaksanaList []domain.PelaksanaPokin
	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse

	for _, pelaksanaReq := range request.Pelaksana {
		// Generate ID untuk pelaksana
		pelaksanaId := fmt.Sprintf("PLKS-%s", uuid.New().String()[:8])

		// Validasi pegawai
		pegawai, err := service.pegawaiRepository.FindById(ctx, tx, pelaksanaReq.PegawaiId)
		if err != nil {
			log.Printf("Error: pegawai dengan ID %s tidak ditemukan", pelaksanaReq.PegawaiId)
			return pohonkinerja.PohonKinerjaAdminResponseData{}, fmt.Errorf("pegawai tidak ditemukan: %v", err)
		}

		pelaksana := domain.PelaksanaPokin{
			Id:        pelaksanaId,
			PegawaiId: pelaksanaReq.PegawaiId,
		}
		pelaksanaList = append(pelaksanaList, pelaksana)

		pelaksanaResponse := pohonkinerja.PelaksanaOpdResponse{
			Id:          pelaksanaId,
			PegawaiId:   pegawai.Id,
			NamaPegawai: pegawai.NamaPegawai,
		}
		pelaksanaResponses = append(pelaksanaResponses, pelaksanaResponse)
	}

	// Logging persiapan indikator
	log.Printf("Mempersiapkan %d indikator", len(request.Indikator))

	// Persiapkan data indikator dan target
	var indikators []domain.Indikator
	for _, ind := range request.Indikator {
		indikatorId := "IND-POKIN-" + uuid.New().String()

		var targets []domain.Target
		for _, t := range ind.Target {
			targetId := "TRGT-IND-POKIN-" + uuid.New().String()
			target := domain.Target{
				Id:          targetId,
				IndikatorId: indikatorId,
				Target:      t.Target,
				Satuan:      t.Satuan,
				Tahun:       request.Tahun,
			}
			targets = append(targets, target)
		}

		indikator := domain.Indikator{
			Id:        indikatorId,
			Indikator: ind.NamaIndikator,
			Tahun:     request.Tahun,
			Target:    targets,
		}
		indikators = append(indikators, indikator)
	}

	pohonKinerja := domain.PohonKinerja{
		Parent:     request.Parent,
		NamaPohon:  request.NamaPohon,
		JenisPohon: request.JenisPohon,
		LevelPohon: request.LevelPohon,
		KodeOpd:    helper.EmptyStringIfNull(request.KodeOpd),
		Keterangan: request.Keterangan,
		Tahun:      request.Tahun,
		Status:     request.Status,
		Pelaksana:  pelaksanaList,
		Indikator:  indikators,
	}

	log.Printf("Menyimpan PohonKinerja dengan NamaPohon: %s, LevelPohon: %d", request.NamaPohon, request.LevelPohon)
	result, err := service.pohonKinerjaRepository.CreatePokinAdmin(ctx, tx, pohonKinerja)
	if err != nil {
		log.Printf("Error saat menyimpan PohonKinerja: %v", err)
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	log.Printf("Berhasil membuat PohonKinerja dengan ID: %d", result.Id)

	// Konversi indikator domain ke IndikatorResponse
	var indikatorResponses []pohonkinerja.IndikatorResponse
	for _, ind := range result.Indikator {
		var targetResponses []pohonkinerja.TargetResponse
		for _, t := range ind.Target {
			targetResponse := pohonkinerja.TargetResponse{
				Id:              t.Id,
				IndikatorId:     t.IndikatorId,
				TargetIndikator: t.Target,
				SatuanIndikator: t.Satuan,
			}
			targetResponses = append(targetResponses, targetResponse)
		}

		indikatorResponse := pohonkinerja.IndikatorResponse{
			Id:            ind.Id,
			NamaIndikator: ind.Indikator,
			Target:        targetResponses,
		}
		indikatorResponses = append(indikatorResponses, indikatorResponse)
	}

	var namaOpd string
	if request.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, request.KodeOpd)
		if err == nil {
			namaOpd = opd.NamaOpd
		}
	}

	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, result.Id)
	helper.PanicIfError(err)

	response := pohonkinerja.PohonKinerjaAdminResponseData{
		Id:         result.Id,
		Parent:     result.Parent,
		NamaPohon:  result.NamaPohon,
		JenisPohon: result.JenisPohon,
		LevelPohon: result.LevelPohon,
		PerangkatDaerah: &opdmaster.OpdResponseForAll{
			KodeOpd: result.KodeOpd,
			NamaOpd: namaOpd,
		},
		Keterangan:  result.Keterangan,
		Tahun:       result.Tahun,
		Status:      result.Status,
		IsActive:    true,
		CountReview: countReview,
		Pelaksana:   pelaksanaResponses,
		Indikators:  indikatorResponses,
	}

	log.Printf("Proses pembuatan PohonKinerja selesai")
	return response, nil
}

func (service *PohonKinerjaAdminServiceImpl) Update(ctx context.Context, request pohonkinerja.PohonKinerjaAdminUpdateRequest) (pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Cek apakah data exists
	existingPokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, request.Id)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Persiapkan data yang akan diupdate
	var pokinsToUpdate []domain.PohonKinerja

	// Tambahkan pokin yang sedang diupdate
	pokinsToUpdate = append(pokinsToUpdate, existingPokin)

	// Jika CloneFrom = 0, cari pokin lain yang memiliki CloneFrom = Id yang sedang diupdate
	if existingPokin.CloneFrom == 0 {
		relatedPokins, err := service.pohonKinerjaRepository.FindPokinByCloneFrom(ctx, tx, request.Id)
		if err != nil {
			return pohonkinerja.PohonKinerjaAdminResponseData{}, err
		}
		pokinsToUpdate = append(pokinsToUpdate, relatedPokins...)
	}

	// Persiapkan data pelaksana
	var pelaksanaList []domain.PelaksanaPokin
	for _, p := range request.Pelaksana {
		pelaksanaId := "PLKS-" + uuid.New().String()[:8]
		pelaksana := domain.PelaksanaPokin{
			Id:        pelaksanaId,
			PegawaiId: p.PegawaiId,
		}
		pelaksanaList = append(pelaksanaList, pelaksana)
	}

	// Persiapkan data indikator dan target untuk pokin asli
	var indikators []domain.Indikator
	for _, ind := range request.Indikator {
		var indikatorId string
		if ind.Id == "" {
			indikatorId = "IND-POKIN-" + uuid.New().String()[:8]
		} else {
			indikatorId = ind.Id
		}

		var targets []domain.Target
		for _, t := range ind.Target {
			var targetId string
			if t.Id == "" {
				targetId = "TRGT-IND-POKIN-" + uuid.New().String()[:8]
			} else {
				targetId = t.Id
			}

			target := domain.Target{
				Id:          targetId,
				IndikatorId: indikatorId,
				Target:      t.Target,
				Satuan:      t.Satuan,
				Tahun:       request.Tahun,
			}
			targets = append(targets, target)
		}

		indikator := domain.Indikator{
			Id:        indikatorId,
			Indikator: ind.NamaIndikator,
			Tahun:     request.Tahun,
			Target:    targets,
		}
		indikators = append(indikators, indikator)
	}

	// Update semua pokin yang terkait
	var updatedPokin domain.PohonKinerja
	for _, pokin := range pokinsToUpdate {
		var pokinIndikators []domain.Indikator

		if pokin.Id == request.Id {
			// Untuk pokin asli, gunakan indikator dari request
			pokinIndikators = indikators
		} else {
			// Untuk pokin yang diclone
			existingIndikators, err := service.pohonKinerjaRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(pokin.Id))
			if err != nil {
				return pohonkinerja.PohonKinerjaAdminResponseData{}, err
			}

			// Proses setiap indikator dari pokin asli
			for _, originalInd := range indikators {
				var clonedIndikator domain.Indikator

				// Cari indikator yang sudah ada dengan clone_from yang sesuai
				var existingInd *domain.Indikator
				for _, ei := range existingIndikators {
					if ei.CloneFrom == originalInd.Id {
						existingInd = &ei
						break
					}
				}

				if existingInd != nil {
					// Gunakan ID yang sudah ada untuk indikator yang di-clone
					clonedIndikator = *existingInd
					clonedIndikator.Indikator = originalInd.Indikator
					clonedIndikator.Tahun = originalInd.Tahun
				} else {
					// Buat indikator baru untuk clone
					clonedIndikator = domain.Indikator{
						Id:        "IND-POKIN-" + uuid.New().String()[:8],
						Indikator: originalInd.Indikator,
						Tahun:     originalInd.Tahun,
						CloneFrom: originalInd.Id,
					}
				}

				// Proses target untuk indikator yang di-clone
				var clonedTargets []domain.Target
				for _, originalTarget := range originalInd.Target {
					var clonedTarget domain.Target

					// Cari target yang sudah ada
					var existingTarget *domain.Target
					if existingInd != nil {
						for _, et := range existingInd.Target {
							if et.CloneFrom == originalTarget.Id {
								existingTarget = &et
								break
							}
						}
					}

					if existingTarget != nil {
						// Gunakan ID yang sudah ada untuk target yang di-clone
						clonedTarget = *existingTarget
						clonedTarget.Target = originalTarget.Target
						clonedTarget.Satuan = originalTarget.Satuan
						clonedTarget.Tahun = originalTarget.Tahun
					} else {
						// Buat target baru untuk clone
						clonedTarget = domain.Target{
							Id:          "TRGT-IND-POKIN-" + uuid.New().String()[:8],
							IndikatorId: clonedIndikator.Id,
							Target:      originalTarget.Target,
							Satuan:      originalTarget.Satuan,
							Tahun:       originalTarget.Tahun,
							CloneFrom:   originalTarget.Id,
						}
					}
					clonedTargets = append(clonedTargets, clonedTarget)
				}

				clonedIndikator.Target = clonedTargets
				pokinIndikators = append(pokinIndikators, clonedIndikator)
			}
		}

		pohonKinerja := domain.PohonKinerja{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  request.NamaPohon,
			JenisPohon: request.JenisPohon,
			LevelPohon: request.LevelPohon,
			KodeOpd:    helper.EmptyStringIfNull(request.KodeOpd),
			Keterangan: request.Keterangan,
			Tahun:      request.Tahun,
			Status:     pokin.Status,
			CloneFrom:  pokin.CloneFrom,
			Pelaksana:  pelaksanaList,
			Indikator:  pokinIndikators,
		}

		result, err := service.pohonKinerjaRepository.UpdatePokinAdmin(ctx, tx, pohonKinerja)
		if err != nil {
			return pohonkinerja.PohonKinerjaAdminResponseData{}, err
		}

		if pokin.Id == request.Id {
			updatedPokin = result
		}
	}

	// Konversi pelaksana domain ke PelaksanaResponse
	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
	for _, p := range updatedPokin.Pelaksana {
		// Ambil data pegawai
		pegawai, err := service.pegawaiRepository.FindById(ctx, tx, p.PegawaiId)
		if err != nil {
			continue // Skip jika pegawai tidak ditemukan
		}

		pelaksanaResponse := pohonkinerja.PelaksanaOpdResponse{
			Id:          p.Id,
			PegawaiId:   pegawai.Id,
			NamaPegawai: pegawai.NamaPegawai,
		}
		pelaksanaResponses = append(pelaksanaResponses, pelaksanaResponse)
	}

	// Konversi indikator domain ke IndikatorResponse
	var indikatorResponses []pohonkinerja.IndikatorResponse
	for _, ind := range updatedPokin.Indikator {
		var targetResponses []pohonkinerja.TargetResponse
		for _, t := range ind.Target {
			targetResponse := pohonkinerja.TargetResponse{
				Id:              t.Id,
				IndikatorId:     t.IndikatorId,
				TargetIndikator: t.Target,
				SatuanIndikator: t.Satuan,
			}
			targetResponses = append(targetResponses, targetResponse)
		}

		indikatorResponse := pohonkinerja.IndikatorResponse{
			Id:            ind.Id,
			NamaIndikator: ind.Indikator,
			Target:        targetResponses,
		}
		indikatorResponses = append(indikatorResponses, indikatorResponse)
	}

	var namaOpd string
	if request.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, request.KodeOpd)
		if err == nil {
			namaOpd = opd.NamaOpd
		}
	}

	findidpokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, updatedPokin.Id)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, updatedPokin.Id)
	helper.PanicIfError(err)

	response := pohonkinerja.PohonKinerjaAdminResponseData{
		Id:         updatedPokin.Id,
		Parent:     updatedPokin.Parent,
		NamaPohon:  updatedPokin.NamaPohon,
		JenisPohon: updatedPokin.JenisPohon,
		LevelPohon: updatedPokin.LevelPohon,
		PerangkatDaerah: &opdmaster.OpdResponseForAll{
			KodeOpd: updatedPokin.KodeOpd,
			NamaOpd: namaOpd,
		},
		Keterangan:  updatedPokin.Keterangan,
		Tahun:       updatedPokin.Tahun,
		Status:      updatedPokin.Status,
		CountReview: countReview,
		Pelaksana:   pelaksanaResponses,
		Indikators:  indikatorResponses,
		IsActive:    findidpokin.IsActive,
	}

	return response, nil
}

func (service *PohonKinerjaAdminServiceImpl) Delete(ctx context.Context, id int) error {
	// Mulai transaksi
	tx, err := service.DB.Begin()
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(tx)

	// Cek apakah data exists sebelum dihapus
	pokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("data tidak ditemukan: %v", err)
	}

	// Validasi level pohon: hanya validasi batas bawah
	if pokin.LevelPohon < 0 {
		return fmt.Errorf("level pohon kinerja tidak valid")
	}

	// Cek apakah data adalah hasil clone
	cloneFrom, err := service.pohonKinerjaRepository.CheckCloneFrom(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("gagal memeriksa status clone: %v", err)
	}

	// tanpa mempengaruhi data asli (clone_from)
	if cloneFrom != 0 {
		err = service.pohonKinerjaRepository.DeleteClonedPokinHierarchy(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("gagal menghapus data clone: %v", err)
		}
		return nil
	}

	// Jika data adalah asli (clone_from = 0), hapus data beserta semua yang terkait
	err = service.pohonKinerjaRepository.DeletePokinAdmin(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus data: %v", err)
	}

	return nil
}

func (service *PohonKinerjaAdminServiceImpl) FindById(ctx context.Context, id int) (pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}
	defer helper.CommitOrRollback(tx)

	pokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, id)
	if err != nil {
		if err.Error() == "pohon kinerja tidak ditemukan" {
			// Jika pohon kinerja tidak ditemukan, kembalikan response kosong
			return pohonkinerja.PohonKinerjaAdminResponseData{}, nil
		}
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Jika pohon kinerja ditemukan, ambil data OPD
	if pokin.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
		if err == nil {
			pokin.NamaOpd = opd.NamaOpd
		}
	}

	// Konversi id ke string untuk pencarian indikator
	pokinIdStr := fmt.Sprint(pokin.Id)

	// Ambil indikator berdasarkan pokin ID
	indikators, err := service.pohonKinerjaRepository.FindIndikatorByPokinId(ctx, tx, pokinIdStr)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Konversi indikator ke response
	var indikatorResponses []pohonkinerja.IndikatorResponse
	for _, ind := range indikators {
		var targetResponses []pohonkinerja.TargetResponse
		for _, t := range ind.Target {
			targetResponse := pohonkinerja.TargetResponse{
				Id:              t.Id,
				IndikatorId:     t.IndikatorId,
				TargetIndikator: t.Target,
				SatuanIndikator: t.Satuan,
			}
			targetResponses = append(targetResponses, targetResponse)
		}

		indikatorResponse := pohonkinerja.IndikatorResponse{
			Id:            ind.Id,
			IdPokin:       ind.PokinId,
			NamaIndikator: ind.Indikator,
			Target:        targetResponses,
		}
		indikatorResponses = append(indikatorResponses, indikatorResponse)
	}

	response := pohonkinerja.PohonKinerjaAdminResponseData{
		Id:         pokin.Id,
		Parent:     pokin.Parent,
		NamaPohon:  pokin.NamaPohon,
		NamaOpd:    pokin.NamaOpd,
		JenisPohon: pokin.JenisPohon,
		LevelPohon: pokin.LevelPohon,
		KodeOpd:    pokin.KodeOpd,
		Keterangan: pokin.Keterangan,
		Tahun:      pokin.Tahun,
		Status:     pokin.Status,
		Indikators: indikatorResponses,
	}

	return response, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindAll(ctx context.Context, tahun string) (pohonkinerja.PohonKinerjaAdminResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil semua data pohon kinerja
	pokins, err := service.pohonKinerjaRepository.FindPokinAdminAll(ctx, tx, tahun)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponse{}, err
	}

	// Buat map untuk menyimpan data berdasarkan level dan parent
	pohonMap := make(map[int]map[int][]domain.PohonKinerja)

	// Kelompokkan data dan ambil data OPD untuk setiap pohon kinerja
	for i := range pokins {
		level := pokins[i].LevelPohon

		// Inisialisasi map untuk level jika belum ada
		if pohonMap[level] == nil {
			pohonMap[level] = make(map[int][]domain.PohonKinerja)
		}

		if pokins[i].KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokins[i].KodeOpd)
			if err == nil {
				pokins[i].NamaOpd = opd.NamaOpd
			}
		}

		pohonMap[level][pokins[i].Parent] = append(
			pohonMap[level][pokins[i].Parent],
			pokins[i],
		)
	}

	// Bangun response dimulai dari Tematik (level 0)
	var tematiks []pohonkinerja.TematikResponse
	for _, tematik := range pohonMap[0][0] {
		tematikResp := helper.BuildTematikResponse(pohonMap, tematik)
		tematiks = append(tematiks, tematikResp)
	}

	return pohonkinerja.PohonKinerjaAdminResponse{
		Tahun:   tahun,
		Tematik: tematiks,
	}, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindSubTematik(ctx context.Context, tahun string) (pohonkinerja.PohonKinerjaAdminResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil semua data pohon kinerja
	pokins, err := service.pohonKinerjaRepository.FindPokinAdminAll(ctx, tx, tahun)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponse{}, err
	}

	// Buat map untuk menyimpan data berdasarkan level dan parent
	pohonMap := make(map[int]map[int][]domain.PohonKinerja)
	for i := 1; i <= 2; i++ { // Hanya inisialisasi level 1 dan 2
		pohonMap[i] = make(map[int][]domain.PohonKinerja)
	}

	// Filter dan kelompokkan data
	for _, p := range pokins {
		// Ambil data OPD jika ada
		if p.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, p.KodeOpd)
			if err == nil {
				p.NamaOpd = opd.NamaOpd
			}
		}

		// Hanya masukkan data level 1 dan 2
		if p.LevelPohon >= 1 && p.LevelPohon <= 2 {
			pohonMap[p.LevelPohon][p.Parent] = append(pohonMap[p.LevelPohon][p.Parent], p)
		}
	}

	// Bangun response dimulai dari SubTematik (level 1)
	var tematiks []pohonkinerja.TematikResponse
	for _, subTematiks := range pohonMap[1] {
		// Urutkan subTematiks berdasarkan Id
		sort.Slice(subTematiks, func(i, j int) bool {
			return subTematiks[i].Id < subTematiks[j].Id
		})

		for _, subTematik := range subTematiks {
			var childs []interface{}

			// Tambahkan subsubtematik ke childs
			if subSubTematiks := pohonMap[2][subTematik.Id]; len(subSubTematiks) > 0 {
				// Urutkan subSubTematiks berdasarkan Id
				sort.Slice(subSubTematiks, func(i, j int) bool {
					return subSubTematiks[i].Id < subSubTematiks[j].Id
				})

				for _, subSubTematik := range subSubTematiks {
					subSubTematikResp := helper.BuildSubSubTematikResponse(pohonMap, subSubTematik)
					childs = append(childs, subSubTematikResp)
				}
			}

			tematikResp := pohonkinerja.TematikResponse{
				Id:         subTematik.Id,
				Parent:     &subTematik.Parent,
				Tema:       subTematik.NamaPohon,
				JenisPohon: subTematik.JenisPohon,
				LevelPohon: subTematik.LevelPohon,
				Keterangan: subTematik.Keterangan,
				Indikators: helper.ConvertToIndikatorResponses(subTematik.Indikator),
				Child:      childs,
			}
			tematiks = append(tematiks, tematikResp)
		}
	}

	// Urutkan hasil akhir berdasarkan Id
	sort.Slice(tematiks, func(i, j int) bool {
		return tematiks[i].Id < tematiks[j].Id
	})

	return pohonkinerja.PohonKinerjaAdminResponse{
		Tahun:   tahun,
		Tematik: tematiks,
	}, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinAdminByIdHierarki(ctx context.Context, idPokin int) (pohonkinerja.TematikResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.TematikResponse{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil data pohon kinerja
	pokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, idPokin)
	if err != nil {
		return pohonkinerja.TematikResponse{}, err
	}

	// Validasi level pohon harus 0
	if pokin.LevelPohon != 0 {
		return pohonkinerja.TematikResponse{}, fmt.Errorf("id yang diberikan bukan merupakan level tematik (level 0)")
	}

	// Ambil semua data pohon kinerja
	pokins, err := service.pohonKinerjaRepository.FindPokinAdminByIdHierarki(ctx, tx, idPokin)
	if err != nil {
		return pohonkinerja.TematikResponse{}, err
	}

	// Buat map untuk menyimpan data berdasarkan level dan parent
	pohonMap := make(map[int]map[int][]domain.PohonKinerja)

	// Kelompokkan data
	for _, p := range pokins {
		level := p.LevelPohon

		// Inisialisasi map untuk level jika belum ada
		if pohonMap[level] == nil {
			pohonMap[level] = make(map[int][]domain.PohonKinerja)
		}

		// Ambil data OPD jika ada
		if p.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, p.KodeOpd)
			if err == nil {
				p.NamaOpd = opd.NamaOpd
			}
		}

		countReview, err := service.reviewRepository.CountReviewByPohonKinerja(ctx, tx, p.Id)
		if err == nil {
			p.CountReview = countReview
		}

		// Ambil data pelaksana untuk level 4 ke atas (strategic, tactical, operational)
		if p.LevelPohon >= 4 {
			pelaksanas, err := service.pohonKinerjaRepository.FindPelaksanaPokin(ctx, tx, fmt.Sprint(p.Id))
			if err == nil {
				for i := range pelaksanas {
					// Ambil detail pegawai untuk setiap pelaksana
					pegawai, err := service.pegawaiRepository.FindById(ctx, tx, pelaksanas[i].PegawaiId)
					if err == nil {
						pelaksanas[i].NamaPegawai = pegawai.NamaPegawai
					}
				}
				p.Pelaksana = pelaksanas
			}
		}

		pohonMap[level][p.Parent] = append(pohonMap[level][p.Parent], p)
	}

	// Tambahkan map untuk melacak indikator yang sudah diproses
	processedIndikators := make(map[string]bool)

	// Bangun response hierarki
	var tematikResponse pohonkinerja.TematikResponse
	if tematik, exists := pohonMap[0][0]; exists && len(tematik) > 0 {
		var childs []interface{}

		// Tambahkan strategic langsung ke childs jika ada
		if strategics := pohonMap[4][tematik[0].Id]; len(strategics) > 0 {
			sort.Slice(strategics, func(i, j int) bool {
				return strategics[i].Id < strategics[j].Id
			})

			for _, strategic := range strategics {
				strategicResp := helper.BuildStrategicResponse(pohonMap, strategic)
				childs = append(childs, strategicResp)
			}
		}

		// Tambahkan subtematik ke childs
		if subTematiks := pohonMap[1][tematik[0].Id]; len(subTematiks) > 0 {
			sort.Slice(subTematiks, func(i, j int) bool {
				return subTematiks[i].Id < subTematiks[j].Id
			})

			for _, subTematik := range subTematiks {
				subTematikResp := helper.BuildSubTematikResponse(pohonMap, subTematik)
				childs = append(childs, subTematikResp)
			}
		}

		// Konversi indikator dengan pengecekan duplikasi
		var uniqueIndikators []pohonkinerja.IndikatorResponse
		for _, ind := range tematik[0].Indikator {
			// Cek apakah indikator sudah diproses
			if !processedIndikators[ind.Id] {
				processedIndikators[ind.Id] = true
				indResp := helper.ConvertToIndikatorResponse(ind)
				uniqueIndikators = append(uniqueIndikators, indResp)
			}
		}

		tematikResponse = pohonkinerja.TematikResponse{
			Id:          tematik[0].Id,
			Parent:      nil,
			Tema:        tematik[0].NamaPohon,
			JenisPohon:  tematik[0].JenisPohon,
			LevelPohon:  tematik[0].LevelPohon,
			Keterangan:  tematik[0].Keterangan,
			IsActive:    tematik[0].IsActive,
			CountReview: tematik[0].CountReview,
			Indikators:  uniqueIndikators,
			Child:       childs,
		}
	}

	return tematikResponse, nil
}

func (service *PohonKinerjaAdminServiceImpl) CreateStrategicAdmin(ctx context.Context, request pohonkinerja.PohonKinerjaAdminStrategicCreateRequest) (pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Cek apakah pohon kinerja sudah pernah diclone
	cloneFrom, err := service.pohonKinerjaRepository.CheckCloneFrom(ctx, tx, request.IdToClone)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	cloneReference := request.IdToClone
	if cloneFrom != 0 {
		cloneReference = cloneFrom
	}

	// Ambil data pohon kinerja yang akan diclone
	existingPokin, err := service.pohonKinerjaRepository.FindPokinToClone(ctx, tx, request.IdToClone)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Validasi parent level
	err = service.pohonKinerjaRepository.ValidateParentLevelTarikStrategiOpd(ctx, tx, request.Parent, existingPokin.LevelPohon)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Set jenis pohon berdasarkan level
	jenisPohon := determineJenisPohon(existingPokin.LevelPohon)

	var namaOpd string
	if existingPokin.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, existingPokin.KodeOpd)
		if err == nil {
			namaOpd = opd.NamaOpd
		}
	}

	// Clone pohon kinerja utama
	newPokin := domain.PohonKinerja{
		Parent:     request.Parent,
		NamaPohon:  existingPokin.NamaPohon,
		JenisPohon: jenisPohon,
		LevelPohon: existingPokin.LevelPohon,
		KodeOpd:    existingPokin.KodeOpd,
		Keterangan: existingPokin.Keterangan,
		Tahun:      existingPokin.Tahun,
		Status:     "tarik pokin opd",
		CloneFrom:  cloneReference,
		IsActive:   existingPokin.IsActive,
	}

	newPokinId, err := service.pohonKinerjaRepository.InsertClonedPokin(ctx, tx, newPokin)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Clone indikator dan target untuk pohon utama
	indikatorResponses, err := service.cloneIndikatorAndTargets(ctx, tx, request.IdToClone, newPokinId)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	var childs []interface{}
	if request.Turunan {
		// Jika turunan true, clone semua child pohon
		childs, err = service.cloneChildrenHierarchy(ctx, tx, request.IdToClone, newPokinId)
		if err != nil {
			return pohonkinerja.PohonKinerjaAdminResponseData{}, err
		}
	}

	response := pohonkinerja.PohonKinerjaAdminResponseData{
		Id:         int(newPokinId),
		Parent:     request.Parent,
		NamaPohon:  existingPokin.NamaPohon,
		JenisPohon: jenisPohon,
		LevelPohon: existingPokin.LevelPohon,
		KodeOpd:    existingPokin.KodeOpd,
		NamaOpd:    namaOpd,
		Keterangan: existingPokin.Keterangan,
		Tahun:      existingPokin.Tahun,
		Status:     "tarik pokin opd",
		IsActive:   existingPokin.IsActive,
		Indikators: indikatorResponses,
		Childs:     childs,
		PerangkatDaerah: &opdmaster.OpdResponseForAll{
			KodeOpd: existingPokin.KodeOpd,
			NamaOpd: namaOpd,
		},
	}

	return response, nil
}

// Helper function untuk menentukan jenis pohon berdasarkan level
func determineJenisPohon(level int) string {
	switch level {
	case 4:
		return "Strategic"
	case 5:
		return "Tactical"
	case 6:
		return "Operational"
	default:
		return "Unknown"
	}
}

// Helper function untuk clone indikator dan target
func (service *PohonKinerjaAdminServiceImpl) cloneIndikatorAndTargets(ctx context.Context, tx *sql.Tx, sourcePokinId int, newPokinId int64) ([]pohonkinerja.IndikatorResponse, error) {
	indikators, err := service.pohonKinerjaRepository.FindIndikatorToClone(ctx, tx, sourcePokinId)
	if err != nil {
		return nil, err
	}

	var indikatorResponses []pohonkinerja.IndikatorResponse
	for _, indikator := range indikators {
		newIndikatorId := "IND-POKIN-" + uuid.New().String()[:6]
		indikator.CloneFrom = indikator.Id

		err = service.pohonKinerjaRepository.InsertClonedIndikator(ctx, tx, newIndikatorId, newPokinId, indikator)
		if err != nil {
			return nil, err
		}

		targets, err := service.pohonKinerjaRepository.FindTargetToClone(ctx, tx, indikator.Id)
		if err != nil {
			return nil, err
		}

		var targetResponses []pohonkinerja.TargetResponse
		for _, target := range targets {
			newTargetId := "TRGT-IND-POKIN-" + uuid.New().String()[:5]
			target.CloneFrom = target.Id

			err = service.pohonKinerjaRepository.InsertClonedTarget(ctx, tx, newTargetId, newIndikatorId, target)
			if err != nil {
				return nil, err
			}

			targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
				Id:              newTargetId,
				IndikatorId:     newIndikatorId,
				TargetIndikator: target.Target,
				SatuanIndikator: target.Satuan,
			})
		}

		indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
			Id:            newIndikatorId,
			IdPokin:       fmt.Sprint(newPokinId),
			NamaIndikator: indikator.Indikator,
			Target:        targetResponses,
		})
	}

	return indikatorResponses, nil
}

// Helper function untuk clone hierarki child
func (service *PohonKinerjaAdminServiceImpl) cloneChildrenHierarchy(ctx context.Context, tx *sql.Tx, sourceParentId int, newParentId int64) ([]interface{}, error) {
	children, err := service.pohonKinerjaRepository.FindChildPokins(ctx, tx, int64(sourceParentId))
	if err != nil {
		return nil, err
	}

	var childResponses []interface{}
	for _, child := range children {
		// Skip jika level di luar range 4-6
		if child.LevelPohon < 4 || child.LevelPohon > 6 {
			continue
		}

		jenisPohon := determineJenisPohon(child.LevelPohon)
		if jenisPohon == "Unknown" {
			continue
		}

		newChild := domain.PohonKinerja{
			Parent:     int(newParentId),
			NamaPohon:  child.NamaPohon,
			JenisPohon: jenisPohon,
			LevelPohon: child.LevelPohon,
			KodeOpd:    child.KodeOpd,
			Keterangan: child.Keterangan,
			Tahun:      child.Tahun,
			Status:     "tarik pokin opd",
			CloneFrom:  child.Id,
			IsActive:   child.IsActive,
		}

		newChildId, err := service.pohonKinerjaRepository.InsertClonedPokin(ctx, tx, newChild)
		if err != nil {
			return nil, err
		}

		indikatorResponses, err := service.cloneIndikatorAndTargets(ctx, tx, child.Id, newChildId)
		if err != nil {
			return nil, err
		}

		var namaOpd string
		if child.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, child.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		var childChilds []interface{}
		if child.LevelPohon < 6 {
			childChilds, err = service.cloneChildrenHierarchy(ctx, tx, child.Id, newChildId)
			if err != nil {
				return nil, err
			}
		}

		childResponse := pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         int(newChildId),
			Parent:     int(newParentId),
			NamaPohon:  child.NamaPohon,
			JenisPohon: jenisPohon,
			LevelPohon: child.LevelPohon,
			KodeOpd:    child.KodeOpd,
			NamaOpd:    namaOpd,
			Keterangan: child.Keterangan,
			Tahun:      child.Tahun,
			Status:     "tarik pokin opd",
			IsActive:   child.IsActive, // Tambahkan ini
			Indikators: indikatorResponses,
			Childs:     childChilds,
			PerangkatDaerah: &opdmaster.OpdResponseForAll{
				KodeOpd: child.KodeOpd,
				NamaOpd: namaOpd,
			},
		}

		childResponses = append(childResponses, childResponse)
	}

	return childResponses, nil
}

func (service *PohonKinerjaAdminServiceImpl) CloneStrategiFromPemda(ctx context.Context, request pohonkinerja.PohonKinerjaAdminStrategicCreateRequest) (pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi status pokin
	status, err := service.pohonKinerjaRepository.CheckPokinStatus(ctx, tx, request.IdToClone)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	if status != "menunggu_disetujui" && status != "ditolak" {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, errors.New("hanya pohon kinerja dengan status menunggu_disetujui atau ditolak yang dapat diclone")
	}

	// Fungsi helper untuk clone indikator dan target
	cloneIndikatorAndTargets := func(ctx context.Context, tx *sql.Tx, oldPokinId int, newPokinId int64) error {
		indikators, err := service.pohonKinerjaRepository.FindIndikatorToClone(ctx, tx, oldPokinId)
		if err != nil {
			return err
		}

		for _, indikator := range indikators {
			newIndikatorId := "IND-POKIN-" + uuid.New().String()[:6]
			indikator.CloneFrom = indikator.Id

			err = service.pohonKinerjaRepository.InsertClonedIndikator(ctx, tx, newIndikatorId, newPokinId, indikator)
			if err != nil {
				return err
			}

			targets, err := service.pohonKinerjaRepository.FindTargetToClone(ctx, tx, indikator.Id)
			if err != nil {
				return err
			}

			for _, target := range targets {
				newTargetId := "TRGT-IND-POKIN-" + uuid.New().String()[:5]
				target.CloneFrom = target.Id

				err = service.pohonKinerjaRepository.InsertClonedTarget(ctx, tx, newTargetId, newIndikatorId, target)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Fungsi rekursif untuk mengupdate status
	var updateStatusRecursive func(ctx context.Context, tx *sql.Tx, pokinId int, parentKodeOpd string) error
	updateStatusRecursive = func(ctx context.Context, tx *sql.Tx, pokinId int, parentKodeOpd string) error {
		// Ambil data pokin saat ini
		pokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, pokinId)
		if err != nil {
			return err
		}

		// Cek apakah kode_opd sama dengan parent atau ini adalah pohon pertama
		if parentKodeOpd == "" || pokin.KodeOpd == parentKodeOpd {
			// Update status hanya jika kode_opd sesuai dan statusnya menunggu_disetujui atau ditolak
			if pokin.Status == "menunggu_disetujui" || pokin.Status == "ditolak" {
				err = service.pohonKinerjaRepository.UpdatePokinStatus(ctx, tx, pokinId, "disetujui")
				if err != nil {
					return err
				}
			}

			// Simpan kode_opd untuk digunakan di level berikutnya
			currentKodeOpd := pokin.KodeOpd

			// Cari child pokin
			childPokins, err := service.pohonKinerjaRepository.FindChildPokins(ctx, tx, int64(pokinId))
			if err != nil {
				return err
			}

			// Update status untuk setiap child secara rekursif
			for _, childPokin := range childPokins {
				err = updateStatusRecursive(ctx, tx, childPokin.Id, currentKodeOpd)
				if err != nil {
					return err
				}
			}
		}
		// Jika kode_opd berbeda, tidak perlu update status dan tidak perlu memeriksa child-nya

		return nil
	}

	// Fungsi rekursif untuk clone pokin dan child-nya
	var clonePokinRecursive func(ctx context.Context, tx *sql.Tx, pokinId int, parentId int, parentKodeOpd string) (int64, error)
	clonePokinRecursive = func(ctx context.Context, tx *sql.Tx, pokinId int, parentId int, parentKodeOpd string) (int64, error) {
		existingPokin, err := service.pohonKinerjaRepository.FindPokinToClone(ctx, tx, pokinId)
		if err != nil {
			return 0, err
		}

		// Hanya clone jika kode_opd sama dengan parent atau ini adalah pohon pertama yang diclone
		if parentKodeOpd != "" && existingPokin.KodeOpd != parentKodeOpd {
			return 0, nil // Skip clone untuk pohon dengan kode_opd berbeda
		}

		// Simpan kode_opd untuk digunakan di level berikutnya
		currentKodeOpd := existingPokin.KodeOpd

		// Siapkan data pokin baru
		newPokin := domain.PohonKinerja{
			Parent:     parentId,
			NamaPohon:  existingPokin.NamaPohon,
			JenisPohon: existingPokin.JenisPohon,
			LevelPohon: existingPokin.LevelPohon,
			KodeOpd:    existingPokin.KodeOpd,
			Keterangan: existingPokin.Keterangan,
			Tahun:      existingPokin.Tahun,
			Status:     "pokin dari pemda",
			CloneFrom:  pokinId,
			Pelaksana:  existingPokin.Pelaksana,
		}

		// Insert pokin baru
		newPokinId, err := service.pohonKinerjaRepository.InsertClonedPokinWithStatus(ctx, tx, newPokin)
		if err != nil {
			return 0, err
		}

		// Clone indikator dan target
		err = cloneIndikatorAndTargets(ctx, tx, pokinId, newPokinId)
		if err != nil {
			return 0, err
		}

		// Clone pelaksana jika ada
		if len(existingPokin.Pelaksana) > 0 {
			for _, pelaksana := range existingPokin.Pelaksana {
				newPelaksanaId := "PLKS-" + uuid.New().String()[:8]
				err = service.pohonKinerjaRepository.InsertClonedPelaksana(ctx, tx, newPelaksanaId, newPokinId, pelaksana)
				if err != nil {
					return 0, err
				}
			}
		}

		// Cari dan clone child pokin
		childPokins, err := service.pohonKinerjaRepository.FindChildPokins(ctx, tx, int64(pokinId))
		if err != nil {
			return 0, err
		}

		for _, childPokin := range childPokins {
			// Gunakan kode_opd saat ini untuk validasi child
			_, err := clonePokinRecursive(ctx, tx, childPokin.Id, int(newPokinId), currentKodeOpd)
			if err != nil {
				return 0, err
			}
		}

		return newPokinId, nil
	}

	// Mulai proses clone dengan kode_opd kosong untuk root
	newPokinId, err := clonePokinRecursive(ctx, tx, request.IdToClone, request.Parent, "")
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Update status untuk pohon asli dan semua child-nya yang berstatus menunggu_disetujui
	err = updateStatusRecursive(ctx, tx, request.IdToClone, "")
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Ambil data lengkap pokin baru untuk response
	result, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, int(newPokinId))
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Siapkan response
	var indikatorResponses []pohonkinerja.IndikatorResponse
	indikators, err := service.pohonKinerjaRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(newPokinId))
	if err == nil {
		for _, ind := range indikators {
			var targetResponses []pohonkinerja.TargetResponse
			targets, err := service.pohonKinerjaRepository.FindTargetByIndikatorId(ctx, tx, ind.Id)
			if err == nil {
				for _, t := range targets {
					targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
						Id:              t.Id,
						IndikatorId:     t.IndikatorId,
						TargetIndikator: t.Target,
						SatuanIndikator: t.Satuan,
					})
				}
			}

			indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
				Id:            ind.Id,
				IdPokin:       fmt.Sprint(newPokinId),
				NamaIndikator: ind.Indikator,
				Target:        targetResponses,
			})
		}
	}

	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
	for _, p := range result.Pelaksana {
		pegawai, err := service.pegawaiRepository.FindById(ctx, tx, p.PegawaiId)
		if err == nil {
			pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
				Id:          p.Id,
				PegawaiId:   p.PegawaiId,
				NamaPegawai: pegawai.NamaPegawai,
			})
		}
	}

	var namaOpd string
	if result.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, result.KodeOpd)
		if err == nil {
			namaOpd = opd.NamaOpd
		}
	}

	response := pohonkinerja.PohonKinerjaAdminResponseData{
		Id:         int(newPokinId),
		Parent:     int(result.Parent),
		NamaPohon:  result.NamaPohon,
		JenisPohon: result.JenisPohon,
		LevelPohon: result.LevelPohon,
		KodeOpd:    result.KodeOpd,
		NamaOpd:    namaOpd,
		Keterangan: result.Keterangan,
		Tahun:      result.Tahun,
		Status:     result.Status,
		Indikators: indikatorResponses,
		Pelaksana:  pelaksanaResponses,
	}

	return response, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinByTematik(ctx context.Context, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	pokins, err := service.pohonKinerjaRepository.FindPokinByJenisPohon(ctx, tx, "Tematik", 0, tahun, "", "")
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			Tahun:      pokin.Tahun,
		})
	}

	return result, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinByStrategic(ctx context.Context, kodeOpd string, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD jika kodeOpd tidak kosong
	if kodeOpd != "" {
		_, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
		if err != nil {
			return nil, errors.New("kode opd tidak ditemukan")
		}
	}

	// Ambil data pohon kinerja dengan jenis "Strategic" dan level 4
	pokins, err := service.pohonKinerjaRepository.FindPokinByJenisPohon(ctx, tx, "Strategic", 4, tahun, kodeOpd, "")
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		// Ambil data OPD jika ada kodeOpd
		var namaOpd string
		if pokin.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    pokin.KodeOpd,
			NamaOpd:    namaOpd,
			Tahun:      pokin.Tahun,
		})
	}

	return result, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinByTactical(ctx context.Context, kodeOpd string, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD jika kodeOpd tidak kosong
	if kodeOpd != "" {
		_, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
		if err != nil {
			return nil, errors.New("kode opd tidak ditemukan")
		}
	}

	// Ambil data pohon kinerja dengan jenis "Strategic" dan level 4
	pokins, err := service.pohonKinerjaRepository.FindPokinByJenisPohon(ctx, tx, "Tactical", 5, tahun, kodeOpd, "")
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		var namaOpd string
		if pokin.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    pokin.KodeOpd,
			NamaOpd:    namaOpd,
			Tahun:      pokin.Tahun,
		})
	}

	return result, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinByOperational(ctx context.Context, kodeOpd string, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD jika kodeOpd tidak kosong
	if kodeOpd != "" {
		_, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
		if err != nil {
			return nil, errors.New("kode opd tidak ditemukan")
		}
	}

	// Ambil data pohon kinerja dengan jenis "Strategic" dan level 4
	pokins, err := service.pohonKinerjaRepository.FindPokinByJenisPohon(ctx, tx, "Operational", 6, tahun, kodeOpd, "")
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		// Ambil data OPD jika ada kodeOpd
		var namaOpd string
		if pokin.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    pokin.KodeOpd,
			NamaOpd:    namaOpd,
			Tahun:      pokin.Tahun,
		})
	}

	return result, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinByStatus(ctx context.Context, kodeOpd string, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	pokins, err := service.pohonKinerjaRepository.FindPokinByStatus(ctx, tx, kodeOpd, tahun, "menunggu_disetujui")
	if err != nil {
		return nil, err
	}

	var pokinResponses []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		pokinResponses = append(pokinResponses, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    pokin.KodeOpd,
			Tahun:      pokin.Tahun,
		})
	}

	return pokinResponses, nil
}

func (service *PohonKinerjaAdminServiceImpl) TolakPokin(ctx context.Context, request pohonkinerja.PohonKinerjaAdminTolakRequest) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	status, err := service.pohonKinerjaRepository.CheckPokinStatus(ctx, tx, request.Id)
	if err != nil {
		return err
	}

	if status != "menunggu_disetujui" {
		return errors.New("hanya pohon kinerja dengan status menunggu_disetujui yang dapat ditolak")
	}

	err = service.pohonKinerjaRepository.UpdatePokinStatusTolak(ctx, tx, request.Id, "ditolak")
	if err != nil {
		return err
	}

	return nil
}

func (service *PohonKinerjaAdminServiceImpl) CrosscuttingOpd(ctx context.Context, request pohonkinerja.PohonKinerjaAdminStrategicCreateRequest) (pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}
	defer helper.CommitOrRollback(tx)

	// Cek apakah pohon kinerja sudah pernah diclone
	cloneFrom, err := service.pohonKinerjaRepository.CheckCloneFrom(ctx, tx, request.IdToClone)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Tentukan referensi clone
	cloneReference := request.IdToClone
	if cloneFrom != 0 {
		cloneReference = cloneFrom
	}

	existingPokin, err := service.pohonKinerjaRepository.FindPokinToClone(ctx, tx, request.IdToClone)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	err = service.pohonKinerjaRepository.ValidateParentLevel(ctx, tx, request.Parent, existingPokin.LevelPohon)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	// Validasi JenisPohon
	if request.JenisPohon == "" {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, errors.New("jenis pohon tidak boleh kosong")
	}

	// Update status pokin yang diclone menjadi disetujui
	err = service.pohonKinerjaRepository.UpdatePokinStatus(ctx, tx, request.IdToClone, "disetujui")
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	var namaOpd string
	if existingPokin.KodeOpd != "" {
		opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, existingPokin.KodeOpd)
		if err == nil {
			namaOpd = opd.NamaOpd
		}
	}

	newPokin := domain.PohonKinerja{
		Parent:     request.Parent,
		NamaPohon:  existingPokin.NamaPohon,
		JenisPohon: request.JenisPohon,
		LevelPohon: existingPokin.LevelPohon,
		KodeOpd:    existingPokin.KodeOpd,
		Keterangan: existingPokin.Keterangan,
		Tahun:      existingPokin.Tahun,
		Status:     "crosscutting_menunggu",
		CloneFrom:  cloneReference,
		Pelaksana:  existingPokin.Pelaksana,
	}

	newPokinId, err := service.pohonKinerjaRepository.InsertClonedPokinWithStatus(ctx, tx, newPokin)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	indikators, err := service.pohonKinerjaRepository.FindIndikatorToClone(ctx, tx, request.IdToClone)
	if err != nil {
		return pohonkinerja.PohonKinerjaAdminResponseData{}, err
	}

	var indikatorResponses []pohonkinerja.IndikatorResponse

	for _, indikator := range indikators {
		newIndikatorId := "IND-POKIN-" + uuid.New().String()[:6]

		err = service.pohonKinerjaRepository.InsertClonedIndikator(ctx, tx, newIndikatorId, newPokinId, indikator)
		if err != nil {
			return pohonkinerja.PohonKinerjaAdminResponseData{}, err
		}

		targets, err := service.pohonKinerjaRepository.FindTargetToClone(ctx, tx, indikator.Id)
		if err != nil {
			return pohonkinerja.PohonKinerjaAdminResponseData{}, err
		}

		var targetResponses []pohonkinerja.TargetResponse

		for _, target := range targets {
			newTargetId := "TRGT-IND-POKIN-" + uuid.New().String()[:5]
			err = service.pohonKinerjaRepository.InsertClonedTarget(ctx, tx, newTargetId, newIndikatorId, target)
			if err != nil {
				return pohonkinerja.PohonKinerjaAdminResponseData{}, err
			}

			targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
				Id:              newTargetId,
				IndikatorId:     newIndikatorId,
				TargetIndikator: target.Target,
				SatuanIndikator: target.Satuan,
			})
		}

		indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
			Id:            newIndikatorId,
			IdPokin:       fmt.Sprint(newPokinId),
			NamaIndikator: indikator.Indikator,
			Target:        targetResponses,
		})
	}

	var pelaksanaResponses []pohonkinerja.PelaksanaOpdResponse
	for _, pelaksana := range existingPokin.Pelaksana {
		// Ambil detail pegawai
		pegawai, err := service.pegawaiRepository.FindById(ctx, tx, pelaksana.PegawaiId)
		if err == nil {
			pelaksanaResponses = append(pelaksanaResponses, pohonkinerja.PelaksanaOpdResponse{
				Id:          pelaksana.Id,
				PegawaiId:   pelaksana.PegawaiId,
				NamaPegawai: pegawai.NamaPegawai,
			})
		}
	}

	response := pohonkinerja.PohonKinerjaAdminResponseData{
		Id:         int(newPokinId),
		Parent:     request.Parent,
		NamaPohon:  existingPokin.NamaPohon,
		JenisPohon: request.JenisPohon,
		LevelPohon: existingPokin.LevelPohon,
		KodeOpd:    existingPokin.KodeOpd,
		NamaOpd:    namaOpd,
		Keterangan: existingPokin.Keterangan,
		Tahun:      existingPokin.Tahun,
		Status:     "crosscutting_menunggu",
		Indikators: indikatorResponses,
		Pelaksana:  pelaksanaResponses,
	}

	return response, nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinByCrosscuttingStatus(ctx context.Context, kodeOpd string, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD jika kodeOpd tidak kosong
	if kodeOpd != "" {
		_, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
		if err != nil {
			return nil, errors.New("kode opd tidak ditemukan")
		}
	}

	// Ambil data pohon kinerja dengan status crosscutting_menunggu
	pokins, err := service.pohonKinerjaRepository.FindPokinByCrosscuttingStatus(ctx, tx, kodeOpd, tahun)
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		// Ambil data OPD jika ada kodeOpd
		var namaOpd string
		if pokin.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		// Ambil data parent pokin untuk mendapatkan pengaju OPD
		var namaOpdPengaju string
		if pokin.Parent != 0 {
			parentPokin, err := service.pohonKinerjaRepository.FindPokinAdminById(ctx, tx, pokin.Parent)
			if err == nil && parentPokin.KodeOpd != "" {
				opdPengaju, err := service.opdRepository.FindByKodeOpd(ctx, tx, parentPokin.KodeOpd)
				if err == nil {
					namaOpdPengaju = opdPengaju.NamaOpd
				}
			}
		}

		// Ambil data indikator
		indikators, err := service.pohonKinerjaRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(pokin.Id))
		if err == nil {
			var indikatorResponses []pohonkinerja.IndikatorResponse
			for _, indikator := range indikators {
				// Ambil data target untuk setiap indikator
				targets, err := service.pohonKinerjaRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id)
				if err != nil {
					continue
				}

				// Konversi target ke response
				var targetResponses []pohonkinerja.TargetResponse
				for _, target := range targets {
					targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
						Id:              target.Id,
						IndikatorId:     target.IndikatorId,
						TargetIndikator: target.Target,
						SatuanIndikator: target.Satuan,
					})
				}

				indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
					Id:            indikator.Id,
					IdPokin:       fmt.Sprint(pokin.Id),
					NamaIndikator: indikator.Indikator,
					Target:        targetResponses,
				})
			}

			result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
				Id:             pokin.Id,
				Parent:         pokin.Parent,
				NamaPohon:      pokin.NamaPohon,
				JenisPohon:     pokin.JenisPohon,
				LevelPohon:     pokin.LevelPohon,
				KodeOpd:        pokin.KodeOpd,
				NamaOpd:        namaOpd,
				NamaOpdPengaju: namaOpdPengaju,
				Keterangan:     pokin.Keterangan,
				Tahun:          pokin.Tahun,
				Status:         pokin.Status,
				Indikators:     indikatorResponses,
			})
		}
	}

	return result, nil
}
func (service *PohonKinerjaAdminServiceImpl) FindPokinFromPemda(ctx context.Context, kodeOpd string, tahun string) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD jika kodeOpd tidak kosong
	if kodeOpd != "" {
		_, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
		if err != nil {
			return nil, errors.New("kode opd tidak ditemukan")
		}
	}

	// Ambil semua data pohon kinerja
	pokins, err := service.pohonKinerjaRepository.FindPokinByJenisPohon(ctx, tx, "", 0, tahun, kodeOpd, "")
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		if pokin.Status != "menunggu_disetujui" && pokin.Status != "ditolak" {
			continue
		}

		// Ambil data OPD jika ada kodeOpd
		var namaOpd string
		if pokin.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		// Ambil data indikator
		indikators, err := service.pohonKinerjaRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(pokin.Id))
		if err != nil {
			return nil, err
		}

		var indikatorResponses []pohonkinerja.IndikatorResponse
		for _, indikator := range indikators {
			// Ambil data target untuk setiap indikator
			targets, err := service.pohonKinerjaRepository.FindTargetByIndikatorId(ctx, tx, indikator.Id)
			if err != nil {
				continue
			}

			// Konversi target ke response
			var targetResponses []pohonkinerja.TargetResponse
			for _, target := range targets {
				targetResponses = append(targetResponses, pohonkinerja.TargetResponse{
					Id:              target.Id,
					IndikatorId:     target.IndikatorId,
					TargetIndikator: target.Target,
					SatuanIndikator: target.Satuan,
				})
			}

			indikatorResponses = append(indikatorResponses, pohonkinerja.IndikatorResponse{
				Id:            indikator.Id,
				IdPokin:       fmt.Sprint(pokin.Id),
				NamaIndikator: indikator.Indikator,
				Target:        targetResponses,
			})
		}

		result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    pokin.KodeOpd,
			NamaOpd:    namaOpd,
			Tahun:      pokin.Tahun,
			Keterangan: pokin.Keterangan,
			Status:     pokin.Status,
			IsActive:   pokin.IsActive,
			Indikators: indikatorResponses,
		})
	}

	return result, nil
}

func (service *PohonKinerjaAdminServiceImpl) TolakCrosscutting(ctx context.Context, request pohonkinerja.PohonKinerjaAdminTolakRequest) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	if request.Id == 0 {
		return errors.New("id tidak boleh kosong")
	}

	status, err := service.pohonKinerjaRepository.CheckPokinStatus(ctx, tx, request.Id)
	if err != nil {
		return err
	}

	if status != "crosscutting_menunggu" {
		return errors.New("hanya pohon kinerja dengan status crosscutting_menunggu yang dapat ditolak")
	}

	err = service.pohonKinerjaRepository.UpdatePokinStatusTolak(ctx, tx, request.Id, "crosscutting_ditolak")
	if err != nil {
		return err
	}

	return nil
}

func (service *PohonKinerjaAdminServiceImpl) SetujuiCrosscutting(ctx context.Context, request pohonkinerja.PohonKinerjaAdminTolakRequest) error {
	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helper.CommitOrRollback(tx)

	if request.Id == 0 {
		return errors.New("id tidak boleh kosong")
	}

	status, err := service.pohonKinerjaRepository.CheckPokinStatus(ctx, tx, request.Id)
	if err != nil {
		return err
	}

	if status != "crosscutting_menunggu" {
		return errors.New("hanya pohon kinerja dengan status crosscutting_menunggu yang dapat disetujui")
	}

	err = service.pohonKinerjaRepository.UpdatePokinStatusTolak(ctx, tx, request.Id, "crosscutting_disetujui")
	if err != nil {
		return err
	}

	return nil
}

func (service *PohonKinerjaAdminServiceImpl) FindPokinFromOpd(ctx context.Context, kodeOpd string, tahun string, levelPohon int) ([]pohonkinerja.PohonKinerjaAdminResponseData, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Validasi OPD jika kodeOpd tidak kosong
	if kodeOpd != "" {
		_, err := service.opdRepository.FindByKodeOpd(ctx, tx, kodeOpd)
		if err != nil {
			return nil, errors.New("kode opd tidak ditemukan")
		}
	}

	pokins, err := service.pohonKinerjaRepository.FindPokinByJenisPohon(ctx, tx, "", levelPohon, tahun, kodeOpd, "")
	if err != nil {
		return nil, err
	}

	if len(pokins) == 0 {
		return nil, nil
	}

	var result []pohonkinerja.PohonKinerjaAdminResponseData
	for _, pokin := range pokins {
		// Skip pohon kinerja dengan status yang tidak diinginkan
		if pokin.Status == "menunggu_disetujui" || pokin.Status == "crosscutting_menunggu" || pokin.Status == "tarik pokin opd" || pokin.Status == "disetujui" || pokin.Status == "ditolak" || pokin.Status == "crosscutting_ditolak" {
			continue
		}

		// Ambil data OPD jika ada kodeOpd
		var namaOpd string
		if pokin.KodeOpd != "" {
			opd, err := service.opdRepository.FindByKodeOpd(ctx, tx, pokin.KodeOpd)
			if err == nil {
				namaOpd = opd.NamaOpd
			}
		}

		// Konversi indikator dan target
		indikators, err := service.pohonKinerjaRepository.FindIndikatorByPokinId(ctx, tx, fmt.Sprint(pokin.Id))
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

		result = append(result, pohonkinerja.PohonKinerjaAdminResponseData{
			Id:         pokin.Id,
			Parent:     pokin.Parent,
			NamaPohon:  pokin.NamaPohon,
			JenisPohon: pokin.JenisPohon,
			LevelPohon: pokin.LevelPohon,
			KodeOpd:    pokin.KodeOpd,
			NamaOpd:    namaOpd,
			Tahun:      pokin.Tahun,
			Status:     pokin.Status,
			Indikators: indikatorResponses,
		})
	}

	return result, nil
}

func (service *PohonKinerjaAdminServiceImpl) AktiforNonAktifTematik(ctx context.Context, request pohonkinerja.TematikStatusRequest) (string, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		if request.IsActive {
			return "gagal diaktifkan", err
		}
		return "gagal dinonaktifkan", err
	}
	defer helper.CommitOrRollback(tx)

	// Verifikasi bahwa pohon kinerja yang akan diubah adalah tematik (level 0)
	pokin, err := service.pohonKinerjaRepository.FindById(ctx, tx, request.Id)
	if err != nil {
		if request.IsActive {
			return "gagal diaktifkan", err
		}
		return "gagal dinonaktifkan", err
	}

	if pokin.LevelPohon != 0 {
		if request.IsActive {
			return "gagal diaktifkan", fmt.Errorf("pohon kinerja dengan id %d bukan merupakan tematik (level 0)", request.Id)
		}
		return "gagal dinonaktifkan", fmt.Errorf("pohon kinerja dengan id %d bukan merupakan tematik (level 0)", request.Id)
	}

	// Dapatkan semua children dan clone yang terkait
	affectedIds, err := service.pohonKinerjaRepository.GetChildrenAndClones(ctx, tx, request.Id, request.IsActive)
	if err != nil {
		if request.IsActive {
			return "gagal diaktifkan", err
		}
		return "gagal dinonaktifkan", err
	}

	// Update status tematik
	err = service.pohonKinerjaRepository.UpdateTematikStatus(ctx, tx, request.Id, request.IsActive)
	if err != nil {
		if request.IsActive {
			return "gagal diaktifkan", err
		}
		return "gagal dinonaktifkan", err
	}

	// Update status semua children dan clone
	for _, id := range affectedIds {
		err = service.pohonKinerjaRepository.UpdateTematikStatus(ctx, tx, id, request.IsActive)
		if err != nil {
			if request.IsActive {
				return "gagal diaktifkan", err
			}
			return "gagal dinonaktifkan", err
		}
	}

	if request.IsActive {
		return "berhasil diaktifkan", nil
	}
	return "berhasil dinonaktifkan", nil
}

func (service *PohonKinerjaAdminServiceImpl) FindListOpdAllTematik(ctx context.Context, tahun string) ([]pohonkinerja.TematikListOpdResponse, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helper.CommitOrRollback(tx)

	// Ambil data pohon kinerja dengan OPD
	pokins, err := service.pohonKinerjaRepository.FindListOpdAllTematik(ctx, tx, tahun)
	if err != nil {
		return nil, err
	}

	// Konversi ke response
	var responses []pohonkinerja.TematikListOpdResponse
	for _, pokin := range pokins {
		var listOpd []pohonkinerja.OpdListResponse
		for _, opd := range pokin.ListOpd {
			listOpd = append(listOpd, pohonkinerja.OpdListResponse{
				KodeOpd:         opd.KodeOpd,
				PerangkatDaerah: opd.PerangkatDaerah,
			})
		}

		// Sort list OPD berdasarkan kode OPD ASC
		sort.Slice(listOpd, func(i, j int) bool {
			return listOpd[i].KodeOpd < listOpd[j].KodeOpd
		})

		response := pohonkinerja.TematikListOpdResponse{
			Tematik:    pokin.NamaPohon,
			LevelPohon: pokin.LevelPohon,
			Tahun:      pokin.Tahun,
			IsActive:   pokin.IsActive,
			ListOpd:    listOpd,
		}
		responses = append(responses, response)
	}

	// Sort responses berdasarkan nama tematik ASC
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].Tematik < responses[j].Tematik
	})

	return responses, nil
}
