package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
	"fmt"
	"sort"
	"strconv"
)

type SasaranPemdaRepositoryImpl struct {
}

func NewSasaranPemdaRepositoryImpl() *SasaranPemdaRepositoryImpl {
	return &SasaranPemdaRepositoryImpl{}
}

func (repository *SasaranPemdaRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, sasaranPemda domain.SasaranPemda) (domain.SasaranPemda, error) {
	// Insert sasaran pemda
	query := "INSERT INTO tb_sasaran_pemda(id, tujuan_pemda_id, subtema_id, sasaran_pemda, periode_id, tahun_awal, tahun_akhir, jenis_periode) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, sasaranPemda.Id, sasaranPemda.TujuanPemdaId, sasaranPemda.SubtemaId, sasaranPemda.SasaranPemda, sasaranPemda.PeriodeId, sasaranPemda.TahunAwal, sasaranPemda.TahunAkhir, sasaranPemda.JenisPeriode)
	if err != nil {
		return sasaranPemda, err
	}

	// Insert indikator
	for _, indikator := range sasaranPemda.Indikator {
		scriptInsertIndikator := `
            INSERT INTO tb_indikator 
                (id, sasaran_pemda_id, indikator, rumus_perhitungan, sumber_data) 
            VALUES 
                (?, ?, ?, ?, ?)`

		_, err := tx.ExecContext(ctx, scriptInsertIndikator,
			indikator.Id,
			sasaranPemda.Id,
			indikator.Indikator,
			indikator.RumusPerhitungan,
			indikator.SumberData)
		if err != nil {
			return sasaranPemda, err
		}

		// Insert target untuk setiap indikator
		for _, target := range indikator.Target {
			// Skip jika target kosong
			if target.Target == "" && target.Satuan == "" {
				continue
			}

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
				return sasaranPemda, err
			}
		}
	}

	return sasaranPemda, nil
}

func (repository *SasaranPemdaRepositoryImpl) CreateIndikator(ctx context.Context, tx *sql.Tx, indikator domain.Indikator) (domain.Indikator, error) {
	query := "INSERT INTO tb_indikator(id, sasaran_pemda_id, indikator, rumus_perhitungan, sumber_data) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, indikator.Id, indikator.SasaranPemdaId, indikator.Indikator, indikator.RumusPerhitungan, indikator.SumberData)
	if err != nil {
		return indikator, err
	}
	return indikator, nil
}

func (repository *SasaranPemdaRepositoryImpl) CreateTarget(ctx context.Context, tx *sql.Tx, target domain.Target) error {
	query := "INSERT INTO tb_target(id, indikator_id, target, satuan, tahun) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, query, target.Id, target.IndikatorId, target.Target, target.Satuan, target.Tahun)
	return err
}

func (repository *SasaranPemdaRepositoryImpl) Update(ctx context.Context, tx *sql.Tx, sasaranPemda domain.SasaranPemda) (domain.SasaranPemda, error) {
	// Update sasaran pemda
	query := "UPDATE tb_sasaran_pemda SET sasaran_pemda = ?, tujuan_pemda_id = ?, periode_id = ?, tahun_awal = ?, tahun_akhir = ?, jenis_periode = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, query, sasaranPemda.SasaranPemda, sasaranPemda.TujuanPemdaId, sasaranPemda.PeriodeId, sasaranPemda.TahunAwal, sasaranPemda.TahunAkhir, sasaranPemda.JenisPeriode, sasaranPemda.Id)
	if err != nil {
		return sasaranPemda, err
	}

	// Hapus semua indikator lama beserta targetnya
	scriptDeleteOldIndicators := "DELETE FROM tb_indikator WHERE sasaran_pemda_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeleteOldIndicators, sasaranPemda.Id)
	if err != nil {
		return sasaranPemda, err
	}

	// Insert indikator baru
	for _, indikator := range sasaranPemda.Indikator {
		scriptInsertIndikator := `
            INSERT INTO tb_indikator 
                (id, sasaran_pemda_id, indikator, rumus_perhitungan, sumber_data) 
            VALUES 
                (?, ?, ?, ?, ?)`

		_, err := tx.ExecContext(ctx, scriptInsertIndikator,
			indikator.Id,
			sasaranPemda.Id,
			indikator.Indikator,
			indikator.RumusPerhitungan,
			indikator.SumberData)
		if err != nil {
			return sasaranPemda, err
		}

		// Hapus semua target lama untuk indikator ini
		scriptDeleteOldTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
		_, err = tx.ExecContext(ctx, scriptDeleteOldTargets, indikator.Id)
		if err != nil {
			return sasaranPemda, err
		}

		// Insert target baru
		for _, target := range indikator.Target {
			// Skip jika target kosong
			if target.Target == "" && target.Satuan == "" {
				continue
			}

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
				return sasaranPemda, err
			}
		}
	}

	return sasaranPemda, nil
}

