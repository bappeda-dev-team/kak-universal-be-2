package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"fmt"
	"sort"
	"strconv"
)

type IkuRepositoryImpl struct {
}

func NewIkuRepositoryImpl() *IkuRepositoryImpl {
	return &IkuRepositoryImpl{}
}

// iku pemda
func (repository *IkuRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.Indikator, error) {
	query := `
		WITH indikator_tujuan AS (
			SELECT 
				i.id as indikator_id,
				i.indikator,
				i.rumus_perhitungan,
				i.sumber_data,
				i.created_at as indikator_created_at,
				t.id as target_id,
				t.target,
				t.satuan,
				t.tahun as target_tahun,
				'Tujuan Pemda' as sumber,
				tp.id as parent_id,
				tp.tujuan_pemda as parent_name,
				tp.tahun_awal_periode,
				tp.tahun_akhir_periode,
				tp.jenis_periode,
				CASE 
					WHEN pk_tematik.id IS NULL THEN false  -- Jika tematik_id tidak ditemukan
					ELSE COALESCE(pk_tematik.is_active, false)  -- Jika ditemukan, gunakan status is_active
				END as is_active,
				CASE 
					WHEN pk_tematik.id IS NULL THEN false  -- Jika tematik_id tidak ditemukan
					ELSE true  -- Jika ditemukan
				END as is_exists
			FROM tb_indikator i
			INNER JOIN tb_tujuan_pemda tp ON i.tujuan_pemda_id = tp.id
			LEFT JOIN tb_target t ON t.indikator_id = i.id
			LEFT JOIN tb_pohon_kinerja pk_tematik ON tp.tematik_id = pk_tematik.id
			WHERE tp.tahun_awal_periode = ? 
			AND tp.tahun_akhir_periode = ?
			AND tp.jenis_periode = ?
		),
		indikator_sasaran AS (
			SELECT 
				i.id as indikator_id,
				i.indikator,
				i.rumus_perhitungan,
				i.sumber_data,
				i.created_at as indikator_created_at,
				t.id as target_id,
				t.target,
				t.satuan,
				t.tahun as target_tahun,
				'Sasaran Pemda' as sumber,
				sp.id as parent_id,
				sp.sasaran_pemda as parent_name,
				sp.tahun_awal,
				sp.tahun_akhir,
				sp.jenis_periode,
				CASE 
					WHEN pk_tematik.id IS NULL THEN false  -- Jika tematik_id tidak ditemukan
					WHEN pk_subtematik.id IS NULL THEN false  -- Jika subtematik_id tidak ditemukan
					ELSE COALESCE(pk_subtematik.is_active, false)  -- Jika keduanya ditemukan, gunakan status is_active subtematik
				END as is_active,
				CASE 
					WHEN pk_tematik.id IS NULL THEN false  -- Jika tematik_id tidak ditemukan
					WHEN pk_subtematik.id IS NULL THEN false  -- Jika subtematik_id tidak ditemukan
					ELSE true  -- Jika keduanya ditemukan
				END as is_exists
			FROM tb_indikator i
			INNER JOIN tb_sasaran_pemda sp ON i.sasaran_pemda_id = sp.id
			LEFT JOIN tb_target t ON t.indikator_id = i.id
			LEFT JOIN tb_tujuan_pemda tp ON sp.tujuan_pemda_id = tp.id
			LEFT JOIN tb_pohon_kinerja pk_tematik ON tp.tematik_id = pk_tematik.id
			LEFT JOIN tb_pohon_kinerja pk_subtematik ON sp.subtema_id = pk_subtematik.id
			WHERE sp.tahun_awal = ? 
			AND sp.tahun_akhir = ?
			AND sp.jenis_periode = ?
		)
		SELECT * FROM (
			SELECT * FROM indikator_tujuan WHERE is_exists = true
			UNION ALL
			SELECT * FROM indikator_sasaran WHERE is_exists = true
		) combined
		WHERE indikator IS NOT NULL
		ORDER BY indikator_created_at ASC`

	rows, err := tx.QueryContext(ctx, query,
		tahunAwal, tahunAkhir, jenisPeriode,
		tahunAwal, tahunAkhir, jenisPeriode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			indikatorId        sql.NullString
			indikator          sql.NullString
			rumusPerhitungan   sql.NullString
			sumberData         sql.NullString
			indikatorCreatedAt sql.NullTime
			targetId           sql.NullString
			target             sql.NullString
			satuan             sql.NullString
			targetTahun        sql.NullString
			sumber             string
			parentId           sql.NullInt64
			parentName         sql.NullString
			tahunAwal          string
			tahunAkhir         string
			jenisPeriodeData   string
			isActive           bool
			isExists           bool
		)

		err := rows.Scan(
			&indikatorId,
			&indikator,
			&rumusPerhitungan,
			&sumberData,
			&indikatorCreatedAt,
			&targetId,
			&target,
			&satuan,
			&targetTahun,
			&sumber,
			&parentId,
			&parentName,
			&tahunAwal,
			&tahunAkhir,
			&jenisPeriodeData,
			&isActive,
			&isExists,
		)
		if err != nil {
			return nil, err
		}

		// Skip jika indikator tidak valid atau tidak exists
		if !indikator.Valid || !indikatorId.Valid || !isExists {
			continue
		}

		item, exists := indikatorMap[indikatorId.String]
		if !exists {
			// Inisialisasi target kosong untuk semua tahun dalam periode
			tahunAwalInt, _ := strconv.Atoi(tahunAwal)
			tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)

			var targets []domain.Target
			for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
				tahunStr := strconv.Itoa(tahun)
				targets = append(targets, domain.Target{
					Id:          "-",
					IndikatorId: indikatorId.String,
					Target:      "",
					Satuan:      "",
					Tahun:       tahunStr,
				})
			}

			item = &domain.Indikator{
				Id:               indikatorId.String,
				Indikator:        indikator.String,
				RumusPerhitungan: rumusPerhitungan,
				SumberData:       sumberData,
				CreatedAt:        indikatorCreatedAt.Time,
				Sumber:           sumber,
				ParentId:         int(parentId.Int64),
				ParentName:       parentName.String,
				Target:           targets,
				IsActive:         isActive,
			}
			indikatorMap[indikatorId.String] = item
		}

		// Update target yang memiliki data
		if targetId.Valid && targetTahun.Valid {
			tahunInt, _ := strconv.Atoi(targetTahun.String)
			tahunAwalInt, _ := strconv.Atoi(tahunAwal)
			if tahunInt >= tahunAwalInt {
				idx := tahunInt - tahunAwalInt
				if idx >= 0 && idx < len(item.Target) {
					item.Target[idx] = domain.Target{
						Id:          targetId.String,
						IndikatorId: indikatorId.String,
						Target:      target.String,
						Satuan:      satuan.String,
						Tahun:       targetTahun.String,
					}
				}
			}
		}
	}

	result := make([]domain.Indikator, 0, len(indikatorMap))
	for _, item := range indikatorMap {
		result = append(result, *item)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].Indikator < result[j].Indikator
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

