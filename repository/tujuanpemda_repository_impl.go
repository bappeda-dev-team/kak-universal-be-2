package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
	"fmt"
	"log"
	"sort"
	"strconv"
)

type TujuanPemdaRepositoryImpl struct {
}

func NewTujuanPemdaRepositoryImpl() *TujuanPemdaRepositoryImpl {
	return &TujuanPemdaRepositoryImpl{}
}

func (repository *TujuanPemdaRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, tujuanPemda domain.TujuanPemda) (domain.TujuanPemda, error) {
	query := "INSERT INTO tb_tujuan_pemda(id, tujuan_pemda, tematik_id, periode_id, tahun_awal_periode, tahun_akhir_periode, jenis_periode, id_visi, id_misi) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, tujuanPemda.Id, tujuanPemda.TujuanPemda, tujuanPemda.TematikId, tujuanPemda.PeriodeId, tujuanPemda.TahunAwalPeriode, tujuanPemda.TahunAkhirPeriode, tujuanPemda.JenisPeriode, tujuanPemda.IdVisi, tujuanPemda.IdMisi)
	if err != nil {
		return tujuanPemda, err
	}
	return tujuanPemda, nil
}

func (repository *TujuanPemdaRepositoryImpl) CreateIndikator(ctx context.Context, tx *sql.Tx, indikator domain.Indikator) (domain.Indikator, error) {
	query := "INSERT INTO tb_indikator(id, tujuan_pemda_id, indikator, rumus_perhitungan, sumber_data) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, indikator.Id, indikator.TujuanPemdaId, indikator.Indikator, indikator.RumusPerhitungan, indikator.SumberData)
	if err != nil {
		return indikator, err
	}
	return indikator, nil
}

func (repository *TujuanPemdaRepositoryImpl) CreateTarget(ctx context.Context, tx *sql.Tx, target domain.Target) error {
	query := "INSERT INTO tb_target(id, indikator_id, target, satuan, tahun) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, target.Id, target.IndikatorId, target.Target, target.Satuan, target.Tahun)
	return err
}

func (repository *TujuanPemdaRepositoryImpl) Update(ctx context.Context, tx *sql.Tx, tujuanPemda domain.TujuanPemda) (domain.TujuanPemda, error) {
	// Update tujuan pemda
	query := "UPDATE tb_tujuan_pemda SET tujuan_pemda = ?, tematik_id = ?, periode_id = ?, tahun_awal_periode = ?, tahun_akhir_periode = ?, jenis_periode = ?, id_visi = ?, id_misi = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, query, tujuanPemda.TujuanPemda, tujuanPemda.TematikId, tujuanPemda.PeriodeId, tujuanPemda.TahunAwalPeriode, tujuanPemda.TahunAkhirPeriode, tujuanPemda.JenisPeriode, tujuanPemda.IdVisi, tujuanPemda.IdMisi, tujuanPemda.Id)
	if err != nil {
		return tujuanPemda, err
	}

	// Hapus semua indikator lama beserta targetnya
	scriptDeleteOldIndicators := "DELETE FROM tb_indikator WHERE tujuan_pemda_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeleteOldIndicators, tujuanPemda.Id)
	if err != nil {
		return tujuanPemda, err
	}

	// Insert indikator baru
	for _, indikator := range tujuanPemda.Indikator {
		scriptInsertIndikator := `
            INSERT INTO tb_indikator 
                (id, tujuan_pemda_id, indikator, rumus_perhitungan, sumber_data) 
            VALUES 
                (?, ?, ?, ?, ?)`

		_, err := tx.ExecContext(ctx, scriptInsertIndikator,
			indikator.Id,
			tujuanPemda.Id,
			indikator.Indikator,
			indikator.RumusPerhitungan,
			indikator.SumberData)
		if err != nil {
			return tujuanPemda, err
		}

		// Hapus semua target lama untuk indikator ini
		scriptDeleteOldTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
		_, err = tx.ExecContext(ctx, scriptDeleteOldTargets, indikator.Id)
		if err != nil {
			return tujuanPemda, err
		}

		// Insert target baru
		for _, target := range indikator.Target {
			scriptInsertTarget := `
                INSERT INTO tb_target 
                    (id, indikator_id, target, satuan, tahun)
                VALUES 
                    (?, ?, ?, ?, ?)`

			_, err := tx.ExecContext(ctx, scriptInsertTarget,
				target.Id,
				indikator.Id,
				target.Target,
				target.Satuan,
				target.Tahun)
			if err != nil {
				return tujuanPemda, err
			}
		}
	}

	return tujuanPemda, nil
}