func (repository *SasaranPemdaRepositoryImpl) Delete(ctx context.Context, tx *sql.Tx, sasaranPemdaId int) error {
	queryDeleteTarget := `
	DELETE t FROM tb_target t
	INNER JOIN tb_indikator i ON t.indikator_id = i.id
	WHERE i.sasaran_pemda_id = ?`
	_, err := tx.ExecContext(ctx, queryDeleteTarget, sasaranPemdaId)
	if err != nil {
		return err
	}

	// 2. Hapus indikator yang terkait dengan tujuan pemda
	queryDeleteIndikator := "DELETE FROM tb_indikator WHERE sasaran_pemda_id = ?"
	_, err = tx.ExecContext(ctx, queryDeleteIndikator, sasaranPemdaId)
	if err != nil {
		return err
	}
	query := "DELETE FROM tb_sasaran_pemda WHERE id = ?"
	_, err = tx.ExecContext(ctx, query, sasaranPemdaId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *SasaranPemdaRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, sasaranPemdaId int) (domain.SasaranPemda, error) {
	query := `
        SELECT DISTINCT
            tp.id,
            tp.tujuan_pemda_id,
            tp.subtema_id,
            tp.sasaran_pemda,
            tp.periode_id,
            COALESCE(p.tahun_awal, '') as tahun_awal,
            COALESCE(p.tahun_akhir, '') as tahun_akhir,
			COALESCE(p.jenis_periode, '') as jenis_periode,
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
            tb_sasaran_pemda tp
            LEFT JOIN tb_periode p ON tp.periode_id = p.id
            LEFT JOIN tb_pohon_kinerja pk ON tp.subtema_id = pk.id
            LEFT JOIN tb_indikator i ON tp.id = i.sasaran_pemda_id
            LEFT JOIN tb_target t ON t.indikator_id = i.id
        WHERE tp.id = ?
        ORDER BY 
            tp.id, i.id, CAST(t.tahun AS SIGNED)`

	rows, err := tx.QueryContext(ctx, query, sasaranPemdaId)
	if err != nil {
		return domain.SasaranPemda{}, fmt.Errorf("error querying sasaran pemda: %v", err)
	}
	defer rows.Close()

	var result domain.SasaranPemda
	var firstRow = true
	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			id, subtemaId, periodeId, tujuanPemdaId                           int
			sasaranPemdaText, tahunAwal, tahunAkhir, jenisPeriode, jenisPohon string // Tambahkan jenisPohon
			indikatorId, indikatorText                                        sql.NullString
			rumusPerhitunganNull, sumberDataNull                              sql.NullString
			targetId, targetValue, targetSatuan, targetTahun                  sql.NullString
		)

		err := rows.Scan(
			&id,
			&tujuanPemdaId,
			&subtemaId,
			&sasaranPemdaText,
			&periodeId,
			&tahunAwal,
			&tahunAkhir,
			&jenisPeriode,
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
			return domain.SasaranPemda{}, fmt.Errorf("error scanning row: %v", err)
		}

		if firstRow {
			result = domain.SasaranPemda{
				Id:            id,
				TujuanPemdaId: tujuanPemdaId,
				SubtemaId:     subtemaId,
				SasaranPemda:  sasaranPemdaText,
				PeriodeId:     periodeId,
				JenisPohon:    jenisPohon,
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
				// Konversi NullString ke string biasa
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
					SasaranPemdaId:   id,
					Indikator:        indikatorText.String,
					RumusPerhitungan: sql.NullString{String: rumusPerhitungan, Valid: rumusPerhitungan != ""},
					SumberData:       sql.NullString{String: sumberData, Valid: sumberData != ""},
					Target:           []domain.Target{},
				}

				// Generate target untuk setiap tahun dalam periode
				if periodeId != 0 && tahunAwal != "" && tahunAkhir != "" {
					tahunAwalInt, _ := strconv.Atoi(tahunAwal)
					tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)

					for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
						tahunStr := strconv.Itoa(tahun)
						indikator.Target = append(indikator.Target, domain.Target{
							Id:          "-",
							IndikatorId: indikatorId.String,
							Target:      "",
							Satuan:      "",
							Tahun:       tahunStr,
						})
					}
				}

				result.Indikator = append(result.Indikator, indikator)
				indikatorMap[indikatorId.String] = &result.Indikator[len(result.Indikator)-1]
				currentIndikator = &result.Indikator[len(result.Indikator)-1]
			}

			// Update target yang ada dengan data sebenarnya
			if targetId.Valid && targetValue.Valid && targetTahun.Valid {
				tahunInt, _ := strconv.Atoi(targetTahun.String)
				tahunAwalInt, _ := strconv.Atoi(tahunAwal)
				if tahunInt >= tahunAwalInt {
					idx := tahunInt - tahunAwalInt
					if idx >= 0 && idx < len(currentIndikator.Target) {
						currentIndikator.Target[idx] = domain.Target{
							Id:          targetId.String,
							IndikatorId: indikatorId.String,
							Target:      targetValue.String,
							Satuan:      targetSatuan.String,
							Tahun:       targetTahun.String,
						}
					}
				}
			}
		}
	}

	if err = rows.Err(); err != nil {
		return domain.SasaranPemda{}, fmt.Errorf("error iterating rows: %v", err)
	}

	if result.Id == 0 {
		return domain.SasaranPemda{}, fmt.Errorf("sasaran pemda dengan id %d tidak ditemukan", sasaranPemdaId)
	}

	// Sort indikator berdasarkan ID
	sort.Slice(result.Indikator, func(i, j int) bool {
		return result.Indikator[i].Id < result.Indikator[j].Id
	})

	return result, nil
}

