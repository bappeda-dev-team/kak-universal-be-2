package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
)

type SasaranOpdRepositoryImpl struct {
}

func NewSasaranOpdRepositoryImpl() *SasaranOpdRepositoryImpl {
	return &SasaranOpdRepositoryImpl{}
}

func (repository *SasaranOpdRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, KodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.SasaranOpd, error) {
	script := `
    SELECT DISTINCT
        pk.id as pokin_id,
        pk.nama_pohon,
        pk.jenis_pohon,
        pk.level_pohon,
        pk.tahun as tahun_pohon,
        pp.id as pelaksana_id,
        pp.pegawai_id,
        p.nip as pelaksana_nip,
        p.nama as nama_pegawai,
        so.id as sasaran_id,
        so.nama_sasaran_opd,
        so.tahun_awal,
        so.tahun_akhir,
        so.jenis_periode,
        i.id as indikator_id,
        i.indikator,
        i.rumus_perhitungan,
        i.sumber_data,
        t.id as target_id,
        t.tahun as target_tahun,
        t.target,
        t.satuan
    FROM tb_pohon_kinerja pk
    LEFT JOIN tb_pelaksana_pokin pp ON pk.id = pp.pohon_kinerja_id
    LEFT JOIN tb_pegawai p ON pp.pegawai_id = p.id
    LEFT JOIN (
        SELECT * FROM tb_sasaran_opd 
        WHERE tahun_awal = ? 
        AND tahun_akhir = ? 
        AND jenis_periode = ?
    ) so ON pk.id = so.pokin_id
    LEFT JOIN tb_indikator i ON so.id = i.sasaran_opd_id
    LEFT JOIN tb_target t ON i.id = t.indikator_id
    WHERE pk.level_pohon = 4 AND pk.parent = 0
    AND pk.kode_opd = ?
    AND CAST(pk.tahun AS UNSIGNED) BETWEEN CAST(? AS UNSIGNED) AND CAST(? AS UNSIGNED)
    ORDER BY pk.nama_pohon ASC, so.nama_sasaran_opd ASC`

	rows, err := tx.QueryContext(ctx, script,
		tahunAwal, tahunAkhir, jenisPeriode,
		KodeOpd,
		tahunAwal, tahunAkhir,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pokinMap := make(map[int]*domain.SasaranOpd)
	pelaksanaMap := make(map[string]bool)

	for rows.Next() {
		var (
			pokinId, levelPohon                  int
			namaPohon, jenisPohon, tahunPohon    string
			pelaksanaId, pegawaiId, pelaksanaNip sql.NullString
			namaPegawai                          sql.NullString
			sasaranId                            sql.NullInt64 // Menggunakan NullInt64 untuk ID integer
			namaSasaranOpd                       sql.NullString
			tahunAwalSasaran, tahunAkhirSasaran  sql.NullString
			jenisPeriodeSasaran                  sql.NullString
			indikatorId, indikator               sql.NullString
			rumusPerhitungan, sumberData         sql.NullString
			targetId, targetTahun                sql.NullString
			targetValue, targetSatuan            sql.NullString
		)

		err := rows.Scan(
			&pokinId, &namaPohon, &jenisPohon, &levelPohon, &tahunPohon,
			&pelaksanaId, &pegawaiId, &pelaksanaNip, &namaPegawai,
			&sasaranId, &namaSasaranOpd,
			&tahunAwalSasaran, &tahunAkhirSasaran, &jenisPeriodeSasaran,
			&indikatorId, &indikator,
			&rumusPerhitungan, &sumberData,
			&targetId, &targetTahun, &targetValue, &targetSatuan,
		)
		if err != nil {
			return nil, err
		}

		// Proses SasaranOpd
		sasaranOpd, exists := pokinMap[pokinId]
		if !exists {
			sasaranOpd = &domain.SasaranOpd{
				Id:         pokinId,
				IdPohon:    pokinId,
				NamaPohon:  namaPohon,
				JenisPohon: jenisPohon,
				LevelPohon: levelPohon,
				TahunPohon: tahunPohon,
				Pelaksana:  make([]domain.PelaksanaPokin, 0),
				SasaranOpd: make([]domain.SasaranOpdDetail, 0),
			}
			pokinMap[pokinId] = sasaranOpd
		}

		// Proses Pelaksana
		if pelaksanaId.Valid && pegawaiId.Valid && pelaksanaNip.Valid && namaPegawai.Valid {
			pelaksanaKey := fmt.Sprintf("%d-%s", pokinId, pelaksanaId.String)
			if !pelaksanaMap[pelaksanaKey] {
				pelaksanaMap[pelaksanaKey] = true
				sasaranOpd.Pelaksana = append(sasaranOpd.Pelaksana, domain.PelaksanaPokin{
					Id:          pelaksanaId.String,
					PegawaiId:   pegawaiId.String,
					Nip:         pelaksanaNip.String,
					NamaPegawai: namaPegawai.String,
				})
			}
		}

		// Proses Sasaran OPD jika ada
		if sasaranId.Valid && namaSasaranOpd.Valid {
			// Cek apakah sasaran OPD sudah ada di slice
			var sasaranExists bool
			var existingSasaran *domain.SasaranOpdDetail

			for i := range sasaranOpd.SasaranOpd {
				if sasaranOpd.SasaranOpd[i].Id == int(sasaranId.Int64) {
					sasaranExists = true
					existingSasaran = &sasaranOpd.SasaranOpd[i]
					break
				}
			}

			if !sasaranExists {
				newSasaran := domain.SasaranOpdDetail{
					Id:             int(sasaranId.Int64),
					IdPohon:        pokinId,
					NamaSasaranOpd: namaSasaranOpd.String,
					TahunAwal:      tahunAwalSasaran.String,
					TahunAkhir:     tahunAkhirSasaran.String,
					JenisPeriode:   jenisPeriodeSasaran.String,
					Indikator:      make([]domain.Indikator, 0),
				}
				sasaranOpd.SasaranOpd = append(sasaranOpd.SasaranOpd, newSasaran)
				existingSasaran = &sasaranOpd.SasaranOpd[len(sasaranOpd.SasaranOpd)-1]
			}

			// Proses Indikator
			if indikatorId.Valid && indikator.Valid {
				var indikatorExists bool
				for i := range existingSasaran.Indikator {
					if existingSasaran.Indikator[i].Id == indikatorId.String {
						indikatorExists = true
						// Update target jika ada
						if targetId.Valid && targetTahun.Valid && targetValue.Valid {
							for j := range existingSasaran.Indikator[i].Target {
								if existingSasaran.Indikator[i].Target[j].Tahun == targetTahun.String {
									existingSasaran.Indikator[i].Target[j] = domain.Target{
										Id:          targetId.String,
										IndikatorId: indikatorId.String,
										Tahun:       targetTahun.String,
										Target:      targetValue.String,
										Satuan:      targetSatuan.String,
									}
									break
								}
							}
						}
						break
					}
				}

				if !indikatorExists {
					newInd := domain.Indikator{
						Id:               indikatorId.String,
						Indikator:        indikator.String,
						RumusPerhitungan: rumusPerhitungan,
						SumberData:       sumberData,
						Target:           make([]domain.Target, 0),
					}

					// Inisialisasi target kosong untuk semua tahun
					tahunAwalInt, _ := strconv.Atoi(tahunAwalSasaran.String)
					tahunAkhirInt, _ := strconv.Atoi(tahunAkhirSasaran.String)

					for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
						targetObj := domain.Target{
							Id:          "",
							IndikatorId: indikatorId.String,
							Tahun:       strconv.Itoa(tahun),
							Target:      "",
							Satuan:      "",
						}

						// Jika ada data target untuk tahun ini, gunakan data tersebut
						if targetId.Valid && targetTahun.Valid && targetValue.Valid &&
							targetTahun.String == strconv.Itoa(tahun) {
							targetObj = domain.Target{
								Id:          targetId.String,
								IndikatorId: indikatorId.String,
								Tahun:       targetTahun.String,
								Target:      targetValue.String,
								Satuan:      targetSatuan.String,
							}
						}

						newInd.Target = append(newInd.Target, targetObj)
					}

					existingSasaran.Indikator = append(existingSasaran.Indikator, newInd)
				}
			}
		}
	}

	// Konversi ke slice
	var result []domain.SasaranOpd
	for _, sasaranOpd := range pokinMap {
		result = append(result, *sasaranOpd)
	}

	return result, nil
}