func (repository *TujuanPemdaRepositoryImpl) Delete(ctx context.Context, tx *sql.Tx, tujuanPemdaId int) error {
	// 1. Hapus target berdasarkan indikator yang terkait dengan tujuan pemda
	queryDeleteTarget := `
        DELETE t FROM tb_target t
        INNER JOIN tb_indikator i ON t.indikator_id = i.id
        WHERE i.tujuan_pemda_id = ?`
	_, err := tx.ExecContext(ctx, queryDeleteTarget, tujuanPemdaId)
	if err != nil {
		return err
	}

	// 2. Hapus indikator yang terkait dengan tujuan pemda
	queryDeleteIndikator := "DELETE FROM tb_indikator WHERE tujuan_pemda_id = ?"
	_, err = tx.ExecContext(ctx, queryDeleteIndikator, tujuanPemdaId)
	if err != nil {
		return err
	}

	// 3. Hapus tujuan pemda
	queryDeleteTujuanPemda := "DELETE FROM tb_tujuan_pemda WHERE id = ?"
	_, err = tx.ExecContext(ctx, queryDeleteTujuanPemda, tujuanPemdaId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *TujuanPemdaRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, tujuanPemdaId int) (domain.TujuanPemda, error) {
	query := `
        SELECT DISTINCT
            tp.id,
            tp.tujuan_pemda,
            tp.tematik_id,
            tp.periode_id,
            tp.tahun_awal_periode,
            tp.tahun_akhir_periode,
            tp.jenis_periode,
            tp.id_visi,
            tp.id_misi,
            pk.jenis_pohon, 
            i.id as indikator_id,
            i.indikator as indikator_text,
            i.rumus_perhitungan,
            i.sumber_data,
            t.id as target_id,
            t.target as target_value,
            t.satuan as target_satuan,
            t.tahun as target_tahun
        FROM 
            tb_tujuan_pemda tp
            LEFT JOIN tb_pohon_kinerja pk ON tp.tematik_id = pk.id 
            LEFT JOIN tb_indikator i ON tp.id = i.tujuan_pemda_id
            LEFT JOIN tb_target t ON t.indikator_id = i.id
        WHERE tp.id = ?
        ORDER BY 
            tp.id, i.id, CAST(t.tahun AS SIGNED)`

	rows, err := tx.QueryContext(ctx, query, tujuanPemdaId)
	if err != nil {
		return domain.TujuanPemda{}, fmt.Errorf("error querying tujuan pemda: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}(rows)

	var result domain.TujuanPemda
	var firstRow = true
	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			id, tematikId, periodeId, idVisi, idMisi                         int
			tujuanPemdaText, tahunAwal, tahunAkhir, jenisPeriode, jenisPohon string
			indikatorId, indikatorText                                       sql.NullString
			rumusPerhitunganNull, sumberDataNull                             sql.NullString
			targetId, targetValue, targetSatuan, targetTahun                 sql.NullString
		)

		err := rows.Scan(
			&id,
			&tujuanPemdaText,
			&tematikId,
			&periodeId,
			&tahunAwal,
			&tahunAkhir,
			&jenisPeriode,
			&idVisi,
			&idMisi,
			&jenisPohon,
			&indikatorId,
			&indikatorText,
			&rumusPerhitunganNull,
			&sumberDataNull,
			&targetId,
			&targetValue,
			&targetSatuan,
			&targetTahun,
		)
		if err != nil {
			return domain.TujuanPemda{}, fmt.Errorf("error scanning row: %v", err)
		}

		if firstRow {
			result = domain.TujuanPemda{
				Id:                id,
				TujuanPemda:       tujuanPemdaText,
				TematikId:         tematikId,
				JenisPohon:        jenisPohon,
				PeriodeId:         periodeId,
				IdVisi:            idVisi,
				IdMisi:            idMisi,
				TahunAwalPeriode:  tahunAwal,
				TahunAkhirPeriode: tahunAkhir,
				JenisPeriode:      jenisPeriode,
				Periode: domain.Periode{
					TahunAwal:    tahunAwal,
					TahunAkhir:   tahunAkhir,
					JenisPeriode: jenisPeriode,
				},
				Indikator: []domain.Indikator{},
			}
			firstRow = false
		}

		if indikatorId.Valid && indikatorText.Valid {
			currentIndikator, exists := indikatorMap[indikatorId.String]

			if !exists {
				// Konversi NullString ke string biasa, jika null maka gunakan string kosong
				rumusPerhitungan := ""
				if rumusPerhitunganNull.Valid {
					rumusPerhitungan = rumusPerhitunganNull.String
				}

				sumberData := ""
				if sumberDataNull.Valid {
					sumberData = sumberDataNull.String
				}

				indikator := domain.Indikator{
					Id:               indikatorId.String,
					TujuanPemdaId:    id,
					Indikator:        indikatorText.String,
					RumusPerhitungan: sql.NullString{String: rumusPerhitungan, Valid: rumusPerhitungan != ""},
					SumberData:       sql.NullString{String: sumberData, Valid: sumberData != ""},
					Target:           []domain.Target{},
				}

				// Generate target untuk setiap tahun dalam periode
				if tahunAwal != "" && tahunAkhir != "" {
					tahunAwalInt, errAwal := strconv.Atoi(tahunAwal)
					tahunAkhirInt, errAkhir := strconv.Atoi(tahunAkhir)

					if errAwal == nil && errAkhir == nil {
						for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
							indikator.Target = append(indikator.Target, domain.Target{
								Id:          "-",
								IndikatorId: indikatorId.String,
								Target:      "-",
								Satuan:      "-",
								Tahun:       strconv.Itoa(tahun),
							})
						}
					}
				}

				result.Indikator = append(result.Indikator, indikator)
				indikatorMap[indikatorId.String] = &result.Indikator[len(result.Indikator)-1]
				currentIndikator = &result.Indikator[len(result.Indikator)-1]
			}

			// Update target yang ada dengan data sebenarnya
			if targetId.Valid && targetValue.Valid && targetTahun.Valid {
				for i := range currentIndikator.Target {
					if currentIndikator.Target[i].Tahun == targetTahun.String {
						currentIndikator.Target[i] = domain.Target{
							Id:          targetId.String,
							IndikatorId: indikatorId.String,
							Target:      targetValue.String,
							Satuan:      targetSatuan.String,
							Tahun:       targetTahun.String,
						}
						break
					}
				}
			}
		}
	}

	if err = rows.Err(); err != nil {
		return domain.TujuanPemda{}, fmt.Errorf("error iterating rows: %v", err)
	}

	if result.Id == 0 {
		return domain.TujuanPemda{}, fmt.Errorf("tujuan pemda dengan id %d tidak ditemukan", tujuanPemdaId)
	}

	return result, nil
}