func (repository *SasaranPemdaRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.SasaranPemda, error) {
	query := `
		SELECT id, sasaran_pemda, tahun_awal, tahun_akhir
		FROM tb_sasaran_pemda
		WHERE CAST(tahun_awal AS UNSIGNED) <= ?
		  AND CAST(tahun_akhir AS UNSIGNED) >= ?
	`

	rows, err := tx.QueryContext(ctx, query, tahun, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.SasaranPemda

	for rows.Next() {
		var sp domain.SasaranPemda

		err := rows.Scan(&sp.Id, &sp.SasaranPemda, &sp.TahunAwal, &sp.TahunAkhir)
		if err != nil {
			return nil, err
		}

		results = append(results, sp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *SasaranPemdaRepositoryImpl) GetIndikatorSasaranByTahun(ctx context.Context, tx *sql.Tx, sasaranPemdaId int, tahun string) ([]domain.Indikator, error) {
	query := `
		SELECT ind.id, ind.sasaran_pemda_id, ind.indikator, ind.sumber_data, ind.rumus_perhitungan,
        tg.id, tg.target, tg.satuan, tg.tahun
		FROM tb_indikator ind
        LEFT JOIN tb_target tg ON tg.indikator_id = ind.id AND tg.tahun = ?
		WHERE ind.sasaran_pemda_id = ?
	`

	rows, err := tx.QueryContext(ctx, query, tahun, sasaranPemdaId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			indId, indikatorStr                         string
			sasaranPemdaId                              int
			sumberData, rumusPerhitungan                sql.NullString
			targetId, targetStr, satuanStr, targetTahun sql.NullString
		)

		err := rows.Scan(
			&indId, &sasaranPemdaId, &indikatorStr, &sumberData, &rumusPerhitungan,
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
				SasaranPemdaId:   sasaranPemdaId,
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

	results := make([]domain.Indikator, 0, len(indikatorMap))
	for _, indikator := range indikatorMap {
		results = append(results, *indikator)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *SasaranPemdaRepositoryImpl) DeleteIndikator(ctx context.Context, tx *sql.Tx, sasaranPemdaId int) error {
	query := "DELETE FROM tb_indikator WHERE sasaran_pemda_id = ?"
	_, err := tx.ExecContext(ctx, query, sasaranPemdaId)
	return err
}

func (repository *SasaranPemdaRepositoryImpl) IsIdExists(ctx context.Context, tx *sql.Tx, id int) bool {
	query := "SELECT COUNT(*) FROM tb_sasaran_pemda WHERE id = ?"
	var count int
	err := tx.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return true
	}
	return count > 0
}

func (repository *SasaranPemdaRepositoryImpl) UpdatePeriode(ctx context.Context, tx *sql.Tx, sasaranPemda domain.SasaranPemda) (domain.SasaranPemda, error) {
	// Update hanya periode_id
	query := "UPDATE tb_sasaran_pemda SET periode_id = ? WHERE id = ?"
	result, err := tx.ExecContext(ctx, query, sasaranPemda.PeriodeId, sasaranPemda.Id)
	if err != nil {
		return domain.SasaranPemda{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.SasaranPemda{}, err
	}

	if rowsAffected == 0 {
		return domain.SasaranPemda{}, fmt.Errorf("periode dengan id %d sudah digunakan", sasaranPemda.PeriodeId)
	}

	// Ambil data terbaru setelah update
	query = `
        SELECT 
            tp.id,
            tp.sasaran_pemda_id,
            tp.periode_id,
            COALESCE(p.tahun_awal, 'Pilih periode') as tahun_awal,
            COALESCE(p.tahun_akhir, 'Pilih periode') as tahun_akhir
        FROM 
            tb_sasaran_pemda tp
            LEFT JOIN tb_periode p ON tp.periode_id = p.id
        WHERE tp.id = ?`

	var updatedSasaranPemda domain.SasaranPemda
	err = tx.QueryRowContext(ctx, query, sasaranPemda.Id).Scan(
		&updatedSasaranPemda.Id,
		&updatedSasaranPemda.SubtemaId,
		&updatedSasaranPemda.SasaranPemda,
		&updatedSasaranPemda.PeriodeId,
		&updatedSasaranPemda.Periode.TahunAwal,
		&updatedSasaranPemda.Periode.TahunAkhir,
	)
	if err != nil {
		return domain.SasaranPemda{}, fmt.Errorf("gagal mengambil data setelah update: %v", err)
	}

	return updatedSasaranPemda, nil
}

func (repository *SasaranPemdaRepositoryImpl) FindAllWithPokin(ctx context.Context, tx *sql.Tx, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.PohonKinerjaWithSasaran, error) {
	query := `
	WITH RECURSIVE pohon_hierarchy AS (
        SELECT 
            pk.id,
            pk.nama_pohon,
            pk.parent,
            pk.level_pohon,
            pk.jenis_pohon,
            pk.keterangan,
            pk.is_active,
            pk.tahun,
            CAST(pk.id AS CHAR(50)) as path,
            pk.id as root_id,
            pk.nama_pohon as root_nama
        FROM tb_pohon_kinerja pk
        WHERE pk.level_pohon = 0
        AND CAST(pk.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)

        UNION ALL

        SELECT 
            c.id,
            c.nama_pohon,
            c.parent,
            c.level_pohon,
            c.jenis_pohon,
            c.keterangan,
            c.is_active,
            c.tahun,
            CONCAT(ph.path, ',', c.id),
            ph.root_id,
            ph.root_nama
        FROM tb_pohon_kinerja c
        JOIN pohon_hierarchy ph ON c.parent = ph.id
        WHERE CAST(c.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)
    )
    SELECT DISTINCT
        pk.id as subtematik_id,
        pk.nama_pohon as nama_subtematik,
        pk.jenis_pohon,
        pk.level_pohon,
        pk.keterangan,
        -- Logika untuk menentukan status aktif final
        CASE 
            WHEN sp.id IS NULL THEN 
                -- Jika tidak terhubung dengan sasaran pemda, cek is_active subtematik saja
                pk.is_active
            ELSE 
                -- Jika terhubung dengan sasaran pemda, cek is_active subtematik dan tematik
                CASE 
                    WHEN pk.is_active = true AND tematik.is_active = true THEN true
                    ELSE false
                END
        END as is_active,
        pk.tahun as pohon_tahun,
        pk.root_id as tematik_id,
        pk.root_nama as nama_tematik,
        sp.id as id_sasaran_pemda,
        sp.sasaran_pemda,
        sp.tahun_awal,
        sp.tahun_akhir,
        sp.jenis_periode,
        i.id as indikator_id,
        i.indikator,
        i.rumus_perhitungan,
        i.sumber_data,
        t.id as target_id,
        t.target,
        t.satuan,
        t.tahun as target_tahun
    FROM pohon_hierarchy pk
    LEFT JOIN tb_sasaran_pemda sp ON pk.id = sp.subtema_id
        AND sp.tahun_awal = ? 
        AND sp.tahun_akhir = ?
        AND sp.jenis_periode = ?
    LEFT JOIN tb_tujuan_pemda tp ON sp.tujuan_pemda_id = tp.id
    LEFT JOIN tb_pohon_kinerja tematik ON tp.tematik_id = tematik.id
    LEFT JOIN tb_indikator i ON sp.id = i.sasaran_pemda_id
    LEFT JOIN tb_target t ON i.id = t.indikator_id
        AND CAST(t.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)
    WHERE pk.level_pohon BETWEEN 1 AND 3
    ORDER BY pk.root_id, pk.id, sp.id, i.id, CAST(t.tahun AS SIGNED)`

	rows, err := tx.QueryContext(ctx, query,
		tahunAwal, tahunAkhir,
		tahunAwal, tahunAkhir,
		tahunAwal, tahunAkhir, jenisPeriode,
		tahunAwal, tahunAkhir)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tematikMap := make(map[int]*domain.PohonKinerjaWithSasaran)

	for rows.Next() {
		var (
			subtematikId                           int
			namaSubtematik, jenisPohon, keterangan string
			levelPohon                             int
			is_active                              bool
			pohonTahun                             string
			tematikId                              int
			namaTematik                            string
			idSasaranPemda                         sql.NullInt64
			sasaranPemda                           sql.NullString
			tahunAwalSasaran                       sql.NullString
			tahunAkhirSasaran                      sql.NullString
			jenisPeriodeSasaran                    sql.NullString
			indikatorId, indikator                 sql.NullString
			rumusPerhitungan, sumberData           sql.NullString
			targetId, target, satuan, targetTahun  sql.NullString
		)
		err := rows.Scan(
			&subtematikId,
			&namaSubtematik,
			&jenisPohon,
			&levelPohon,
			&keterangan,
			&is_active,
			&pohonTahun,
			&tematikId,
			&namaTematik,
			&idSasaranPemda,
			&sasaranPemda,
			&tahunAwalSasaran,
			&tahunAkhirSasaran,
			&jenisPeriodeSasaran,
			&indikatorId,
			&indikator,
			&rumusPerhitungan,
			&sumberData,
			&targetId,
			&target,
			&satuan,
			&targetTahun,
		)
		if err != nil {
			return nil, err
		}

		// Inisialisasi tematik jika belum ada
		tematik, exists := tematikMap[tematikId]
		if !exists {
			tematik = &domain.PohonKinerjaWithSasaran{
				TematikId:   tematikId,
				NamaTematik: namaTematik,
				Subtematik:  []domain.SubtematikWithSasaran{},
			}
			tematikMap[tematikId] = tematik
		}

		// Cari subtematik yang sudah ada
		var foundSubtematik *domain.SubtematikWithSasaran
		for i := range tematik.Subtematik {
			if tematik.Subtematik[i].SubtematikId == subtematikId {
				foundSubtematik = &tematik.Subtematik[i]
				break
			}
		}

		// Buat subtematik baru jika belum ada
		if foundSubtematik == nil {
			newSubtematik := domain.SubtematikWithSasaran{
				SubtematikId:     subtematikId,
				NamaSubtematik:   namaSubtematik,
				JenisPohon:       jenisPohon,
				LevelPohon:       levelPohon,
				IsActive:         is_active,
				SasaranPemdaList: []domain.SasaranPemdaDetail{},
			}
			tematik.Subtematik = append(tematik.Subtematik, newSubtematik)
			foundSubtematik = &tematik.Subtematik[len(tematik.Subtematik)-1]
		}

		// Proses sasaran pemda jika ada
		if idSasaranPemda.Valid && sasaranPemda.Valid {
			var foundSasaran *domain.SasaranPemdaDetail
			for i := range foundSubtematik.SasaranPemdaList {
				if foundSubtematik.SasaranPemdaList[i].Id == int(idSasaranPemda.Int64) {
					foundSasaran = &foundSubtematik.SasaranPemdaList[i]
					break
				}
			}

			if foundSasaran == nil {
				newSasaran := domain.SasaranPemdaDetail{
					Id:           int(idSasaranPemda.Int64),
					SasaranPemda: sasaranPemda.String,
					Periode: domain.Periode{
						TahunAwal:    tahunAwal,
						TahunAkhir:   tahunAkhir,
						JenisPeriode: jenisPeriode,
					},
					Indikator: []domain.IndikatorDetail{},
				}
				foundSubtematik.SasaranPemdaList = append(foundSubtematik.SasaranPemdaList, newSasaran)
				foundSasaran = &foundSubtematik.SasaranPemdaList[len(foundSubtematik.SasaranPemdaList)-1]
			}

			// Proses indikator jika ada
			if indikatorId.Valid {
				var foundIndikator *domain.IndikatorDetail
				for i := range foundSasaran.Indikator {
					if foundSasaran.Indikator[i].Id == indikatorId.String {
						foundIndikator = &foundSasaran.Indikator[i]
						break
					}
				}

				if foundIndikator == nil {
					newIndikator := domain.IndikatorDetail{
						Id:               indikatorId.String,
						Indikator:        indikator.String,
						RumusPerhitungan: rumusPerhitungan,
						SumberData:       sumberData,
						Target:           []domain.TargetDetail{},
					}

					// Buat target kosong untuk setiap tahun dalam range
					tahunAwalInt, _ := strconv.Atoi(tahunAwal)
					tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)
					for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
						tahunStr := strconv.Itoa(tahun)
						newIndikator.Target = append(newIndikator.Target, domain.TargetDetail{
							Id:     "-",
							Target: "",
							Satuan: "",
							Tahun:  tahunStr,
						})
					}

					foundSasaran.Indikator = append(foundSasaran.Indikator, newIndikator)
					foundIndikator = &foundSasaran.Indikator[len(foundSasaran.Indikator)-1]
				}

				// Update target jika ada
				if targetId.Valid && targetTahun.Valid {
					tahunTarget, _ := strconv.Atoi(targetTahun.String)
					tahunAwalInt, _ := strconv.Atoi(tahunAwal)
					tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)

					if tahunTarget >= tahunAwalInt && tahunTarget <= tahunAkhirInt {
						for i := range foundIndikator.Target {
							if foundIndikator.Target[i].Tahun == targetTahun.String {
								foundIndikator.Target[i] = domain.TargetDetail{
									Id:     targetId.String,
									Target: target.String,
									Satuan: satuan.String,
									Tahun:  targetTahun.String,
								}
								break
							}
						}
					}
				}
				sort.Slice(foundIndikator.Target, func(i, j int) bool {
					tahunI, _ := strconv.Atoi(foundIndikator.Target[i].Tahun)
					tahunJ, _ := strconv.Atoi(foundIndikator.Target[j].Tahun)
					return tahunI < tahunJ
				})
			}

		}
	}

	// Convert map to slice
	var result []domain.PohonKinerjaWithSasaran
	for _, tematik := range tematikMap {
		result = append(result, *tematik)
	}

	return result, nil
}

func (repository *SasaranPemdaRepositoryImpl) IsSubtemaIdExists(ctx context.Context, tx *sql.Tx, subtemaId int) bool {
	query := "SELECT COUNT(*) FROM tb_sasaran_pemda WHERE subtema_id = ?"
	var count int
	err := tx.QueryRowContext(ctx, query, subtemaId).Scan(&count)
	if err != nil {
		return false // Ubah return value jika error menjadi false
	}
	return count > 0
}