func (repository *SasaranOpdRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (*domain.SasaranOpd, error) {
	script := `
    SELECT DISTINCT
        pk.id as pokin_id,
        pk.nama_pohon,
        pk.jenis_pohon,
        pk.level_pohon,
        pk.tahun as tahun_pohon,
        pp.id as pelaksana_id,
        pp.pegawai_id,
        p.nip as pelaksana_nip,
        p.nama as nama_pegawai,
        so.id as sasaran_id,
        so.nama_sasaran_opd,
        so.tahun_awal,
        so.tahun_akhir,
        so.jenis_periode,
        i.id as indikator_id,
        i.indikator,
        i.rumus_perhitungan,
        i.sumber_data,
        t.id as target_id,
        t.tahun as target_tahun,
        t.target,
        t.satuan
    FROM tb_sasaran_opd so
    JOIN tb_pohon_kinerja pk ON so.pokin_id = pk.id
    LEFT JOIN tb_pelaksana_pokin pp ON pk.id = pp.pohon_kinerja_id
    LEFT JOIN tb_pegawai p ON pp.pegawai_id = p.id
    LEFT JOIN tb_indikator i ON so.id = i.sasaran_opd_id
    LEFT JOIN tb_target t ON i.id = t.indikator_id
    WHERE so.id = ?`

	rows, err := tx.QueryContext(ctx, script, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sasaranOpd *domain.SasaranOpd
	pelaksanaMap := make(map[string]bool)
	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			pokinId, levelPohon                  int
			namaPohon, jenisPohon, tahunPohon    string
			pelaksanaId, pegawaiId, pelaksanaNip sql.NullString
			namaPegawai                          sql.NullString
			sasaranId                            sql.NullInt64 // Ubah ke NullInt64
			namaSasaranOpd                       sql.NullString
			tahunAwalSasaran, tahunAkhirSasaran  sql.NullString
			jenisPeriodeSasaran                  sql.NullString
			indikatorId, indikator               sql.NullString
			rumusPerhitungan, sumberData         sql.NullString
			targetId, targetTahun                sql.NullString
			targetValue, targetSatuan            sql.NullString
		)

		err := rows.Scan(
			&pokinId, &namaPohon, &jenisPohon, &levelPohon, &tahunPohon,
			&pelaksanaId, &pegawaiId, &pelaksanaNip, &namaPegawai,
			&sasaranId, &namaSasaranOpd,
			&tahunAwalSasaran, &tahunAkhirSasaran, &jenisPeriodeSasaran,
			&indikatorId, &indikator,
			&rumusPerhitungan, &sumberData,
			&targetId, &targetTahun, &targetValue, &targetSatuan,
		)
		if err != nil {
			return nil, err
		}

		// Inisialisasi SasaranOpd jika belum ada
		if sasaranOpd == nil {
			sasaranOpd = &domain.SasaranOpd{
				Id:         pokinId,
				IdPohon:    pokinId,
				NamaPohon:  namaPohon,
				JenisPohon: jenisPohon,
				LevelPohon: levelPohon,
				TahunPohon: tahunPohon,
				Pelaksana:  make([]domain.PelaksanaPokin, 0),
				SasaranOpd: make([]domain.SasaranOpdDetail, 0),
			}

			// Tambahkan SasaranOpdDetail
			sasaranDetail := domain.SasaranOpdDetail{
				Id:             int(sasaranId.Int64), // Konversi ke int
				IdPohon:        pokinId,
				NamaSasaranOpd: namaSasaranOpd.String,
				TahunAwal:      tahunAwalSasaran.String,
				TahunAkhir:     tahunAkhirSasaran.String,
				JenisPeriode:   jenisPeriodeSasaran.String,
				Indikator:      make([]domain.Indikator, 0),
			}
			sasaranOpd.SasaranOpd = append(sasaranOpd.SasaranOpd, sasaranDetail)
		}

		// Proses Pelaksana
		if pelaksanaId.Valid && pegawaiId.Valid && pelaksanaNip.Valid && namaPegawai.Valid {
			pelaksanaKey := fmt.Sprintf("%d-%s", pokinId, pelaksanaId.String)
			if !pelaksanaMap[pelaksanaKey] {
				pelaksanaMap[pelaksanaKey] = true
				sasaranOpd.Pelaksana = append(sasaranOpd.Pelaksana, domain.PelaksanaPokin{
					Id:          pelaksanaId.String,
					PegawaiId:   pegawaiId.String,
					Nip:         pelaksanaNip.String,
					NamaPegawai: namaPegawai.String,
				})
			}
		}

		// Proses Indikator
		if indikatorId.Valid && indikator.Valid {
			ind, exists := indikatorMap[indikatorId.String]
			if !exists {
				ind = &domain.Indikator{
					Id:               indikatorId.String,
					Indikator:        indikator.String,
					RumusPerhitungan: rumusPerhitungan,
					SumberData:       sumberData,
					Target:           make([]domain.Target, 0),
				}

				// Inisialisasi target untuk semua tahun
				tahunAwalInt, _ := strconv.Atoi(tahunAwalSasaran.String)
				tahunAkhirInt, _ := strconv.Atoi(tahunAkhirSasaran.String)

				for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
					tahunStr := strconv.Itoa(tahun)
					ind.Target = append(ind.Target, domain.Target{
						Id:          "",
						IndikatorId: indikatorId.String,
						Tahun:       tahunStr,
						Target:      "",
						Satuan:      "",
					})
				}

				indikatorMap[indikatorId.String] = ind
				sasaranOpd.SasaranOpd[0].Indikator = append(sasaranOpd.SasaranOpd[0].Indikator, *ind)
			}

			// Update target jika ada
			if targetId.Valid && targetTahun.Valid && targetValue.Valid {
				for i := range ind.Target {
					if ind.Target[i].Tahun == targetTahun.String {
						ind.Target[i] = domain.Target{
							Id:          targetId.String,
							IndikatorId: indikatorId.String,
							Tahun:       targetTahun.String,
							Target:      targetValue.String,
							Satuan:      targetSatuan.String,
						}

						// Update target di sasaranOpd
						for j := range sasaranOpd.SasaranOpd[0].Indikator {
							if sasaranOpd.SasaranOpd[0].Indikator[j].Id == indikatorId.String {
								sasaranOpd.SasaranOpd[0].Indikator[j].Target[i] = ind.Target[i]
								break
							}
						}
						break
					}
				}
			}
		}
	}

	if sasaranOpd == nil {
		return nil, errors.New("sasaran opd not found")
	}

	return sasaranOpd, nil
}