func (repository *IkuRepositoryImpl) FindAllIkuOpd(ctx context.Context, tx *sql.Tx, kodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.Indikator, error) {
	// Query untuk mengambil indikator dari tujuan OPD
	scriptTujuan := `
        SELECT 
            'Tujuan OPD' as jenis,
            t.id as parent_id,
            t.tujuan as nama_parent,
            i.id as indikator_id,
            i.indikator,
           	i.rumus_perhitungan,
          	i.sumber_data,
            tg.id as target_id,
            tg.target,
            tg.satuan,
            tg.tahun
        FROM tb_tujuan_opd t
        LEFT JOIN tb_indikator i ON t.id = i.tujuan_opd_id
        LEFT JOIN tb_target tg ON i.id = tg.indikator_id
        WHERE t.kode_opd = ?
        AND t.tahun_awal = ?
        AND t.tahun_akhir = ?
        AND t.jenis_periode = ?
        AND (tg.tahun IS NULL OR (CAST(tg.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)))
    `

	// Query untuk mengambil indikator dari sasaran OPD
	scriptSasaran := `
		SELECT 
			'Sasaran OPD' as jenis,
			so.id as parent_id,
			so.nama_sasaran_opd as nama_parent,
			i.id as indikator_id,
			i.indikator,
			COALESCE(i.rumus_perhitungan, '') as rumus_perhitungan,  -- Ubah dari m.formula
			COALESCE(i.sumber_data, '') as sumber_data,              -- Ubah dari m.sumber_data
			tg.id as target_id,
			tg.target,
			tg.satuan,
			tg.tahun
			FROM tb_sasaran_opd so
			INNER JOIN tb_pohon_kinerja pk ON so.pokin_id = pk.id
			LEFT JOIN tb_indikator i ON so.id = i.sasaran_opd_id
			LEFT JOIN tb_target tg ON i.id = tg.indikator_id
			WHERE pk.kode_opd = ?
			AND so.tahun_awal = ?
			AND so.tahun_akhir = ?
			AND so.jenis_periode = ?
			AND (tg.tahun IS NULL OR (CAST(tg.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)))
		`

	// Map untuk menyimpan hasil
	ikuMap := make(map[string]*domain.Indikator)

	// Fungsi helper untuk membuat target kosong
	createEmptyTargets := func(indikatorId string, tahunAwal, tahunAkhir string) []domain.Target {
		tahunAwalInt, _ := strconv.Atoi(tahunAwal)
		tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)
		var targets []domain.Target

		for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
			tahunStr := strconv.Itoa(tahun)
			targets = append(targets, domain.Target{
				Id:          "-",
				IndikatorId: indikatorId,
				Target:      "",
				Satuan:      "",
				Tahun:       tahunStr,
			})
		}
		return targets
	}

	// Fungsi helper untuk memproses rows
	processRows := func(rows *sql.Rows) error {
		for rows.Next() {
			var (
				jenis                        string
				parentId, namaParent         sql.NullString
				indikatorId, namaIndikator   sql.NullString
				rumusPerhitungan, sumberData sql.NullString
				targetId                     sql.NullString
				target                       sql.NullString
				satuan                       sql.NullString
				tahun                        sql.NullString
			)

			err := rows.Scan(
				&jenis,
				&parentId,
				&namaParent,
				&indikatorId,
				&namaIndikator,
				&rumusPerhitungan,
				&sumberData,
				&targetId,
				&target,
				&satuan,
				&tahun,
			)
			if err != nil {
				return err
			}

			// Skip jika indikatorId NULL
			if !indikatorId.Valid {
				continue
			}

			// Buat key untuk map
			key := fmt.Sprintf("%s-%s-%s", jenis,
				helper.GetNullStringValue(parentId),
				indikatorId.String)

			// Cek apakah indikator sudah ada di map
			iku, exists := ikuMap[key]
			if !exists {
				// Buat array target kosong untuk semua tahun
				emptyTargets := createEmptyTargets(indikatorId.String, tahunAwal, tahunAkhir)

				iku = &domain.Indikator{
					Id:               indikatorId.String,
					AsalIku:          jenis,
					ParentOpdId:      helper.GetNullStringValue(parentId),
					ParentName:       helper.GetNullStringValue(namaParent),
					Indikator:        helper.GetNullStringValue(namaIndikator),
					RumusPerhitungan: rumusPerhitungan,
					SumberData:       sumberData,
					TahunAwal:        tahunAwal,
					TahunAkhir:       tahunAkhir,
					JenisPeriode:     jenisPeriode,
					Target:           emptyTargets,
				}
				ikuMap[key] = iku
			}

			// Update target jika ada
			if targetId.Valid && tahun.Valid && target.Valid {
				tahunInt, _ := strconv.Atoi(tahun.String)
				tahunAwalInt, _ := strconv.Atoi(tahunAwal)
				idx := tahunInt - tahunAwalInt

				if idx >= 0 && idx < len(iku.Target) {
					iku.Target[idx] = domain.Target{
						Id:          targetId.String,
						IndikatorId: indikatorId.String,
						Target:      target.String,
						Satuan:      helper.GetNullStringValue(satuan),
						Tahun:       tahun.String,
					}
				}
			}
		}
		return nil
	}

	// Proses rows tujuan dan sasaran
	rowsTujuan, err := tx.QueryContext(ctx, scriptTujuan,
		kodeOpd, tahunAwal, tahunAkhir, jenisPeriode, tahunAwal, tahunAkhir)
	if err != nil {
		return nil, err
	}
	defer rowsTujuan.Close()
	if err := processRows(rowsTujuan); err != nil {
		return nil, err
	}

	rowsSasaran, err := tx.QueryContext(ctx, scriptSasaran,
		kodeOpd, tahunAwal, tahunAkhir, jenisPeriode, tahunAwal, tahunAkhir)
	if err != nil {
		return nil, err
	}
	defer rowsSasaran.Close()
	if err := processRows(rowsSasaran); err != nil {
		return nil, err
	}

	// Convert map ke slice
	var result []domain.Indikator
	for _, iku := range ikuMap {
		// Sort target berdasarkan tahun
		sort.Slice(iku.Target, func(i, j int) bool {
			tahunI, _ := strconv.Atoi(iku.Target[i].Tahun)
			tahunJ, _ := strconv.Atoi(iku.Target[j].Tahun)
			return tahunI < tahunJ
		})
		result = append(result, *iku)
	}

	// Sort hasil berdasarkan jenis (Tujuan OPD dulu) dan indikator ASC
	sort.Slice(result, func(i, j int) bool {
		// Jika jenis berbeda, Tujuan OPD harus di depan
		if result[i].AsalIku != result[j].AsalIku {
			return result[i].AsalIku == "Tujuan OPD"
		}
		// Jika jenis sama, urutkan berdasarkan indikator ASC
		return result[i].Indikator < result[j].Indikator
	})

	return result, nil
}