func (repository *TujuanPemdaRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, tahun string, jenisPeriode string) ([]domain.TujuanPemda, error) {
	query := `
        SELECT DISTINCT
            tp.id,
            tp.tujuan_pemda,
            tp.tematik_id,
            tp.periode_id,
            p.tahun_awal,
            p.tahun_akhir,
            p.jenis_periode
        FROM 
            tb_tujuan_pemda tp
            INNER JOIN tb_periode p ON tp.periode_id = p.id
            INNER JOIN tb_pohon_kinerja pk ON tp.tematik_id = pk.id
        WHERE 
            CAST(? AS SIGNED) BETWEEN CAST(p.tahun_awal AS SIGNED) AND CAST(p.tahun_akhir AS SIGNED)
            AND p.jenis_periode = ?
            AND pk.is_active = true
            AND pk.level_pohon = 0
        ORDER BY 
            tp.id`

	rows, err := tx.QueryContext(ctx, query, tahun, jenisPeriode)
	if err != nil {
		return nil, fmt.Errorf("error querying tujuan pemda: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}(rows)

	var result []domain.TujuanPemda

	for rows.Next() {
		var tujuanPemda domain.TujuanPemda
		err := rows.Scan(
			&tujuanPemda.Id,
			&tujuanPemda.TujuanPemda,
			&tujuanPemda.TematikId,
			&tujuanPemda.PeriodeId,
			&tujuanPemda.Periode.TahunAwal,
			&tujuanPemda.Periode.TahunAkhir,
			&tujuanPemda.Periode.JenisPeriode,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		result = append(result, tujuanPemda)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return result, nil
}

func (repository *TujuanPemdaRepositoryImpl) DeleteIndikator(ctx context.Context, tx *sql.Tx, tujuanPemdaId int) error {
	query := "DELETE FROM tb_indikator WHERE tujuan_pemda_id = ?"
	_, err := tx.ExecContext(ctx, query, tujuanPemdaId)
	return err
}

func (repository *TujuanPemdaRepositoryImpl) IsIdExists(ctx context.Context, tx *sql.Tx, id int) bool {
	query := "SELECT COUNT(*) FROM tb_tujuan_pemda WHERE id = ?"
	var count int
	err := tx.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return true
	}
	return count > 0
}

func (repository *TujuanPemdaRepositoryImpl) UpdatePeriode(ctx context.Context, tx *sql.Tx, tujuanPemda domain.TujuanPemda) (domain.TujuanPemda, error) {
	// Update hanya periode_id
	query := "UPDATE tb_tujuan_pemda SET periode_id = ? WHERE id = ?"
	result, err := tx.ExecContext(ctx, query, tujuanPemda.PeriodeId, tujuanPemda.Id)
	if err != nil {
		return domain.TujuanPemda{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.TujuanPemda{}, err
	}

	if rowsAffected == 0 {
		return domain.TujuanPemda{}, fmt.Errorf("periode dengan id %d sudah digunakan", tujuanPemda.PeriodeId)
	}

	// Ambil data terbaru setelah update
	query = `
        SELECT 
            tp.id,
            tp.tujuan_pemda,
            tp.tematik_id,
            tp.periode_id,
            COALESCE(p.tahun_awal, 'Pilih periode') as tahun_awal,
            COALESCE(p.tahun_akhir, 'Pilih periode') as tahun_akhir
        FROM 
            tb_tujuan_pemda tp
            LEFT JOIN tb_periode p ON tp.periode_id = p.id
        WHERE tp.id = ?`

	var updatedTujuanPemda domain.TujuanPemda
	err = tx.QueryRowContext(ctx, query, tujuanPemda.Id).Scan(
		&updatedTujuanPemda.Id,
		&updatedTujuanPemda.TujuanPemda,
		&updatedTujuanPemda.TematikId,
		&updatedTujuanPemda.PeriodeId,
		&updatedTujuanPemda.Periode.TahunAwal,
		&updatedTujuanPemda.Periode.TahunAkhir,
	)
	if err != nil {
		return domain.TujuanPemda{}, fmt.Errorf("gagal mengambil data setelah update: %v", err)
	}

	return updatedTujuanPemda, nil
}

func (repository *TujuanPemdaRepositoryImpl) FindAllWithPokin(ctx context.Context, tx *sql.Tx, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.TujuanPemdaWithPokin, error) {
	validateQuery := `SELECT EXISTS (
        SELECT 1 FROM tb_periode 
        WHERE tahun_awal = ? 
        AND tahun_akhir = ? 
        AND jenis_periode = ?
    )`
	var exists bool
	err := tx.QueryRowContext(ctx, validateQuery, tahunAwal, tahunAkhir, jenisPeriode).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("periode dengan tahun awal %s, tahun akhir %s, dan jenis periode %s tidak ditemukan",
			tahunAwal, tahunAkhir, jenisPeriode)
	}

	query := `
      SELECT 
        pk.id as pokin_id,
        pk.nama_pohon,
        pk.jenis_pohon,
        pk.level_pohon,
		pk.is_active,
        pk.kode_opd,
        pk.keterangan,
        pk.tahun as tahun_pokin,
        tp.id as tujuan_id,
        tp.tujuan_pemda,
        tp.id_visi,
        tp.id_misi,
        tp.tahun_awal_periode,
        tp.tahun_akhir_periode,
        tp.jenis_periode,
        i.id as indikator_id,
        i.indikator,
        i.rumus_perhitungan,
        i.sumber_data,
        t.id as target_id,
        t.target,
        t.satuan,
        t.tahun
    FROM 
        tb_pohon_kinerja pk
    LEFT JOIN 
        tb_tujuan_pemda tp ON pk.id = tp.tematik_id
        AND tp.tahun_awal_periode = ? 
        AND tp.tahun_akhir_periode = ? 
        AND tp.jenis_periode = ?
    LEFT JOIN 
        tb_indikator i ON tp.id = i.tujuan_pemda_id
    LEFT JOIN 
        tb_target t ON i.id = t.indikator_id
        AND CAST(t.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)
    WHERE 
        pk.level_pohon = 0
        AND CAST(pk.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)
    ORDER BY 
        pk.id, tp.id, i.id, t.tahun`

	rows, err := tx.QueryContext(ctx, query,
		tahunAwal, tahunAkhir, jenisPeriode,
		tahunAwal, tahunAkhir,
		tahunAwal, tahunAkhir)
	if err != nil {
		return nil, fmt.Errorf("error querying data: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}(rows)

	pokinMap := make(map[int]*domain.TujuanPemdaWithPokin)
	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			pokinId                                                int
			namaPohon, jenisPohon, kodeOpd, keterangan, tahunPokin string
			levelPohon                                             int
			is_active                                              bool
			tujuanId                                               sql.NullInt64
			tujuanPemda                                            sql.NullString
			idVisi, idMisi                                         sql.NullInt64
			tahunAwalPeriode, tahunAkhirPeriode, jenisPeriodeVal   sql.NullString
			indikatorId, indikatorText                             sql.NullString
			rumusPerhitungan, sumberData                           sql.NullString
			targetId, targetValue, targetSatuan, targetTahun       sql.NullString
		)

		err := rows.Scan(
			&pokinId,
			&namaPohon,
			&jenisPohon,
			&levelPohon,
			&is_active,
			&kodeOpd,
			&keterangan,
			&tahunPokin,
			&tujuanId,
			&tujuanPemda,
			&idVisi,
			&idMisi,
			&tahunAwalPeriode,
			&tahunAkhirPeriode,
			&jenisPeriodeVal,
			&indikatorId,
			&indikatorText,
			&rumusPerhitungan,
			&sumberData,
			&targetId,
			&targetValue,
			&targetSatuan,
			&targetTahun,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Inisialisasi atau ambil data pokin dari map
		pokin, exists := pokinMap[pokinId]
		if !exists {
			pokin = &domain.TujuanPemdaWithPokin{
				PokinId:     pokinId,
				NamaPohon:   namaPohon,
				JenisPohon:  jenisPohon,
				LevelPohon:  levelPohon,
				IsActive:    is_active,
				KodeOpd:     kodeOpd,
				Keterangan:  keterangan,
				TahunPokin:  tahunPokin,
				TujuanPemda: []domain.TujuanPemda{},
			}
			pokinMap[pokinId] = pokin
		}

		// Jika ada data tujuan pemda
		if tujuanId.Valid && tujuanPemda.Valid {
			var existingTujuanPemda *domain.TujuanPemda
			for i := range pokin.TujuanPemda {
				if pokin.TujuanPemda[i].Id == int(tujuanId.Int64) {
					existingTujuanPemda = &pokin.TujuanPemda[i]
					break
				}
			}

			if existingTujuanPemda == nil {
				newTujuanPemda := domain.TujuanPemda{
					Id:          int(tujuanId.Int64),
					TujuanPemda: tujuanPemda.String,
					TematikId:   pokinId,
					IsActive:    is_active,
					IdVisi:      int(idVisi.Int64),
					IdMisi:      int(idMisi.Int64),
					Indikator:   []domain.Indikator{},
				}
				pokin.TujuanPemda = append(pokin.TujuanPemda, newTujuanPemda)
				existingTujuanPemda = &pokin.TujuanPemda[len(pokin.TujuanPemda)-1]
			}

			// Proses indikator
			if indikatorId.Valid && indikatorText.Valid {
				if _, exists := indikatorMap[indikatorId.String]; !exists {
					rumusPerhitunganStr := ""
					if rumusPerhitungan.Valid {
						rumusPerhitunganStr = rumusPerhitungan.String
					}
					sumberDataStr := ""
					if sumberData.Valid {
						sumberDataStr = sumberData.String
					}

					indikator := &domain.Indikator{
						Id:               indikatorId.String,
						TujuanPemdaId:    int(tujuanId.Int64),
						Indikator:        indikatorText.String,
						RumusPerhitungan: sql.NullString{String: rumusPerhitunganStr, Valid: rumusPerhitunganStr != ""},
						SumberData:       sql.NullString{String: sumberDataStr, Valid: sumberDataStr != ""},
						Target:           []domain.Target{},
					}

					// Generate target untuk setiap tahun dalam periode
					tahunAwalInt, _ := strconv.Atoi(tahunAwal)
					tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)

					for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
						tahunStr := strconv.Itoa(tahun)
						indikator.Target = append(indikator.Target, domain.Target{
							Id:          "-",
							IndikatorId: indikatorId.String,
							Target:      "-",
							Satuan:      "-",
							Tahun:       tahunStr,
						})
					}

					indikatorMap[indikatorId.String] = indikator
					existingTujuanPemda.Indikator = append(existingTujuanPemda.Indikator, *indikator)
				}

				// Update target jika ada
				if targetId.Valid && targetValue.Valid && targetTahun.Valid {
					currentIndikator := indikatorMap[indikatorId.String]
					for i := range currentIndikator.Target {
						if currentIndikator.Target[i].Tahun == targetTahun.String {
							currentIndikator.Target[i] = domain.Target{
								Id:          targetId.String,
								IndikatorId: indikatorId.String,
								Target:      targetValue.String,
								Satuan:      targetSatuan.String,
								Tahun:       targetTahun.String,
							}
							break
						}
					}
				}
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Konversi map ke slice
	result := make([]domain.TujuanPemdaWithPokin, 0, len(pokinMap))
	for _, pokin := range pokinMap {
		result = append(result, *pokin)
	}

	// Sort hasil akhir berdasarkan pokin_id
	sort.Slice(result, func(i, j int) bool {
		return result[i].PokinId < result[j].PokinId
	})

	return result, nil
}

func (repository *TujuanPemdaRepositoryImpl) IsPokinIdExists(ctx context.Context, tx *sql.Tx, pokinId int) (bool, error) {
	query := "SELECT COUNT(*) FROM tb_tujuan_pemda WHERE tematik_id = ?"
	var count int
	err := tx.QueryRowContext(ctx, query, pokinId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *TujuanPemdaRepositoryImpl) GetAllIndikatorTujuanPemda(ctx context.Context, tx *sql.Tx) ([]domain.Indikator, error) {
	query := `
SELECT ind.id, ind.tujuan_pemda_id, ind.indikator, ind.sumber_data, ind.rumus_perhitungan,
       tg.id, tg.target, tg.satuan, tg.tahun
FROM tb_indikator ind
LEFT JOIN tb_target tg ON tg.indikator_id = ind.id
JOIN tb_tujuan_pemda tp ON ind.tujuan_pemda_id = tp.id
ORDER BY tg.tahun;
`
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}(rows)

	indikatorMap := make(map[string]*domain.Indikator)
	for rows.Next() {
		var (
			indId, indikatorStr                         string
			tujuanPemdaId                               int
			sumberData, rumusPerhitungan                sql.NullString
			targetId, targetStr, satuanStr, targetTahun sql.NullString
		)

		err := rows.Scan(
			&indId, &tujuanPemdaId, &indikatorStr, &sumberData, &rumusPerhitungan,
			&targetId, &targetStr, &satuanStr, &targetTahun,
		)
		if err != nil {
			return nil, err
		}

		// cek indikator ada di map ??
		indikator, exists := indikatorMap[indId]
		if !exists {
			indikator = &domain.Indikator{
				Id:               indId,
				TujuanPemdaId:    tujuanPemdaId,
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
				Id:     targetId.String,
				Target: targetStr.String,
				Satuan: satuanStr.String,
				Tahun:  targetTahun.String,
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

func (repository *TujuanPemdaRepositoryImpl) GetAllIndikatorTujuanPemdaByTahun(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.Indikator, error) {
	query := `
SELECT ind.id, ind.tujuan_pemda_id, ind.indikator, ind.sumber_data, ind.rumus_perhitungan,
       tg.id, tg.target, tg.satuan, tg.tahun
FROM tb_indikator ind
LEFT JOIN tb_target tg ON tg.indikator_id = ind.id AND tg.tahun = ?
JOIN tb_tujuan_pemda tp ON ind.tujuan_pemda_id = tp.id
	AND CAST(tp.tahun_awal_periode AS UNSIGNED) <= ?
	AND CAST(tp.tahun_akhir_periode AS UNSIGNED) >= ?
ORDER BY tg.tahun;
`
	rows, err := tx.QueryContext(ctx, query, tahun, tahun, tahun)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}(rows)

	indikatorMap := make(map[string]*domain.Indikator)
	for rows.Next() {
		var (
			indId, indikatorStr                         string
			tujuanPemdaId                               int
			sumberData, rumusPerhitungan                sql.NullString
			targetId, targetStr, satuanStr, targetTahun sql.NullString
		)

		err := rows.Scan(
			&indId, &tujuanPemdaId, &indikatorStr, &sumberData, &rumusPerhitungan,
			&targetId, &targetStr, &satuanStr, &targetTahun,
		)
		if err != nil {
			return nil, err
		}

		// cek indikator ada di map ??
		indikator, exists := indikatorMap[indId]
		if !exists {
			indikator = &domain.Indikator{
				Id:               indId,
				TujuanPemdaId:    tujuanPemdaId,
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
				Id:     targetId.String,
				Target: targetStr.String,
				Satuan: satuanStr.String,
				Tahun:  targetTahun.String,
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