func (repository *SasaranOpdRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, sasaranOpd domain.SasaranOpdDetail) error {
	// Insert Sasaran OPD
	scriptSasaran := `INSERT INTO tb_sasaran_opd 
        (id, pokin_id, nama_sasaran_opd, tahun_awal, tahun_akhir, jenis_periode) 
        VALUES (?, ?, ?, ?, ?, ?)`

	_, err := tx.ExecContext(ctx, scriptSasaran,
		sasaranOpd.Id,
		sasaranOpd.IdPohon,
		sasaranOpd.NamaSasaranOpd,
		sasaranOpd.TahunAwal,
		sasaranOpd.TahunAkhir,
		sasaranOpd.JenisPeriode,
	)
	if err != nil {
		return err
	}

	// Insert Indikator
	for _, indikator := range sasaranOpd.Indikator {
		scriptIndikator := `INSERT INTO tb_indikator 
            (id, sasaran_opd_id, indikator, rumus_perhitungan, sumber_data) 
            VALUES (?, ?, ?, ?, ?)`

		_, err = tx.ExecContext(ctx, scriptIndikator,
			indikator.Id,
			sasaranOpd.Id,
			indikator.Indikator,
			indikator.RumusPerhitungan,
			indikator.SumberData,
		)
		if err != nil {
			return err
		}

		// Insert Target
		for _, target := range indikator.Target {
			if target.Target != "" {
				scriptTarget := `INSERT INTO tb_target 
                    (id, indikator_id, tahun, target, satuan) 
                    VALUES (?, ?, ?, ?, ?)`

				_, err = tx.ExecContext(ctx, scriptTarget,
					target.Id,
					indikator.Id,
					target.Tahun,
					target.Target,
					target.Satuan,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (repository *SasaranOpdRepositoryImpl) Update(ctx context.Context, tx *sql.Tx, sasaranOpd domain.SasaranOpdDetail) (domain.SasaranOpdDetail, error) {
	// Update tb_sasaran_opd
	scriptSasaran := `
        UPDATE tb_sasaran_opd 
        SET nama_sasaran_opd = ?, 
            tahun_awal = ?,
            tahun_akhir = ?,
            jenis_periode = ?
        WHERE id = ?`

	_, err := tx.ExecContext(ctx, scriptSasaran,
		sasaranOpd.NamaSasaranOpd,
		sasaranOpd.TahunAwal,
		sasaranOpd.TahunAkhir,
		sasaranOpd.JenisPeriode,
		sasaranOpd.Id)
	if err != nil {
		return sasaranOpd, err
	}

	// Ambil existing indikator dan target
	existingIndikatorMap := make(map[string]bool)
	existingTargetMap := make(map[string]map[string]bool) // map[indikatorId]map[targetId]bool

	// Get existing indikator
	scriptGetIndikator := "SELECT id FROM tb_indikator WHERE sasaran_opd_id = ?"
	rows, err := tx.QueryContext(ctx, scriptGetIndikator, sasaranOpd.Id)
	if err != nil {
		return sasaranOpd, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return sasaranOpd, err
		}
		existingIndikatorMap[id] = true
		existingTargetMap[id] = make(map[string]bool)
	}

	// Get existing target untuk setiap indikator
	for indikatorId := range existingIndikatorMap {
		scriptGetTarget := "SELECT id FROM tb_target WHERE indikator_id = ?"
		targetRows, err := tx.QueryContext(ctx, scriptGetTarget, indikatorId)
		if err != nil {
			return sasaranOpd, err
		}
		defer targetRows.Close()

		for targetRows.Next() {
			var targetId string
			if err := targetRows.Scan(&targetId); err != nil {
				return sasaranOpd, err
			}
			existingTargetMap[indikatorId][targetId] = true
		}
	}

	// Proses setiap indikator
	for _, indikator := range sasaranOpd.Indikator {
		if existingIndikatorMap[indikator.Id] {
			// Update indikator yang sudah ada
			scriptUpdateIndikator := `
                UPDATE tb_indikator 
                SET indikator = ?,
                    rumus_perhitungan = ?,
                    sumber_data = ?
                WHERE id = ? AND sasaran_opd_id = ?`

			_, err := tx.ExecContext(ctx, scriptUpdateIndikator,
				indikator.Indikator,
				indikator.RumusPerhitungan.String,
				indikator.SumberData.String,
				indikator.Id,
				sasaranOpd.Id)
			if err != nil {
				return sasaranOpd, err
			}
		} else {
			// Insert indikator baru
			scriptInsertIndikator := `
                INSERT INTO tb_indikator (id, sasaran_opd_id, indikator, rumus_perhitungan, sumber_data) 
                VALUES (?, ?, ?, ?, ?)`

			_, err := tx.ExecContext(ctx, scriptInsertIndikator,
				indikator.Id,
				sasaranOpd.Id,
				indikator.Indikator,
				indikator.RumusPerhitungan.String,
				indikator.SumberData.String)
			if err != nil {
				return sasaranOpd, err
			}
		}

		// Proses target untuk indikator ini
		for _, target := range indikator.Target {
			if existingTargetMap[indikator.Id][target.Id] {
				// Update target yang sudah ada
				scriptUpdateTarget := `
                    UPDATE tb_target 
                    SET target = ?,
                        satuan = ?,
                        tahun = ?
                    WHERE id = ? AND indikator_id = ?`

				_, err := tx.ExecContext(ctx, scriptUpdateTarget,
					target.Target,
					target.Satuan,
					target.Tahun,
					target.Id,
					indikator.Id)
				if err != nil {
					return sasaranOpd, err
				}
			} else {
				// Insert target baru
				scriptInsertTarget := `
                    INSERT INTO tb_target (id, indikator_id, tahun, target, satuan)
                    VALUES (?, ?, ?, ?, ?)`

				_, err := tx.ExecContext(ctx, scriptInsertTarget,
					target.Id,
					indikator.Id,
					target.Tahun,
					target.Target,
					target.Satuan)
				if err != nil {
					return sasaranOpd, err
				}
			}
		}

		// Hapus target yang tidak ada dalam request
		if targetMap, exists := existingTargetMap[indikator.Id]; exists {
			for existingTargetId := range targetMap {
				found := false
				for _, target := range indikator.Target {
					if target.Id == existingTargetId {
						found = true
						break
					}
				}
				if !found {
					scriptDeleteTarget := "DELETE FROM tb_target WHERE id = ? AND indikator_id = ?"
					_, err = tx.ExecContext(ctx, scriptDeleteTarget, existingTargetId, indikator.Id)
					if err != nil {
						return sasaranOpd, err
					}
				}
			}
		}
	}

	// Hapus indikator yang tidak ada dalam request
	for existingId := range existingIndikatorMap {
		found := false
		for _, indikator := range sasaranOpd.Indikator {
			if indikator.Id == existingId {
				found = true
				break
			}
		}
		if !found {
			// Hapus semua target untuk indikator ini
			scriptDeleteTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
			_, err = tx.ExecContext(ctx, scriptDeleteTargets, existingId)
			if err != nil {
				return sasaranOpd, err
			}

			// Kemudian hapus indikator
			scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE id = ?"
			_, err = tx.ExecContext(ctx, scriptDeleteIndikator, existingId)
			if err != nil {
				return sasaranOpd, err
			}
		}
	}

	return sasaranOpd, nil
}

func (repository *SasaranOpdRepositoryImpl) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	// Delete targets first (cascade)
	scriptDeleteTargets := `DELETE t FROM tb_target t 
                           INNER JOIN tb_indikator i ON t.indikator_id = i.id 
                           WHERE i.sasaran_opd_id = ?`
	_, err := tx.ExecContext(ctx, scriptDeleteTargets, id)
	if err != nil {
		return err
	}

	// Delete indikators
	scriptDeleteIndikators := `DELETE FROM tb_indikator WHERE sasaran_opd_id = ?`
	_, err = tx.ExecContext(ctx, scriptDeleteIndikators, id)
	if err != nil {
		return err
	}

	// Delete sasaran opd
	scriptDeleteSasaran := `DELETE FROM tb_sasaran_opd WHERE id = ?`
	result, err := tx.ExecContext(ctx, scriptDeleteSasaran, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("sasaran opd not found")
	}

	return nil
}

func (repository *SasaranOpdRepositoryImpl) FindByIdSasaran(ctx context.Context, tx *sql.Tx, id int) (*domain.SasaranOpdDetail, error) {
	fmt.Printf("Repository FindByIdSasaran - Query untuk ID: %d\n", id)

	// Query untuk mendapatkan data sasaran OPD
	scriptSasaran := `
    SELECT 
        CAST(so.id AS CHAR) as id, 
        so.pokin_id,
        so.nama_sasaran_opd,
        so.tahun_awal,
        so.tahun_akhir,
        so.jenis_periode
    FROM tb_sasaran_opd so
    WHERE so.id = ?`

	var sasaranOpd domain.SasaranOpdDetail
	err := tx.QueryRowContext(ctx, scriptSasaran, id).Scan(
		&sasaranOpd.Id, // Sekarang akan menerima string
		&sasaranOpd.IdPohon,
		&sasaranOpd.NamaSasaranOpd,
		&sasaranOpd.TahunAwal,
		&sasaranOpd.TahunAkhir,
		&sasaranOpd.JenisPeriode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("sasaran opd not found")
		}
		return nil, fmt.Errorf("error scanning sasaran opd: %v", err)
	}

	// Query untuk mendapatkan indikator dan target
	scriptIndikatorTarget := `
    SELECT 
        i.id,
        i.indikator,
        i.rumus_perhitungan,
        i.sumber_data,
        t.id,
        t.tahun,
        t.target,
        t.satuan
    FROM tb_indikator i
    LEFT JOIN tb_target t ON i.id = t.indikator_id
    WHERE i.sasaran_opd_id = ?`

	rows, err := tx.QueryContext(ctx, scriptIndikatorTarget, id)
	if err != nil {
		return nil, fmt.Errorf("error querying indikator: %v", err)
	}
	defer rows.Close()

	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			indikatorId, indikator          string
			rumusPerhitungan, sumberData    sql.NullString
			targetId, tahun, target, satuan sql.NullString
		)

		err := rows.Scan(
			&indikatorId,
			&indikator,
			&rumusPerhitungan,
			&sumberData,
			&targetId,
			&tahun,
			&target,
			&satuan,
		)
		if err != nil {
			return nil, err
		}

		// Cek apakah indikator sudah ada di map
		ind, exists := indikatorMap[indikatorId]
		if !exists {
			ind = &domain.Indikator{
				Id:               indikatorId,
				Indikator:        indikator,
				RumusPerhitungan: rumusPerhitungan,
				SumberData:       sumberData,
				Target:           make([]domain.Target, 0),
			}
			indikatorMap[indikatorId] = ind
		}

		// Tambahkan target jika ada
		if targetId.Valid && tahun.Valid {
			target := domain.Target{
				Id:          targetId.String,
				IndikatorId: indikatorId,
				Tahun:       tahun.String,
				Target:      target.String,
				Satuan:      satuan.String,
			}
			ind.Target = append(ind.Target, target)
		}
	}

	// Convert map ke slice
	sasaranOpd.Indikator = make([]domain.Indikator, 0, len(indikatorMap))
	for _, ind := range indikatorMap {
		sasaranOpd.Indikator = append(sasaranOpd.Indikator, *ind)
	}

	return &sasaranOpd, nil
}

