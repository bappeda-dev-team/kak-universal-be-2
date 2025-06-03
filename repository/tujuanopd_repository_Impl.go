package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/model/domain"
	"fmt"
	"sort"
	"strconv"
)

type TujuanOpdRepositoryImpl struct {
}

func NewTujuanOpdRepositoryImpl() *TujuanOpdRepositoryImpl {
	return &TujuanOpdRepositoryImpl{}
}

func (repository *TujuanOpdRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, tujuanOpd domain.TujuanOpd) (domain.TujuanOpd, error) {
	script := "INSERT INTO tb_tujuan_opd (kode_opd, kode_bidang_urusan, tujuan, periode_id, tahun_awal, tahun_akhir, jenis_periode) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := tx.ExecContext(ctx, script, tujuanOpd.KodeOpd, tujuanOpd.KodeBidangUrusan, tujuanOpd.Tujuan, tujuanOpd.PeriodeId.Id, tujuanOpd.TahunAwal, tujuanOpd.TahunAkhir, tujuanOpd.JenisPeriode)
	if err != nil {
		return domain.TujuanOpd{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.TujuanOpd{}, err
	}

	tujuanOpd.Id = int(id)

	for _, indikator := range tujuanOpd.Indikator {
		scriptIndikator := "INSERT INTO tb_indikator (id, tujuan_opd_id, indikator, rumus_perhitungan, sumber_data) VALUES (?, ?, ?, ?, ?)"
		_, err := tx.ExecContext(ctx, scriptIndikator, indikator.Id, id, indikator.Indikator, indikator.RumusPerhitungan, indikator.SumberData)
		if err != nil {
			return domain.TujuanOpd{}, err
		}

		for _, target := range indikator.Target {
			scriptTarget := "INSERT INTO tb_target (id, indikator_id, target, satuan, tahun) VALUES (?, ?, ?, ?, ?)"
			_, err := tx.ExecContext(ctx, scriptTarget, target.Id, indikator.Id, target.Target, target.Satuan, target.Tahun)
			if err != nil {
				return domain.TujuanOpd{}, err
			}
		}
	}

	return tujuanOpd, nil
}

func (repository *TujuanOpdRepositoryImpl) Update(ctx context.Context, tx *sql.Tx, tujuanOpd domain.TujuanOpd) error {
	// Update tujuan OPD
	script := "UPDATE tb_tujuan_opd SET kode_opd = ?, kode_bidang_urusan = ?, tujuan = ?, periode_id = ?, tahun_awal = ?, tahun_akhir = ?, jenis_periode = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, script,
		tujuanOpd.KodeOpd,
		tujuanOpd.KodeBidangUrusan,
		tujuanOpd.Tujuan,
		tujuanOpd.PeriodeId.Id,
		tujuanOpd.TahunAwal,
		tujuanOpd.TahunAkhir,
		tujuanOpd.JenisPeriode,
		tujuanOpd.Id)
	if err != nil {
		return err
	}

	// Hapus indikator dan target lama
	scriptDeleteTarget := `
        DELETE t FROM tb_target t
        INNER JOIN tb_indikator i ON t.indikator_id = i.id
        WHERE i.tujuan_opd_id = ?
    `
	_, err = tx.ExecContext(ctx, scriptDeleteTarget, tujuanOpd.Id)
	if err != nil {
		return err
	}

	scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE tujuan_opd_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeleteIndikator, tujuanOpd.Id)
	if err != nil {
		return err
	}

	// Insert indikator dan target baru
	for _, indikator := range tujuanOpd.Indikator {
		// Insert indikator
		scriptIndikator := "INSERT INTO tb_indikator (id, tujuan_opd_id, indikator, rumus_perhitungan, sumber_data) VALUES (?, ?, ?, ?, ?)"
		_, err = tx.ExecContext(ctx, scriptIndikator,
			indikator.Id,
			tujuanOpd.Id,
			indikator.Indikator,
			indikator.RumusPerhitungan,
			indikator.SumberData)
		if err != nil {
			return err
		}

		// Insert target untuk setiap indikator
		for _, target := range indikator.Target {
			scriptTarget := "INSERT INTO tb_target (id, indikator_id, target, satuan, tahun) VALUES (?, ?, ?, ?, ?)"
			_, err = tx.ExecContext(ctx, scriptTarget,
				target.Id,
				indikator.Id,
				target.Target,
				target.Satuan,
				target.Tahun)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (repository *TujuanOpdRepositoryImpl) Delete(ctx context.Context, tx *sql.Tx, tujuanOpdId int) error {
	scriptDeleteTarget := `
        DELETE t FROM tb_target t
        INNER JOIN tb_indikator i ON t.indikator_id = i.id
        WHERE i.tujuan_opd_id = ?
    `
	_, err := tx.ExecContext(ctx, scriptDeleteTarget, tujuanOpdId)
	if err != nil {
		return err
	}

	scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE tujuan_opd_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeleteIndikator, tujuanOpdId)
	if err != nil {
		return err
	}

	scriptDeleteTujuanOpd := "DELETE FROM tb_tujuan_opd WHERE id = ?"
	_, err = tx.ExecContext(ctx, scriptDeleteTujuanOpd, tujuanOpdId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *TujuanOpdRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, tujuanOpdId int) (domain.TujuanOpd, error) {
	script := `
        SELECT 
            t.id, 
            t.kode_opd,
            COALESCE(t.kode_bidang_urusan, '') as kode_bidang_urusan,
            t.tujuan, 
            t.tahun_awal,
            t.tahun_akhir,
            t.jenis_periode,
            i.id as indikator_id,
            i.indikator,
            i.rumus_perhitungan, 
            i.sumber_data,
            COALESCE(tg.id, '') as target_id,
            COALESCE(tg.target, '') as target_value,
            COALESCE(tg.satuan, '') as satuan,
            COALESCE(tg.tahun, '') as tahun_target
        FROM tb_tujuan_opd t
        LEFT JOIN tb_indikator i ON t.id = i.tujuan_opd_id
        LEFT JOIN tb_target tg ON i.id = tg.indikator_id
        WHERE t.id = ?
        ORDER BY i.id ASC, tg.tahun ASC
    `

	rows, err := tx.QueryContext(ctx, script, tujuanOpdId)
	if err != nil {
		return domain.TujuanOpd{}, err
	}
	defer rows.Close()

	var tujuanOpd *domain.TujuanOpd
	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			id               int
			kodeOpd          string
			kodeBidangUrusan string
			tujuan           string
			tahunAwal        string
			tahunAkhir       string
			jenisPeriode     string
			indikatorId      sql.NullString
			indikatorNama    sql.NullString
			rumusPerhitungan sql.NullString
			sumberData       sql.NullString
			targetId         sql.NullString
			targetValue      sql.NullString
			satuan           sql.NullString
			tahunTarget      sql.NullString
		)

		err := rows.Scan(
			&id,
			&kodeOpd,
			&kodeBidangUrusan,
			&tujuan,
			&tahunAwal,
			&tahunAkhir,
			&jenisPeriode,
			&indikatorId,
			&indikatorNama,
			&rumusPerhitungan,
			&sumberData,
			&targetId,
			&targetValue,
			&satuan,
			&tahunTarget,
		)
		if err != nil {
			return domain.TujuanOpd{}, err
		}

		if tujuanOpd == nil {
			tujuanOpd = &domain.TujuanOpd{
				Id:               id,
				KodeOpd:          kodeOpd,
				KodeBidangUrusan: kodeBidangUrusan,
				Tujuan:           tujuan,
				TahunAwal:        tahunAwal,
				TahunAkhir:       tahunAkhir,
				JenisPeriode:     jenisPeriode,
				Indikator:        []domain.Indikator{},
			}
		}

		if indikatorId.Valid {
			if _, exists := indikatorMap[indikatorId.String]; !exists {
				indikatorMap[indikatorId.String] = &domain.Indikator{
					Id:               indikatorId.String,
					Indikator:        indikatorNama.String,
					RumusPerhitungan: rumusPerhitungan,
					SumberData:       sumberData,
					Target:           []domain.Target{},
				}
				tujuanOpd.Indikator = append(tujuanOpd.Indikator, *indikatorMap[indikatorId.String])
			}

			if targetId.Valid && tahunTarget.Valid {
				tahunTargetInt, _ := strconv.Atoi(tahunTarget.String)
				tahunAwalInt, _ := strconv.Atoi(tahunAwal)
				tahunAkhirInt, _ := strconv.Atoi(tahunAkhir)

				// Hanya tambahkan target jika tahunnya dalam range
				if tahunTargetInt >= tahunAwalInt && tahunTargetInt <= tahunAkhirInt {
					target := domain.Target{
						Id:          targetId.String,
						IndikatorId: indikatorId.String,
						Target:      targetValue.String,
						Satuan:      satuan.String,
						Tahun:       tahunTarget.String,
					}
					for i := range tujuanOpd.Indikator {
						if tujuanOpd.Indikator[i].Id == indikatorId.String {
							tujuanOpd.Indikator[i].Target = append(tujuanOpd.Indikator[i].Target, target)
							break
						}
					}
				}
			}
		}
	}

	if tujuanOpd == nil {
		return domain.TujuanOpd{}, fmt.Errorf("tujuan opd with id %d not found", tujuanOpdId)
	}

	// Generate target lengkap untuk setiap indikator
	for i := range tujuanOpd.Indikator {
		tahunAwalInt, _ := strconv.Atoi(tujuanOpd.TahunAwal)
		tahunAkhirInt, _ := strconv.Atoi(tujuanOpd.TahunAkhir)

		// Buat map untuk target yang sudah ada
		existingTargets := make(map[string]domain.Target)
		for _, target := range tujuanOpd.Indikator[i].Target {
			existingTargets[target.Tahun] = target
		}

		// Reset dan generate ulang target
		var completeTargets []domain.Target
		for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
			tahunStr := strconv.Itoa(tahun)
			if target, exists := existingTargets[tahunStr]; exists {
				completeTargets = append(completeTargets, target)
			} else {
				completeTargets = append(completeTargets, domain.Target{
					Id:          "",
					IndikatorId: tujuanOpd.Indikator[i].Id,
					Target:      "",
					Satuan:      "",
					Tahun:       tahunStr,
				})
			}
		}

		// Sort target berdasarkan tahun
		sort.Slice(completeTargets, func(i, j int) bool {
			tahunI, _ := strconv.Atoi(completeTargets[i].Tahun)
			tahunJ, _ := strconv.Atoi(completeTargets[j].Tahun)
			return tahunI < tahunJ
		})

		tujuanOpd.Indikator[i].Target = completeTargets
	}

	return *tujuanOpd, nil
}