// ini sudah bisa kurang
func (repository *SasaranOpdRepositoryImpl) FindByIdPokin(ctx context.Context, tx *sql.Tx, idPokin int, tahun string) (*domain.SasaranOpd, error) {
	// Query dimodifikasi untuk validasi dengan tb_periode dan mengambil data dari tb_indikator
	query := `
    WITH target_data AS (
        SELECT 
            id,
            indikator_id,
            tahun,
            target,
            satuan
        FROM tb_target 
        WHERE tahun = ?
    )
    SELECT DISTINCT
        pk.id as pokin_id,
        pk.nama_pohon,
        pk.jenis_pohon,
        pk.level_pohon,
        pk.tahun as tahun_pohon,
        pp.id as pelaksana_id,
        pp.pegawai_id,
        p.nip as pelaksana_nip,
        p.nama as nama_pegawai,
        so.id as sasaran_id,
        so.nama_sasaran_opd,
        so.tahun_awal,
        so.tahun_akhir,
        so.jenis_periode,
        i.id as indikator_id,
        i.indikator,
        i.rumus_perhitungan,
        i.sumber_data,
        t.id as target_id,
        t.tahun as target_tahun,
        t.target,
        t.satuan
    FROM tb_pohon_kinerja pk
    LEFT JOIN tb_pelaksana_pokin pp ON pk.id = pp.pohon_kinerja_id
    LEFT JOIN tb_pegawai p ON pp.pegawai_id = p.id
    LEFT JOIN tb_sasaran_opd so ON pk.id = so.pokin_id
    INNER JOIN tb_periode per ON (so.tahun_awal = per.tahun_awal AND so.tahun_akhir = per.tahun_akhir)
    LEFT JOIN tb_indikator i ON so.id = i.sasaran_opd_id
    LEFT JOIN target_data t ON i.id = t.indikator_id
    WHERE pk.id = ?
    ORDER BY so.nama_sasaran_opd ASC, i.id ASC`

	rows, err := tx.QueryContext(ctx, query, tahun, idPokin)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	sasaranOpd := &domain.SasaranOpd{}
	pelaksanaMap := make(map[string]bool)
	sasaranMap := make(map[int]*domain.SasaranOpdDetail)
	indikatorMap := make(map[string]*domain.Indikator)
	firstRow := true

	for rows.Next() {
		var (
			pokinId                                           int
			namaPohon, jenisPohon                             string
			levelPohon                                        int
			tahunPohon                                        string
			pelaksanaId, pegawaiId, pelaksanaNip, namaPegawai sql.NullString
			sasaranId                                         sql.NullInt64
			namaSasaran, tahunAwal, tahunAkhir, jenisPeriode  sql.NullString
			indikatorId, indikatorNama                        sql.NullString
			rumusPerhitungan, sumberData                      sql.NullString
			targetId, targetTahun, targetValue, targetSatuan  sql.NullString
		)

		if err := rows.Scan(
			&pokinId, &namaPohon, &jenisPohon, &levelPohon, &tahunPohon,
			&pelaksanaId, &pegawaiId, &pelaksanaNip, &namaPegawai,
			&sasaranId, &namaSasaran, &tahunAwal, &tahunAkhir, &jenisPeriode,
			&indikatorId, &indikatorNama, &rumusPerhitungan, &sumberData,
			&targetId, &targetTahun, &targetValue, &targetSatuan,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Set pohon kinerja info pada baris pertama
		if firstRow {
			sasaranOpd.IdPohon = pokinId
			sasaranOpd.NamaPohon = namaPohon
			sasaranOpd.JenisPohon = jenisPohon
			sasaranOpd.LevelPohon = levelPohon
			sasaranOpd.TahunPohon = tahunPohon
			firstRow = false
		}

		// Process Pelaksana
		if pelaksanaId.Valid && !pelaksanaMap[pelaksanaId.String] {
			pelaksanaMap[pelaksanaId.String] = true
			sasaranOpd.Pelaksana = append(sasaranOpd.Pelaksana, domain.PelaksanaPokin{
				Id:          pelaksanaId.String,
				PegawaiId:   pegawaiId.String,
				Nip:         pelaksanaNip.String,
				NamaPegawai: namaPegawai.String,
			})
		}

		// Process Sasaran OPD
		if sasaranId.Valid {
			sasaranIdInt := int(sasaranId.Int64)
			sasaran, exists := sasaranMap[sasaranIdInt]
			if !exists {
				sasaran = &domain.SasaranOpdDetail{
					Id:             sasaranIdInt,
					IdPohon:        pokinId,
					NamaSasaranOpd: namaSasaran.String,
					TahunAwal:      tahunAwal.String,
					TahunAkhir:     tahunAkhir.String,
					JenisPeriode:   jenisPeriode.String,
					Indikator:      make([]domain.Indikator, 0),
				}
				sasaranMap[sasaranIdInt] = sasaran
			}

			// Process Indikator
			if indikatorId.Valid {
				indikator, exists := indikatorMap[indikatorId.String]
				if !exists {
					// Buat indikator baru
					indikator = &domain.Indikator{
						Id:               indikatorId.String,
						SasaranOpdId:     sasaranIdInt,
						Indikator:        indikatorNama.String,
						RumusPerhitungan: rumusPerhitungan,
						SumberData:       sumberData,
						Target:           make([]domain.Target, 0),
					}

					// Add empty target by default
					target := domain.Target{
						Id:          "",
						IndikatorId: indikatorId.String,
						Tahun:       tahun,
						Target:      "",
						Satuan:      "",
					}

					// Update target if exists
					if targetId.Valid {
						target.Id = targetId.String
						target.Target = targetValue.String
						target.Satuan = targetSatuan.String
					}

					indikator.Target = append(indikator.Target, target)
					indikatorMap[indikatorId.String] = indikator
					sasaran.Indikator = append(sasaran.Indikator, *indikator)
				} else {
					// Jika indikator sudah ada, update target jika diperlukan
					if targetId.Valid {
						target := domain.Target{
							Id:          targetId.String,
							IndikatorId: indikatorId.String,
							Tahun:       targetTahun.String,
							Target:      targetValue.String,
							Satuan:      targetSatuan.String,
						}
						indikator.Target = append(indikator.Target, target)
					}
				}
			}
		}
	}

	// Convert maps to slices and sort by nama_sasaran_opd
	var sasaranOpdSlice []domain.SasaranOpdDetail
	for _, sasaran := range sasaranMap {
		sasaranOpdSlice = append(sasaranOpdSlice, *sasaran)
	}

	// Urutkan slice berdasarkan nama_sasaran_opd
	sort.Slice(sasaranOpdSlice, func(i, j int) bool {
		return sasaranOpdSlice[i].NamaSasaranOpd < sasaranOpdSlice[j].NamaSasaranOpd
	})

	sasaranOpd.SasaranOpd = sasaranOpdSlice

	return sasaranOpd, nil
}

// sek iki lali opo
func (repository *SasaranOpdRepositoryImpl) FindIdPokinSasaran(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error) {
	scriptPokin := `
    SELECT DISTINCT
        pk.id, 
        pk.parent, 
        pk.nama_pohon, 
        pk.jenis_pohon, 
        pk.level_pohon, 
        pk.kode_opd, 
        pk.keterangan, 
        pk.tahun,
        pk.status,
        i.id as indikator_id,
        i.indikator as nama_indikator,
        t.id as target_id,
        t.target,
        t.satuan,
        t.tahun as tahun_target
    FROM 
        tb_pohon_kinerja pk 
        LEFT JOIN tb_indikator i ON pk.id = i.pokin_id
        LEFT JOIN tb_target t ON i.id = t.indikator_id
    WHERE 
        pk.id = ?
    ORDER BY t.id DESC
    LIMIT 1`

	rows, err := tx.QueryContext(ctx, scriptPokin, id)
	if err != nil {
		return domain.PohonKinerja{}, fmt.Errorf("error querying pohon kinerja: %v", err)
	}
	defer rows.Close()

	var pohonKinerja domain.PohonKinerja
	indikatorMap := make(map[string]*domain.Indikator)
	dataFound := false

	for rows.Next() {
		var (
			indikatorId, namaIndikator            sql.NullString
			targetId, target, satuan, tahunTarget sql.NullString
		)

		err := rows.Scan(
			&pohonKinerja.Id,
			&pohonKinerja.Parent,
			&pohonKinerja.NamaPohon,
			&pohonKinerja.JenisPohon,
			&pohonKinerja.LevelPohon,
			&pohonKinerja.KodeOpd,
			&pohonKinerja.Keterangan,
			&pohonKinerja.Tahun,
			&pohonKinerja.Status,
			&indikatorId,
			&namaIndikator,
			&targetId,
			&target,
			&satuan,
			&tahunTarget,
		)
		if err != nil {
			return domain.PohonKinerja{}, fmt.Errorf("error scanning row: %v", err)
		}

		dataFound = true

		if indikatorId.Valid && namaIndikator.Valid {
			ind := &domain.Indikator{
				Id:        indikatorId.String,
				Indikator: namaIndikator.String,
				PokinId:   fmt.Sprint(pohonKinerja.Id),
				Target:    []domain.Target{},
			}

			if targetId.Valid && target.Valid && satuan.Valid {
				targetObj := domain.Target{
					Id:          targetId.String,
					IndikatorId: indikatorId.String,
					Target:      target.String,
					Satuan:      satuan.String,
					Tahun:       tahunTarget.String,
				}
				ind.Target = append(ind.Target, targetObj)
			}

			indikatorMap[indikatorId.String] = ind
			pohonKinerja.Indikator = append(pohonKinerja.Indikator, *ind)
		}
	}

	if !dataFound {
		return domain.PohonKinerja{}, fmt.Errorf("pohon kinerja with id %d not found", id)
	}

	return pohonKinerja, nil
}

func (repository *SasaranOpdRepositoryImpl) GetByTahun(ctx context.Context, tx *sql.Tx, tahun string, kodeOpd string) ([]domain.SasaranOpdTahunan, error) {
	query := `
	SELECT so.id, so.pokin_id, tp.kode_opd, opd.nama_opd, so.nama_sasaran_opd, so.tahun_awal, so.tahun_akhir, so.jenis_periode, tp.jenis_pohon, tp.is_active
	FROM tb_sasaran_opd so
	JOIN tb_pohon_kinerja tp ON tp.id = so.pokin_id
    LEFT JOIN tb_operasional_daerah opd ON opd.kode_opd = tp.kode_opd
	WHERE CAST(so.tahun_awal AS UNSIGNED) <= ?
	AND CAST(so.tahun_akhir AS UNSIGNED) >= ?
	AND tp.kode_opd = ?
	AND tp.is_active IS TRUE
`

	rows, err := tx.QueryContext(ctx, query, tahun, tahun, kodeOpd)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var results []domain.SasaranOpdTahunan
	for rows.Next() {
		var sasOpd domain.SasaranOpdTahunan

		err := rows.Scan(
			&sasOpd.Id,
			&sasOpd.IdPohon,
			&sasOpd.KodeOpd,
			&sasOpd.NamaOpd,
			&sasOpd.SasaranOpd,
			&sasOpd.TahunAwal,
			&sasOpd.TahunAkhir,
			&sasOpd.JenisPeriode,
			&sasOpd.JenisPohon,
			&sasOpd.PohonAktif,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, sasOpd)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *SasaranOpdRepositoryImpl) GetIndikatorSasaranOpdByTahun(ctx context.Context, tx *sql.Tx, tahun string, kodeOpd string) ([]domain.Indikator, error) {
	query := `
	SELECT ind.id, ind.sasaran_opd_id, ind.indikator, ind.sumber_data, ind.rumus_perhitungan,
		tg.id, tg.indikator_id, tg.target, tg.satuan, tg.tahun
	FROM tb_indikator ind
	LEFT JOIN tb_target tg ON tg.indikator_id = ind.id AND tg.tahun = ?
	JOIN tb_sasaran_opd sp ON ind.sasaran_opd_id = sp.id
		AND CAST(sp.tahun_awal AS UNSIGNED) <= ?
		AND CAST(sp.tahun_akhir AS UNSIGNED) >= ?
	JOIN tb_pohon_kinerja tp ON tp.id = sp.pokin_id
		AND tp.kode_opd = ?
		AND tp.is_active IS TRUE
	ORDER BY tg.tahun;
	`

	rows, err := tx.QueryContext(ctx, query, tahun, tahun, tahun, kodeOpd)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	indikatorMap := make(map[string]*domain.Indikator)
	for rows.Next() {
		var (
			indId, indikatorStr                                      string
			sasaranOpdId                                             int
			sumberData, rumusPerhitungan                             sql.NullString
			targetId, indikatorId, targetStr, satuanStr, targetTahun sql.NullString
		)

		err := rows.Scan(
			&indId, &sasaranOpdId, &indikatorStr, &sumberData, &rumusPerhitungan,
			&targetId, &indikatorId, &targetStr, &satuanStr, &targetTahun,
		)
		if err != nil {
			return nil, err
		}

		// cek indikator ada di map ??
		indikator, exists := indikatorMap[indId]
		if !exists {
			indikator = &domain.Indikator{
				Id:               indId,
				SasaranOpdId:     sasaranOpdId,
				Indikator:        indikatorStr,
				SumberData:       sumberData,
				RumusPerhitungan: rumusPerhitungan,
				Target:           []domain.Target{},
			}
			indikatorMap[indId] = indikator
		}

		// tambah target jika ada
		if targetId.Valid {
			indikator.Target = append(indikator.Target, domain.Target{
				Id:          targetId.String,
				IndikatorId: indikatorId.String,
				Target:      targetStr.String,
				Satuan:      satuanStr.String,
				Tahun:       targetTahun.String,
			})
		}
	}

	result := make([]domain.Indikator, 0, len(indikatorMap))
	for _, indikator := range indikatorMap {
		result = append(result, *indikator)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