func (repository *TujuanOpdRepositoryImpl) FindIndikatorByTujuanId(ctx context.Context, tx *sql.Tx, tujuanOpdId int) ([]domain.Indikator, error) {
	script := `SELECT id, indikator 
               FROM tb_indikator 
               WHERE tujuan_opd_id = ?`

	rows, err := tx.QueryContext(ctx, script, tujuanOpdId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indikators []domain.Indikator
	for rows.Next() {
		var indikator domain.Indikator
		err := rows.Scan(&indikator.Id, &indikator.Indikator)
		if err != nil {
			return nil, err
		}
		indikators = append(indikators, indikator)
	}

	return indikators, nil
}

func (repository *TujuanOpdRepositoryImpl) FindTargetByIndikatorId(ctx context.Context, tx *sql.Tx, indikatorId string, tahun string) ([]domain.Target, error) {
	script := `
        SELECT id, target, satuan, tahun
        FROM tb_target 
        WHERE indikator_id = ?
        AND tahun = ?  -- Ubah dari tahun <= ? menjadi tahun = ?
        ORDER BY tahun ASC
    `

	rows, err := tx.QueryContext(ctx, script, indikatorId, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []domain.Target
	for rows.Next() {
		var target domain.Target
		err := rows.Scan(
			&target.Id,
			&target.Target,
			&target.Satuan,
			&target.Tahun,
		)
		if err != nil {
			return nil, err
		}
		target.IndikatorId = indikatorId
		targets = append(targets, target)
	}

	return targets, nil
}

func (repository *TujuanOpdRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, kodeOpd string, tahunAwal string, tahunAkhir string, jenisPeriode string) ([]domain.TujuanOpd, error) {
	scriptTujuan := `
        SELECT 
            t.id, 
            t.kode_opd,
            COALESCE(t.kode_bidang_urusan, '') as kode_bidang_urusan,
            t.tujuan, 
            t.tahun_awal,
            t.tahun_akhir,
            t.jenis_periode,
            i.id as indikator_id,
            i.indikator,
            i.rumus_perhitungan, 
            i.sumber_data,
            tg.id as target_id,
            tg.target as target_value,
            tg.satuan,
            tg.tahun as tahun_target
        FROM tb_tujuan_opd t
        LEFT JOIN tb_indikator i ON t.id = i.tujuan_opd_id
        LEFT JOIN tb_target tg ON i.id = tg.indikator_id
        WHERE t.kode_opd = ? 
        AND t.tahun_awal = ?
        AND t.tahun_akhir = ?
        AND t.jenis_periode = ?
        AND (tg.tahun IS NULL OR (CAST(tg.tahun AS SIGNED) BETWEEN CAST(? AS SIGNED) AND CAST(? AS SIGNED)))
        ORDER BY t.id ASC, i.id ASC, tg.tahun ASC
    `

	rows, err := tx.QueryContext(ctx, scriptTujuan,
		kodeOpd,
		tahunAwal,
		tahunAkhir,
		jenisPeriode,
		tahunAwal,
		tahunAkhir,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tujuanOpdMap := make(map[int]*domain.TujuanOpd)
	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var (
			tujuanId         int
			kodeOpd          string
			kodeBidangUrusan string
			tujuan           string
			// periodeId        int
			tahunAwalData    string
			tahunAkhirData   string
			jenisPeriodeData string
			indikatorId      sql.NullString
			indikatorNama    sql.NullString
			rumusPerhitungan sql.NullString
			sumberData       sql.NullString
			targetId         sql.NullString
			targetValue      sql.NullString
			satuan           sql.NullString
			tahunTarget      sql.NullString
		)

		err := rows.Scan(
			&tujuanId,
			&kodeOpd,
			&kodeBidangUrusan,
			&tujuan,
			// &periodeId,
			&tahunAwalData,
			&tahunAkhirData,
			&jenisPeriodeData,
			&indikatorId,
			&indikatorNama,
			&rumusPerhitungan,
			&sumberData,
			&targetId,
			&targetValue,
			&satuan,
			&tahunTarget,
		)
		if err != nil {
			return nil, err
		}

		// Buat atau ambil TujuanOpd
		if _, exists := tujuanOpdMap[tujuanId]; !exists {
			tujuanOpdMap[tujuanId] = &domain.TujuanOpd{
				Id:               tujuanId,
				KodeOpd:          kodeOpd,
				KodeBidangUrusan: kodeBidangUrusan,
				Tujuan:           tujuan,
				TahunAwal:        tahunAwalData,
				TahunAkhir:       tahunAkhirData,
				JenisPeriode:     jenisPeriodeData,
				Indikator:        []domain.Indikator{},
			}
		}

		// Buat atau ambil Indikator jika ada
		if indikatorId.Valid {
			if _, exists := indikatorMap[indikatorId.String]; !exists {
				indikatorMap[indikatorId.String] = &domain.Indikator{
					Id:               indikatorId.String,
					Indikator:        indikatorNama.String,
					RumusPerhitungan: rumusPerhitungan,
					SumberData:       sumberData,
					Target:           []domain.Target{},
				}
				tujuanOpdMap[tujuanId].Indikator = append(tujuanOpdMap[tujuanId].Indikator, *indikatorMap[indikatorId.String])
			}

			// Tambahkan target jika ada
			if targetId.Valid && tahunTarget.Valid {
				target := domain.Target{
					Id:          targetId.String,
					IndikatorId: indikatorId.String,
					Target:      targetValue.String,
					Satuan:      satuan.String,
					Tahun:       tahunTarget.String,
				}
				// Update target langsung ke indikator di map
				for i := range tujuanOpdMap[tujuanId].Indikator {
					if tujuanOpdMap[tujuanId].Indikator[i].Id == indikatorId.String {
						tujuanOpdMap[tujuanId].Indikator[i].Target = append(tujuanOpdMap[tujuanId].Indikator[i].Target, target)
						break
					}
				}
			}
		}
	}

	// Perbaikan pada bagian generate target
	var result []domain.TujuanOpd
	for _, tujuanOpd := range tujuanOpdMap {
		for i := range tujuanOpd.Indikator {
			tahunAwalInt, _ := strconv.Atoi(tujuanOpd.TahunAwal)
			tahunAkhirInt, _ := strconv.Atoi(tujuanOpd.TahunAkhir)

			// Buat map untuk target yang sudah ada
			existingTargets := make(map[string]domain.Target)
			for _, target := range tujuanOpd.Indikator[i].Target {
				if target.Id != "" { // Hanya masukkan target yang valid
					existingTargets[target.Tahun] = target
				}
			}

			// Reset target array
			var completeTargets []domain.Target

			// Generate target untuk setiap tahun dalam range
			for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
				tahunStr := strconv.Itoa(tahun)
				if target, exists := existingTargets[tahunStr]; exists {
					// Gunakan target yang sudah ada
					completeTargets = append(completeTargets, target)
				} else {
					// Buat target kosong dengan tahun yang sesuai
					completeTargets = append(completeTargets, domain.Target{
						Id:          "",
						IndikatorId: tujuanOpd.Indikator[i].Id,
						Target:      "",
						Satuan:      "",
						Tahun:       tahunStr,
					})
				}
			}

			// Sort target berdasarkan tahun
			sort.Slice(completeTargets, func(i, j int) bool {
				tahunI, _ := strconv.Atoi(completeTargets[i].Tahun)
				tahunJ, _ := strconv.Atoi(completeTargets[j].Tahun)
				return tahunI < tahunJ
			})

			tujuanOpd.Indikator[i].Target = completeTargets
		}
		result = append(result, *tujuanOpd)
	}

	if len(result) == 0 {
		return make([]domain.TujuanOpd, 0), nil
	}

	return result, nil
}

func (repository *TujuanOpdRepositoryImpl) FindTujuanOpdByTahun(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string, jenisPeriode string) ([]domain.TujuanOpd, error) {
	script := `
		SELECT
			t.id,
			t.kode_opd,
			t.tujuan,
			t.tahun_awal,
			t.tahun_akhir,
			t.jenis_periode
		FROM tb_tujuan_opd t
		INNER JOIN tb_periode p ON 
			t.tahun_awal = p.tahun_awal 
			AND t.tahun_akhir = p.tahun_akhir 
			AND t.jenis_periode = p.jenis_periode
		WHERE t.kode_opd = ?
		AND CAST(? AS SIGNED) BETWEEN CAST(t.tahun_awal AS SIGNED) AND CAST(t.tahun_akhir AS SIGNED)
		AND t.jenis_periode = ?
		ORDER BY t.id ASC
    `

	rows, err := tx.QueryContext(ctx, script, kodeOpd, tahun, jenisPeriode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tujuanOpds []domain.TujuanOpd
	for rows.Next() {
		var tujuanOpd domain.TujuanOpd
		err := rows.Scan(
			&tujuanOpd.Id,
			&tujuanOpd.KodeOpd,
			&tujuanOpd.Tujuan,
			&tujuanOpd.TahunAwal,
			&tujuanOpd.TahunAkhir,
			&tujuanOpd.JenisPeriode,
		)
		if err != nil {
			return nil, err
		}
		tujuanOpds = append(tujuanOpds, tujuanOpd)
	}

	// Untuk setiap tujuan OPD, ambil indikatornya
	for i := range tujuanOpds {
		indikators, err := repository.FindIndikatorByTujuanOpdId(ctx, tx, tujuanOpds[i].Id)
		if err != nil {
			return nil, err
		}
		tujuanOpds[i].Indikator = indikators

		// Generate target untuk setiap indikator
		for j := range tujuanOpds[i].Indikator {
			// Ambil target yang sudah ada dari database
			existingTargets, err := repository.FindTargetByIndikatorId(ctx, tx, tujuanOpds[i].Indikator[j].Id, tujuanOpds[i].TahunAwal)
			if err != nil {
				return nil, err
			}

			// Buat map untuk target yang sudah ada
			targetMap := make(map[string]domain.Target)
			for _, target := range existingTargets {
				targetMap[target.Tahun] = target
			}

			// Generate target untuk semua tahun dalam periode
			tahunAwalInt, _ := strconv.Atoi(tujuanOpds[i].TahunAwal)
			tahunAkhirInt, _ := strconv.Atoi(tujuanOpds[i].TahunAkhir)

			var targets []domain.Target
			for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
				tahunStr := strconv.Itoa(tahun)
				// Jika target untuk tahun ini sudah ada, gunakan yang ada
				if existingTarget, exists := targetMap[tahunStr]; exists {
					targets = append(targets, existingTarget)
				} else {
					// Jika belum ada, buat target kosong
					targets = append(targets, domain.Target{
						Id:          "",
						IndikatorId: tujuanOpds[i].Indikator[j].Id,
						Target:      "",
						Satuan:      "",
						Tahun:       tahunStr,
					})
				}
			}
			tujuanOpds[i].Indikator[j].Target = targets
		}
	}

	return tujuanOpds, nil
}

// Perbaikan pada FindIndikatorByTujuanOpdId untuk menyertakan rumus_perhitungan dan sumber_data
func (repository *TujuanOpdRepositoryImpl) FindIndikatorByTujuanOpdId(ctx context.Context, tx *sql.Tx, tujuanOpdId int) ([]domain.Indikator, error) {
	script := `
        SELECT 
            id, 
            indikator,
            COALESCE(rumus_perhitungan, '') as rumus_perhitungan,
            COALESCE(sumber_data, '') as sumber_data
        FROM tb_indikator
        WHERE tujuan_opd_id = ?
        ORDER BY id ASC
    `

	rows, err := tx.QueryContext(ctx, script, tujuanOpdId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indikators []domain.Indikator
	for rows.Next() {
		var indikator domain.Indikator
		var rumusPerhitungan, sumberData string

		err := rows.Scan(
			&indikator.Id,
			&indikator.Indikator,
			&rumusPerhitungan,
			&sumberData,
		)
		if err != nil {
			return nil, err
		}

		indikator.TujuanOpdId = tujuanOpdId
		indikator.RumusPerhitungan = sql.NullString{String: rumusPerhitungan, Valid: rumusPerhitungan != ""}
		indikator.SumberData = sql.NullString{String: sumberData, Valid: sumberData != ""}
		indikators = append(indikators, indikator)
	}

	return indikators, nil
}

func (repository *TujuanOpdRepositoryImpl) FindTujuanOpdForCascadingOpd(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string, jenisPeriode string) ([]domain.TujuanOpd, error) {
	script := `
		SELECT
			t.id,
			t.kode_opd,
			t.tujuan,
			t.tahun_awal,
			t.tahun_akhir,
			t.jenis_periode,
			t.kode_bidang_urusan
		FROM tb_tujuan_opd t
		INNER JOIN tb_periode p ON 
			t.tahun_awal = p.tahun_awal 
			AND t.tahun_akhir = p.tahun_akhir 
			AND t.jenis_periode = p.jenis_periode
		WHERE t.kode_opd = ?
		AND CAST(? AS SIGNED) BETWEEN CAST(p.tahun_awal AS SIGNED) AND CAST(p.tahun_akhir AS SIGNED)
		AND p.jenis_periode = ?
		ORDER BY t.id ASC
    `

	rows, err := tx.QueryContext(ctx, script, kodeOpd, tahun, jenisPeriode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tujuanOpds []domain.TujuanOpd
	for rows.Next() {
		var tujuanOpd domain.TujuanOpd
		err := rows.Scan(
			&tujuanOpd.Id,
			&tujuanOpd.KodeOpd,
			&tujuanOpd.Tujuan,
			&tujuanOpd.TahunAwal,
			&tujuanOpd.TahunAkhir,
			&tujuanOpd.JenisPeriode,
			&tujuanOpd.KodeBidangUrusan,
		)
		if err != nil {
			return nil, err
		}
		tujuanOpds = append(tujuanOpds, tujuanOpd)
	}

	// Untuk setiap tujuan OPD, ambil indikatornya
	for i := range tujuanOpds {
		indikators, err := repository.FindIndikatorByTujuanOpdId(ctx, tx, tujuanOpds[i].Id)
		if err != nil {
			return nil, err
		}
		tujuanOpds[i].Indikator = indikators

		// Generate target untuk setiap indikator
		for j := range tujuanOpds[i].Indikator {
			tahunAwalInt, _ := strconv.Atoi(tujuanOpds[i].TahunAwal)
			tahunAkhirInt, _ := strconv.Atoi(tujuanOpds[i].TahunAkhir)

			var targets []domain.Target
			for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
				targets = append(targets, domain.Target{
					Id:          "",
					IndikatorId: tujuanOpds[i].Indikator[j].Id,
					Target:      "",
					Satuan:      "",
					Tahun:       strconv.Itoa(tahun),
				})
			}
			tujuanOpds[i].Indikator[j].Target = targets
		}
	}

	return tujuanOpds, nil
}
