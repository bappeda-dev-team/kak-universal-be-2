package repository

import (
	"context"
	"database/sql"
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/model/domain"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type PohonKinerjaRepositoryImpl struct {
}

func NewPohonKinerjaRepositoryImpl() *PohonKinerjaRepositoryImpl {
	return &PohonKinerjaRepositoryImpl{}
}

func (repository *PohonKinerjaRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, pohonKinerja domain.PohonKinerja) (domain.PohonKinerja, error) {
	scriptPokin := "INSERT INTO tb_pohon_kinerja (nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := tx.ExecContext(ctx, scriptPokin,
		pohonKinerja.NamaPohon,
		pohonKinerja.Parent,
		pohonKinerja.JenisPohon,
		pohonKinerja.LevelPohon,
		pohonKinerja.KodeOpd,
		pohonKinerja.Keterangan,
		pohonKinerja.Tahun,
		pohonKinerja.Status)
	if err != nil {
		return pohonKinerja, err
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return pohonKinerja, err
	}
	pohonKinerja.Id = int(lastInsertId)

	scriptPelaksana := "INSERT INTO tb_pelaksana_pokin (id, pohon_kinerja_id, pegawai_id) VALUES (?, ?, ?)"
	for _, pelaksana := range pohonKinerja.Pelaksana {
		_, err = tx.ExecContext(ctx, scriptPelaksana,
			pelaksana.Id,                // id pelaksana_pokin
			fmt.Sprint(pohonKinerja.Id), // pohon_kinerja_id dalam string
			pelaksana.PegawaiId)         // pegawai_id
		if err != nil {
			return pohonKinerja, err
		}
	}

	// Insert indikator
	for _, indikator := range pohonKinerja.Indikator {
		scriptIndikator := "INSERT INTO tb_indikator (id, pokin_id, indikator, tahun) VALUES (?, ?, ?, ?)"
		_, err := tx.ExecContext(ctx, scriptIndikator,
			indikator.Id,
			pohonKinerja.Id,
			indikator.Indikator,
			indikator.Tahun)
		if err != nil {
			return pohonKinerja, err
		}

		// Insert target untuk setiap indikator
		for _, target := range indikator.Target {
			scriptTarget := "INSERT INTO tb_target (id, indikator_id, target, satuan, tahun) VALUES (?, ?, ?, ?, ?)"
			_, err := tx.ExecContext(ctx, scriptTarget,
				target.Id,
				indikator.Id,
				target.Target,
				target.Satuan,
				target.Tahun)
			if err != nil {
				return pohonKinerja, err
			}
		}
	}

	return pohonKinerja, nil
}

func (repository *PohonKinerjaRepositoryImpl) Update(ctx context.Context, tx *sql.Tx, pohonKinerja domain.PohonKinerja) (domain.PohonKinerja, error) {
	// Update tb_pohon_kinerja
	scriptPokin := `
        UPDATE tb_pohon_kinerja 
        SET nama_pohon = ?, 
            parent = CASE 
                WHEN clone_from = 0 THEN ? 
                ELSE parent 
            END,
            jenis_pohon = ?, 
            level_pohon = ?, 
            kode_opd = ?, 
            keterangan = ?, 
            tahun = ?
        WHERE id = ?`

	_, err := tx.ExecContext(ctx, scriptPokin,
		pohonKinerja.NamaPohon,
		pohonKinerja.Parent,
		pohonKinerja.JenisPohon,
		pohonKinerja.LevelPohon,
		pohonKinerja.KodeOpd,
		pohonKinerja.Keterangan,
		pohonKinerja.Tahun,
		pohonKinerja.Id)
	if err != nil {
		return pohonKinerja, err
	}

	// Update pelaksana
	scriptDeletePelaksana := "DELETE FROM tb_pelaksana_pokin WHERE pohon_kinerja_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeletePelaksana, fmt.Sprint(pohonKinerja.Id))
	if err != nil {
		return pohonKinerja, err
	}

	for _, pelaksana := range pohonKinerja.Pelaksana {
		scriptPelaksana := "INSERT INTO tb_pelaksana_pokin (id, pohon_kinerja_id, pegawai_id) VALUES (?, ?, ?)"
		_, err := tx.ExecContext(ctx, scriptPelaksana,
			pelaksana.Id,
			fmt.Sprint(pohonKinerja.Id),
			pelaksana.PegawaiId)
		if err != nil {
			return pohonKinerja, err
		}
	}

	// Proses indikator
	for _, indikator := range pohonKinerja.Indikator {
		// Update atau insert indikator dengan clone_from
		scriptUpdateIndikator := `
			INSERT INTO tb_indikator (id, pokin_id, indikator, tahun, clone_from) 
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE 
				indikator = VALUES(indikator),
				tahun = VALUES(tahun),
				clone_from = VALUES(clone_from)`

		_, err := tx.ExecContext(ctx, scriptUpdateIndikator,
			indikator.Id,
			pohonKinerja.Id,
			indikator.Indikator,
			indikator.Tahun,
			indikator.CloneFrom)
		if err != nil {
			return pohonKinerja, err
		}

		// Hapus target lama untuk indikator ini
		scriptDeleteTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
		_, err = tx.ExecContext(ctx, scriptDeleteTargets, indikator.Id)
		if err != nil {
			return pohonKinerja, err
		}

		// Insert target baru dengan clone_from
		for _, target := range indikator.Target {
			// Log untuk debugging
			fmt.Printf("Inserting target: ID=%s, IndikatorID=%s, CloneFrom=%s\n",
				target.Id, target.IndikatorId, target.CloneFrom)

			scriptInsertTarget := `
				INSERT INTO tb_target 
					(id, indikator_id, target, satuan, tahun, clone_from)
				VALUES 
					(?, ?, ?, ?, ?, ?)`

			_, err := tx.ExecContext(ctx, scriptInsertTarget,
				target.Id,
				target.IndikatorId,
				target.Target,
				target.Satuan,
				target.Tahun,
				target.CloneFrom) // Pastikan clone_from dimasukkan
			if err != nil {
				return pohonKinerja, fmt.Errorf("error inserting target: %v", err)
			}
		}
	}

	// Hapus indikator yang tidak ada dalam request
	var existingIndikatorIds []string
	scriptGetExisting := "SELECT id FROM tb_indikator WHERE pokin_id = ?"
	rows, err := tx.QueryContext(ctx, scriptGetExisting, pohonKinerja.Id)
	if err != nil {
		return pohonKinerja, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return pohonKinerja, err
		}
		existingIndikatorIds = append(existingIndikatorIds, id)
	}

	// Buat map untuk indikator baru
	newIndikatorIds := make(map[string]bool)
	for _, ind := range pohonKinerja.Indikator {
		newIndikatorIds[ind.Id] = true
	}

	// Hapus indikator yang tidak ada dalam request
	for _, existingId := range existingIndikatorIds {
		if !newIndikatorIds[existingId] {
			// Hapus target terlebih dahulu
			scriptDeleteTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
			_, err = tx.ExecContext(ctx, scriptDeleteTargets, existingId)
			if err != nil {
				return pohonKinerja, err
			}

			// Kemudian hapus indikator
			scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE id = ?"
			_, err = tx.ExecContext(ctx, scriptDeleteIndikator, existingId)
			if err != nil {
				return pohonKinerja, err
			}
		}
	}

	return pohonKinerja, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error) {
	scriptPokin := `
        SELECT 
            pk.id, 
            COALESCE(pk.parent, 0) as parent, 
            COALESCE(pk.nama_pohon, '') as nama_pohon, 
            COALESCE(pk.jenis_pohon, '') as jenis_pohon, 
            COALESCE(pk.level_pohon, 0) as level_pohon, 
            COALESCE(pk.kode_opd, '') as kode_opd, 
            COALESCE(pk.keterangan, '') as keterangan, 
			COALESCE(pk.keterangan_crosscutting, '') as keterangan_crosscutting,
            COALESCE(pk.tahun, '') as tahun,
            COALESCE(pk.status, '') as status
        FROM 
            tb_pohon_kinerja pk 
        WHERE 
            pk.id = ?`

	rows, err := tx.QueryContext(ctx, scriptPokin, id)
	if err != nil {
		return domain.PohonKinerja{}, err
	}
	defer rows.Close()

	pohonKinerja := domain.PohonKinerja{}
	if rows.Next() {
		err := rows.Scan(
			&pohonKinerja.Id,
			&pohonKinerja.Parent,
			&pohonKinerja.NamaPohon,
			&pohonKinerja.JenisPohon,
			&pohonKinerja.LevelPohon,
			&pohonKinerja.KodeOpd,
			&pohonKinerja.Keterangan,
			&pohonKinerja.KeteranganCrosscutting,
			&pohonKinerja.Tahun,
			&pohonKinerja.Status,
		)
		if err != nil {
			return domain.PohonKinerja{}, err
		}
	}

	return pohonKinerja, nil
}

//pokin lama
// func (repository *PohonKinerjaRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, kodeOpd, tahun string) ([]domain.PohonKinerja, error) {
// 	script := `
//         SELECT
//             id,
//             COALESCE(nama_pohon, '') as nama_pohon,
//             COALESCE(parent, 0) as parent,
//             COALESCE(jenis_pohon, '') as jenis_pohon,
//             COALESCE(level_pohon, 0) as level_pohon,
//             COALESCE(kode_opd, '') as kode_opd,
//             COALESCE(keterangan, '') as keterangan,
//             COALESCE(keterangan_crosscutting, '') as keterangan_crosscutting,
//             COALESCE(tahun, '') as tahun,
//             COALESCE(status, '') as status,
// 			COALESCE(is_active) as is_active
//         FROM tb_pohon_kinerja
//         WHERE kode_opd = ?
// 		AND tahun = ?
// 		AND status NOT IN ('menunggu_disetujui', 'tarik pokin opd', 'disetujui', 'ditolak', 'crosscutting_menunggu', 'crosscutting_ditolak')
//         ORDER BY level_pohon, id ASC`

// 	rows, err := tx.QueryContext(ctx, script, kodeOpd, tahun)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var pokins []domain.PohonKinerja
// 	for rows.Next() {
// 		var pokin domain.PohonKinerja
// 		err := rows.Scan(
// 			&pokin.Id,
// 			&pokin.NamaPohon,
// 			&pokin.Parent,
// 			&pokin.JenisPohon,
// 			&pokin.LevelPohon,
// 			&pokin.KodeOpd,
// 			&pokin.Keterangan,
// 			&pokin.KeteranganCrosscutting,
// 			&pokin.Tahun,
// 			&pokin.Status,
// 			&pokin.IsActive,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		pokins = append(pokins, pokin)
// 	}

// 	// Inisialisasi slice kosong jika tidak ada data
// 	if pokins == nil {
// 		pokins = make([]domain.PohonKinerja, 0)
// 	}

// 	return pokins, nil
// }

// pokin baru
func (repository *PohonKinerjaRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx, kodeOpd, tahun string) ([]domain.PohonKinerja, error) {
	script := `
    SELECT 
            id,
            COALESCE(nama_pohon, '') as nama_pohon,
            COALESCE(parent, 0) as parent,
            COALESCE(jenis_pohon, '') as jenis_pohon,
            COALESCE(level_pohon, 0) as level_pohon,
            COALESCE(kode_opd, '') as kode_opd,
            COALESCE(keterangan, '') as keterangan,
            COALESCE(keterangan_crosscutting, '') as keterangan_crosscutting,
            COALESCE(tahun, '') as tahun,
            COALESCE(status, '') as status,
            COALESCE(is_active) as is_active
        FROM tb_pohon_kinerja 
        WHERE kode_opd = ? 
        AND tahun = ?
        AND status NOT IN ('menunggu_disetujui', 'tarik pokin opd', 'disetujui', 'ditolak', 'crosscutting_menunggu', 'crosscutting_ditolak')
        ORDER BY 
            CASE 
                WHEN status = 'pokin dari pemda' THEN 0 
                ELSE 1 
            END,
            level_pohon, 
            id ASC`

	rows, err := tx.QueryContext(ctx, script, kodeOpd, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokins []domain.PohonKinerja
	for rows.Next() {
		var pokin domain.PohonKinerja
		err := rows.Scan(
			&pokin.Id,
			&pokin.NamaPohon,
			&pokin.Parent,
			&pokin.JenisPohon,
			&pokin.LevelPohon,
			&pokin.KodeOpd,
			&pokin.Keterangan,
			&pokin.KeteranganCrosscutting,
			&pokin.Tahun,
			&pokin.Status,
			&pokin.IsActive,
		)
		if err != nil {
			return nil, err
		}
		pokins = append(pokins, pokin)
	}

	// Inisialisasi slice kosong jika tidak ada data
	if pokins == nil {
		pokins = make([]domain.PohonKinerja, 0)
	}

	return pokins, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindStrategicNoParent(ctx context.Context, tx *sql.Tx, levelPohon, parent int, kodeOpd, tahun string) ([]domain.PohonKinerja, error) {
	script := "SELECT id, nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun FROM tb_pohon_kinerja WHERE level_pohon = ? AND parent = ? AND kode_opd = ? AND tahun = ?"
	rows, err := tx.QueryContext(ctx, script, levelPohon, parent, kodeOpd, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.PohonKinerja
	for rows.Next() {
		var pokin domain.PohonKinerja
		err := rows.Scan(&pokin.Id, &pokin.NamaPohon, &pokin.Parent, &pokin.JenisPohon, &pokin.LevelPohon, &pokin.KodeOpd, &pokin.Keterangan, &pokin.Tahun)
		if err != nil {
			return nil, err
		}
		result = append(result, pokin)
	}
	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	// Temukan semua ID anak pohon secara rekursif
	scriptFindChildren := `
        WITH RECURSIVE child_tree AS (
            SELECT id, parent, clone_from, level_pohon
            FROM tb_pohon_kinerja
            WHERE id = ?

            UNION ALL

            SELECT t.id, t.parent, t.clone_from, t.level_pohon
            FROM tb_pohon_kinerja t
            JOIN child_tree ct ON t.parent = ct.id
        )
        SELECT id, clone_from FROM child_tree
    `

	rows, err := tx.QueryContext(ctx, scriptFindChildren, id)
	if err != nil {
		return fmt.Errorf("gagal mencari turunan pohon: %v", err)
	}
	defer rows.Close()

	// Kumpulkan semua ID anak pohon dan clone_from
	var nodeIds []string
	cloneFromMap := make(map[string]int) // map[nodeId]cloneFromId
	for rows.Next() {
		var nodeId string
		var cloneFrom sql.NullInt64
		if err := rows.Scan(&nodeId, &cloneFrom); err != nil {
			return fmt.Errorf("gagal membaca ID turunan pohon: %v", err)
		}
		nodeIds = append(nodeIds, nodeId)
		if cloneFrom.Valid {
			cloneFromMap[nodeId] = int(cloneFrom.Int64)
		}
	}

	// Update status untuk semua node asli yang di-clone
	for _, cloneFromId := range cloneFromMap {
		scriptUpdateStatus := `
            UPDATE tb_pohon_kinerja 
            SET status = 'menunggu_disetujui'
            WHERE id = ?
        `
		if _, err := tx.ExecContext(ctx, scriptUpdateStatus, cloneFromId); err != nil {
			return fmt.Errorf("gagal mengupdate status node asli ID %d: %v", cloneFromId, err)
		}
	}

	// Proses penghapusan untuk setiap node
	for _, nodeId := range nodeIds {
		// 1. Hapus target
		scriptDeleteTarget := `
            DELETE FROM tb_target 
            WHERE indikator_id IN (
                SELECT id FROM tb_indikator 
                WHERE pokin_id = ?
            )`
		if _, err := tx.ExecContext(ctx, scriptDeleteTarget, nodeId); err != nil {
			return fmt.Errorf("gagal menghapus target: %v", err)
		}

		// 2. Hapus indikator
		scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE pokin_id = ?"
		if _, err := tx.ExecContext(ctx, scriptDeleteIndikator, nodeId); err != nil {
			return fmt.Errorf("gagal menghapus indikator: %v", err)
		}

		// 3. Hapus pelaksana
		scriptDeletePelaksana := "DELETE FROM tb_pelaksana_pokin WHERE pohon_kinerja_id = ?"
		if _, err := tx.ExecContext(ctx, scriptDeletePelaksana, nodeId); err != nil {
			return fmt.Errorf("gagal menghapus pelaksana: %v", err)
		}

		// 4. Tangani crosscutting
		// Hapus crosscutting yang menunggu/ditolak
		scriptDeletePendingCrosscutting := `
            DELETE FROM tb_crosscutting 
            WHERE crosscutting_from = ?
            AND status IN ('menunggu_disetujui', 'ditolak')`
		if _, err := tx.ExecContext(ctx, scriptDeletePendingCrosscutting, nodeId); err != nil {
			return fmt.Errorf("gagal menghapus crosscutting pending: %v", err)
		}

		// Update status crosscutting yang disetujui
		scriptUpdateCrosscuttingStatus := `
            UPDATE tb_crosscutting 
            SET status = 'crosscutting_ditolak', 
                crosscutting_to = 0 
            WHERE crosscutting_to = ?
            AND status = 'crosscutting_disetujui'`
		if _, err := tx.ExecContext(ctx, scriptUpdateCrosscuttingStatus, nodeId); err != nil {
			return fmt.Errorf("gagal mengupdate status crosscutting: %v", err)
		}

		// 5. Hapus pohon kinerja
		scriptDeletePokin := "DELETE FROM tb_pohon_kinerja WHERE id = ?"
		if _, err := tx.ExecContext(ctx, scriptDeletePokin, nodeId); err != nil {
			return fmt.Errorf("gagal menghapus pohon kinerja: %v", err)
		}
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) recursiveDelete(ctx context.Context, tx *sql.Tx, id string) error {
	// Temukan semua ID anak pohon secara rekursif
	scriptFindRelatedIds := `
        WITH RECURSIVE related_ids AS (
            SELECT id FROM tb_pohon_kinerja WHERE id = ?

            UNION ALL

            SELECT t.id
            FROM tb_pohon_kinerja t
            JOIN related_ids r ON t.parent = r.id
        )
        SELECT id FROM related_ids
    `

	rows, err := tx.QueryContext(ctx, scriptFindRelatedIds, id)
	if err != nil {
		return fmt.Errorf("gagal mencari ID terkait: %v", err)
	}
	defer rows.Close()

	// Kumpulkan semua ID terkait
	var relatedIds []string
	for rows.Next() {
		var relatedId string
		if err := rows.Scan(&relatedId); err != nil {
			return fmt.Errorf("gagal membaca ID terkait: %v", err)
		}
		relatedIds = append(relatedIds, relatedId)
	}

	if len(relatedIds) == 0 {
		// Jika tidak ada ID terkait, tambahkan ID asli
		relatedIds = append(relatedIds, id)
	}

	// Hapus data yang terkait
	return repository.deleteRelatedData(ctx, tx, relatedIds)
}

func (repository *PohonKinerjaRepositoryImpl) deleteRelatedData(ctx context.Context, tx *sql.Tx, ids []string) error {
	if len(ids) == 0 {
		return nil // Return early jika tidak ada data yang perlu dihapus
	}

	// Hapus crosscutting yang terkait dengan pohon kinerja yang akan dihapus
	scriptDeleteCrosscutting := `
        DELETE FROM tb_crosscutting 
        WHERE crosscutting_from IN (` + placeholders(len(ids)) + `)
        OR crosscutting_to IN (` + placeholders(len(ids)) + `)`
	if _, err := tx.ExecContext(ctx, scriptDeleteCrosscutting,
		append(convertToInterface(ids), convertToInterface(ids)...)...); err != nil {
		return fmt.Errorf("gagal menghapus crosscutting: %v", err)
	}

	// Hapus target
	scriptDeleteTarget := `
        DELETE FROM tb_target 
        WHERE indikator_id IN (
            SELECT id FROM tb_indikator 
            WHERE pokin_id IN (` + placeholders(len(ids)) + `)
        )`
	if _, err := tx.ExecContext(ctx, scriptDeleteTarget, convertToInterface(ids)...); err != nil {
		return fmt.Errorf("gagal menghapus target: %v", err)
	}

	// Hapus indikator
	scriptDeleteIndikator := `
        DELETE FROM tb_indikator 
        WHERE pokin_id IN (` + placeholders(len(ids)) + `)`
	if _, err := tx.ExecContext(ctx, scriptDeleteIndikator, convertToInterface(ids)...); err != nil {
		return fmt.Errorf("gagal menghapus indikator: %v", err)
	}

	// Hapus pelaksana
	scriptDeletePelaksana := `
        DELETE FROM tb_pelaksana_pokin 
        WHERE pohon_kinerja_id IN (` + placeholders(len(ids)) + `)`
	if _, err := tx.ExecContext(ctx, scriptDeletePelaksana, convertToInterface(ids)...); err != nil {
		return fmt.Errorf("gagal menghapus pelaksana: %v", err)
	}

	// Hapus pohon kinerja22
	scriptDeletePokin := `
        DELETE FROM tb_pohon_kinerja 
        WHERE id IN (` + placeholders(len(ids)) + `)`
	if _, err := tx.ExecContext(ctx, scriptDeletePokin, convertToInterface(ids)...); err != nil {
		return fmt.Errorf("gagal menghapus pohon kinerja: %v", err)
	}

	return nil
}

// Fungsi untuk membuat placeholder dinamis
func placeholders(n int) string {
	ph := make([]string, n)
	for i := range ph {
		ph[i] = "?"
	}
	return strings.Join(ph, ", ")
}

// Fungsi untuk mengonversi slice string ke slice interface{}
func convertToInterface(ids []string) []interface{} {
	result := make([]interface{}, len(ids))
	for i, id := range ids {
		result[i] = id
	}
	return result
}

func (repository *PohonKinerjaRepositoryImpl) FindPelaksanaPokin(ctx context.Context, tx *sql.Tx, pohonKinerjaId string) ([]domain.PelaksanaPokin, error) {
	script := "SELECT id, pohon_kinerja_id, pegawai_id FROM tb_pelaksana_pokin WHERE pohon_kinerja_id = ?"
	rows, err := tx.QueryContext(ctx, script, pohonKinerjaId)
	helper.PanicIfError(err)
	defer rows.Close()

	var result []domain.PelaksanaPokin
	for rows.Next() {
		var pelaksana domain.PelaksanaPokin
		err := rows.Scan(&pelaksana.Id, &pelaksana.PohonKinerjaId, &pelaksana.PegawaiId)
		helper.PanicIfError(err)
		result = append(result, pelaksana)
	}
	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) DeletePelaksanaPokin(ctx context.Context, tx *sql.Tx, pelaksanaId string) error {
	script := "DELETE FROM tb_pelaksana_pokin WHERE id = ?"
	_, err := tx.ExecContext(ctx, script, pelaksanaId)
	return err
}

// admin pokin
func (repository *PohonKinerjaRepositoryImpl) CreatePokinAdmin(ctx context.Context, tx *sql.Tx, pokinAdmin domain.PohonKinerja) (domain.PohonKinerja, error) {
	// Insert pohon kinerja tanpa ID
	scriptPokin := "INSERT INTO tb_pohon_kinerja (nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := tx.ExecContext(ctx, scriptPokin,
		pokinAdmin.NamaPohon, pokinAdmin.Parent, pokinAdmin.JenisPohon, pokinAdmin.LevelPohon, pokinAdmin.KodeOpd, pokinAdmin.Keterangan, pokinAdmin.Tahun, pokinAdmin.Status)
	if err != nil {
		return pokinAdmin, err
	}

	// Dapatkan ID yang baru dibuat
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return pokinAdmin, err
	}
	pokinAdmin.Id = int(lastInsertId)

	// Tambahkan insert pelaksana
	scriptPelaksana := "INSERT INTO tb_pelaksana_pokin (id, pohon_kinerja_id, pegawai_id) VALUES (?, ?, ?)"
	for _, pelaksana := range pokinAdmin.Pelaksana {
		_, err = tx.ExecContext(ctx, scriptPelaksana,
			pelaksana.Id,
			fmt.Sprint(pokinAdmin.Id),
			pelaksana.PegawaiId)
		if err != nil {
			return pokinAdmin, err
		}
	}

	// Insert indikator
	for _, indikator := range pokinAdmin.Indikator {
		scriptIndikator := "INSERT INTO tb_indikator (id, pokin_id, indikator, tahun) VALUES (?, ?, ?, ?)"
		_, err := tx.ExecContext(ctx, scriptIndikator,
			indikator.Id, pokinAdmin.Id, indikator.Indikator,
			indikator.Tahun)
		if err != nil {
			return pokinAdmin, err
		}

		// Insert target untuk setiap indikator
		for _, target := range indikator.Target {
			scriptTarget := "INSERT INTO tb_target (id, indikator_id, target, satuan, tahun) VALUES (?, ?, ?, ?, ?)"
			_, err := tx.ExecContext(ctx, scriptTarget, target.Id, indikator.Id, target.Target, target.Satuan, target.Tahun)
			if err != nil {
				return pokinAdmin, err
			}
		}
	}

	return pokinAdmin, nil
}

func (repository *PohonKinerjaRepositoryImpl) UpdatePokinAdmin(ctx context.Context, tx *sql.Tx, pokinAdmin domain.PohonKinerja) (domain.PohonKinerja, error) {
	// Update tb_pohon_kinerja dengan mempertahankan status
	scriptPokin := `
        UPDATE tb_pohon_kinerja 
        SET nama_pohon = ?, 
            parent = CASE 
                WHEN clone_from = 0 THEN ? 
                ELSE parent 
            END,
            jenis_pohon = ?, 
            level_pohon = ?, 
            kode_opd = ?, 
            keterangan = ?, 
            tahun = ?
        WHERE id = ?`

	_, err := tx.ExecContext(ctx, scriptPokin,
		pokinAdmin.NamaPohon,
		pokinAdmin.Parent,
		pokinAdmin.JenisPohon,
		pokinAdmin.LevelPohon,
		pokinAdmin.KodeOpd,
		pokinAdmin.Keterangan,
		pokinAdmin.Tahun,
		pokinAdmin.Id)
	if err != nil {
		return pokinAdmin, err
	}

	// Update pelaksana
	scriptDeletePelaksana := "DELETE FROM tb_pelaksana_pokin WHERE pohon_kinerja_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeletePelaksana, fmt.Sprint(pokinAdmin.Id))
	if err != nil {
		return pokinAdmin, err
	}

	for _, pelaksana := range pokinAdmin.Pelaksana {
		scriptPelaksana := "INSERT INTO tb_pelaksana_pokin (id, pohon_kinerja_id, pegawai_id) VALUES (?, ?, ?)"
		_, err := tx.ExecContext(ctx, scriptPelaksana,
			pelaksana.Id,
			fmt.Sprint(pokinAdmin.Id),
			pelaksana.PegawaiId)
		if err != nil {
			return pokinAdmin, err
		}
	}

	// Proses indikator
	for _, indikator := range pokinAdmin.Indikator {
		// Update atau insert indikator
		scriptUpdateIndikator := `
			INSERT INTO tb_indikator (id, pokin_id, indikator, tahun, clone_from) 
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE 
				indikator = VALUES(indikator),
				tahun = VALUES(tahun)`

		_, err := tx.ExecContext(ctx, scriptUpdateIndikator,
			indikator.Id,
			pokinAdmin.Id,
			indikator.Indikator,
			indikator.Tahun,
			indikator.CloneFrom)
		if err != nil {
			return pokinAdmin, err
		}

		// Hapus target lama untuk indikator ini
		scriptDeleteTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
		_, err = tx.ExecContext(ctx, scriptDeleteTargets, indikator.Id)
		if err != nil {
			return pokinAdmin, err
		}

		// Insert target baru
		for _, target := range indikator.Target {
			scriptInsertTarget := `
				INSERT INTO tb_target (id, indikator_id, target, satuan, tahun, clone_from)
				VALUES (?, ?, ?, ?, ?, ?)`

			_, err := tx.ExecContext(ctx, scriptInsertTarget,
				target.Id,
				indikator.Id,
				target.Target,
				target.Satuan,
				target.Tahun,
				target.CloneFrom)
			if err != nil {
				return pokinAdmin, err
			}
		}
	}

	// Hapus indikator yang tidak ada dalam request
	var existingIndikatorIds []string
	scriptGetExisting := "SELECT id FROM tb_indikator WHERE pokin_id = ?"
	rows, err := tx.QueryContext(ctx, scriptGetExisting, pokinAdmin.Id)
	if err != nil {
		return pokinAdmin, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return pokinAdmin, err
		}
		existingIndikatorIds = append(existingIndikatorIds, id)
	}

	// Buat map untuk indikator baru
	newIndikatorIds := make(map[string]bool)
	for _, ind := range pokinAdmin.Indikator {
		newIndikatorIds[ind.Id] = true
	}

	// Hapus indikator yang tidak ada dalam request
	for _, existingId := range existingIndikatorIds {
		if !newIndikatorIds[existingId] {
			// Hapus target terlebih dahulu
			scriptDeleteTargets := "DELETE FROM tb_target WHERE indikator_id = ?"
			_, err = tx.ExecContext(ctx, scriptDeleteTargets, existingId)
			if err != nil {
				return pokinAdmin, err
			}

			// Kemudian hapus indikator
			scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE id = ?"
			_, err = tx.ExecContext(ctx, scriptDeleteIndikator, existingId)
			if err != nil {
				return pokinAdmin, err
			}
		}
	}

	return pokinAdmin, nil
}

func (repository *PohonKinerjaRepositoryImpl) DeletePokinAdmin(ctx context.Context, tx *sql.Tx, id int) error {
	// Query untuk mendapatkan semua ID yang akan dihapus
	findIdsScript := `
        WITH RECURSIVE pohon_hierarki AS (
            -- Base case: node yang akan dihapus
            SELECT id, parent, level_pohon, clone_from 
            FROM tb_pohon_kinerja 
            WHERE id = ?
            
            UNION ALL
            
            -- Recursive case: child nodes dan data clone
            SELECT pk.id, pk.parent, pk.level_pohon, pk.clone_from
            FROM tb_pohon_kinerja pk
            INNER JOIN pohon_hierarki ph ON 
                -- Ambil child nodes langsung
                pk.parent = ph.id OR 
                -- Jika data asli, ambil yang mengclone-nya
                (ph.clone_from = 0 AND pk.clone_from = ph.id)
        ),
        clone_hierarki AS (
            -- Base case: data yang mengclone dan data yang parent-nya terhubung dengan id yang dihapus
            SELECT id, parent, level_pohon, clone_from
            FROM tb_pohon_kinerja
            WHERE clone_from IN (SELECT id FROM pohon_hierarki)
            OR parent IN (SELECT id FROM pohon_hierarki)
            
            UNION ALL
            
            -- Recursive case: child nodes dari data clone
            SELECT pk.id, pk.parent, pk.level_pohon, pk.clone_from
            FROM tb_pohon_kinerja pk
            INNER JOIN clone_hierarki ch ON 
                pk.parent = ch.id
        ),
        parent_hierarki AS (
            -- Ambil data yang parent-nya adalah id yang akan dihapus
            SELECT id, parent, level_pohon, clone_from
            FROM tb_pohon_kinerja
            WHERE parent = ?
        )
        SELECT id FROM pohon_hierarki
        UNION
        SELECT id FROM clone_hierarki
        UNION
        SELECT id FROM parent_hierarki;
    `

	rows, err := tx.QueryContext(ctx, findIdsScript, id, id)
	if err != nil {
		return fmt.Errorf("gagal mengambil hierarki ID: %v", err)
	}
	defer rows.Close()

	// Kumpulkan semua ID yang akan dihapus
	var idsToDelete []interface{}
	for rows.Next() {
		var idToDelete int
		if err := rows.Scan(&idToDelete); err != nil {
			return fmt.Errorf("gagal scan ID: %v", err)
		}
		idsToDelete = append(idsToDelete, idToDelete)
	}

	if len(idsToDelete) == 0 {
		return fmt.Errorf("tidak ada data yang akan dihapus")
	}

	// Buat placeholder untuk query IN clause
	placeholders := make([]string, len(idsToDelete))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	// Hapus target terlebih dahulu
	scriptDeleteTarget := fmt.Sprintf(`
        DELETE FROM tb_target 
        WHERE indikator_id IN (
            SELECT id FROM tb_indikator 
            WHERE pokin_id IN (%s)
        )`, inClause)
	_, err = tx.ExecContext(ctx, scriptDeleteTarget, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus target: %v", err)
	}

	// Hapus indikator
	scriptDeleteIndikator := fmt.Sprintf("DELETE FROM tb_indikator WHERE pokin_id IN (%s)", inClause)
	_, err = tx.ExecContext(ctx, scriptDeleteIndikator, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus indikator: %v", err)
	}

	// Hapus pelaksana
	scriptDeletePelaksana := fmt.Sprintf("DELETE FROM tb_pelaksana_pokin WHERE pohon_kinerja_id IN (%s)", inClause)
	_, err = tx.ExecContext(ctx, scriptDeletePelaksana, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus pelaksana: %v", err)
	}

	// Hapus pohon kinerja
	scriptDeletePokin := fmt.Sprintf("DELETE FROM tb_pohon_kinerja WHERE id IN (%s)", inClause)
	_, err = tx.ExecContext(ctx, scriptDeletePokin, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus pohon kinerja: %v", err)
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinAdminById(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error) {
	script := `
        SELECT 
            pk.id, 
            pk.parent, 
            pk.nama_pohon, 
            pk.jenis_pohon, 
            pk.level_pohon, 
            pk.kode_opd, 
            pk.keterangan, 
            pk.tahun,
            pk.status,
			pk.is_active
        FROM 
            tb_pohon_kinerja pk 
        WHERE 
            pk.id = ?`

	var pokin domain.PohonKinerja
	err := tx.QueryRowContext(ctx, script, id).Scan(
		&pokin.Id,
		&pokin.Parent,
		&pokin.NamaPohon,
		&pokin.JenisPohon,
		&pokin.LevelPohon,
		&pokin.KodeOpd,
		&pokin.Keterangan,
		&pokin.Tahun,
		&pokin.Status,
		&pokin.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.PohonKinerja{}, fmt.Errorf("pohon kinerja tidak ditemukan")
		}
		return domain.PohonKinerja{}, err
	}
	return pokin, nil
}
func (repository *PohonKinerjaRepositoryImpl) FindPokinAdminAll(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.PohonKinerja, error) {
	script := `
        SELECT 
            pk.id,
            pk.nama_pohon,
            pk.parent,
            pk.jenis_pohon,
            pk.level_pohon,
            pk.kode_opd,
            pk.keterangan,
            pk.tahun,
            i.id as indikator_id,
            i.indikator as nama_indikator,
            t.id as target_id,
            t.target as target_value,
            t.satuan as target_satuan
        FROM 
            tb_pohon_kinerja pk
        LEFT JOIN 
            tb_indikator i ON pk.id = i.pokin_id
        LEFT JOIN 
            tb_target t ON i.id = t.indikator_id
        WHERE 
            pk.tahun = ?
        ORDER BY 
            pk.level_pohon, pk.id, i.id, t.id
    `

	rows, err := tx.QueryContext(ctx, script, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map untuk menyimpan pohon kinerja yang sudah diproses
	pokinMap := make(map[int]domain.PohonKinerja)
	indikatorMap := make(map[string]domain.Indikator)

	for rows.Next() {
		var (
			pokinId, parent, levelPohon                            int
			namaPohon, jenisPohon, kodeOpd, keterangan, tahunPokin string
			indikatorId, namaIndikator                             sql.NullString
			targetId, targetValue, targetSatuan                    sql.NullString
		)

		err := rows.Scan(
			&pokinId, &namaPohon, &parent, &jenisPohon, &levelPohon,
			&kodeOpd, &keterangan, &tahunPokin,
			&indikatorId, &namaIndikator,
			&targetId, &targetValue, &targetSatuan,
		)
		if err != nil {
			return nil, err
		}

		// Proses Pohon Kinerja
		pokin, exists := pokinMap[pokinId]
		if !exists {
			pokin = domain.PohonKinerja{
				Id:         pokinId,
				NamaPohon:  namaPohon,
				Parent:     parent,
				JenisPohon: jenisPohon,
				LevelPohon: levelPohon,
				KodeOpd:    kodeOpd,
				Keterangan: keterangan,
				Tahun:      tahunPokin,
			}
			pokinMap[pokinId] = pokin
		}

		// Proses Indikator jika ada
		if indikatorId.Valid && namaIndikator.Valid {
			indikator, exists := indikatorMap[indikatorId.String]
			if !exists {
				indikator = domain.Indikator{
					Id:        indikatorId.String,
					PokinId:   fmt.Sprint(pokinId),
					Indikator: namaIndikator.String,
					Tahun:     tahunPokin,
				}
			}

			// Proses Target jika ada
			if targetId.Valid && targetValue.Valid && targetSatuan.Valid {
				target := domain.Target{
					Id:          targetId.String,
					IndikatorId: indikatorId.String,
					Target:      targetValue.String,
					Satuan:      targetSatuan.String,
					Tahun:       tahunPokin,
				}
				indikator.Target = append(indikator.Target, target)
			}

			indikatorMap[indikatorId.String] = indikator

			// Update indikator di pokin
			pokin.Indikator = append(pokin.Indikator, indikator)
			pokinMap[pokinId] = pokin
		}
	}

	// Konversi map ke slice
	var result []domain.PohonKinerja
	for _, pokin := range pokinMap {
		result = append(result, pokin)
	}

	// Urutkan berdasarkan level dan ID
	sort.Slice(result, func(i, j int) bool {
		if result[i].LevelPohon == result[j].LevelPohon {
			return result[i].Id < result[j].Id
		}
		return result[i].LevelPohon < result[j].LevelPohon
	})

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinAdminByIdHierarki(ctx context.Context, tx *sql.Tx, idPokin int) ([]domain.PohonKinerja, error) {
	script := `
        WITH RECURSIVE pohon_hierarki AS (
            -- Base case: pilih node yang diminta
            SELECT id, nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status, is_active
            FROM tb_pohon_kinerja 
            WHERE id = ?
            
            UNION ALL
            
            -- Recursive case: ambil semua child nodes
            SELECT pk.id, pk.nama_pohon, pk.parent, pk.jenis_pohon, pk.level_pohon, pk.kode_opd, pk.keterangan, pk.tahun, pk.status, pk.is_active
            FROM tb_pohon_kinerja pk
            INNER JOIN pohon_hierarki ph ON pk.parent = ph.id
        )
        SELECT 
            ph.id,
            ph.nama_pohon,
            ph.parent,
            ph.jenis_pohon,
            ph.level_pohon,
            ph.kode_opd,
            ph.keterangan,
            ph.tahun,
            ph.status,
			ph.is_active,
            i.id as indikator_id,
            i.indikator as nama_indikator,
            t.id as target_id,
            t.target as target_value,
            t.satuan as target_satuan,
            pp.id as pelaksana_id,
            pp.pegawai_id
        FROM 
            pohon_hierarki ph
        LEFT JOIN 
            tb_indikator i ON ph.id = i.pokin_id
        LEFT JOIN 
            tb_target t ON i.id = t.indikator_id
        LEFT JOIN 
            tb_pelaksana_pokin pp ON ph.id = pp.pohon_kinerja_id
        ORDER BY 
            ph.level_pohon, ph.id, i.id, t.id, pp.id
    `

	rows, err := tx.QueryContext(ctx, script, idPokin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map untuk menyimpan pohon kinerja yang sudah diproses
	pokinMap := make(map[int]domain.PohonKinerja)

	// Gunakan map untuk melacak indikator yang sudah diproses
	processedIndikators := make(map[string]bool)

	for rows.Next() {
		var (
			pokinId, parent, levelPohon                                    int
			namaPohon, jenisPohon, kodeOpd, keterangan, tahunPokin, status string
			is_active                                                      bool
			indikatorId, namaIndikator                                     sql.NullString
			targetId, targetValue, targetSatuan                            sql.NullString
			pelaksanaId, pegawaiId                                         sql.NullString
		)

		err := rows.Scan(
			&pokinId, &namaPohon, &parent, &jenisPohon, &levelPohon,
			&kodeOpd, &keterangan, &tahunPokin, &status, &is_active,
			&indikatorId, &namaIndikator,
			&targetId, &targetValue, &targetSatuan,
			&pelaksanaId, &pegawaiId,
		)
		if err != nil {
			return nil, err
		}

		// Proses Pohon Kinerja
		pokin, exists := pokinMap[pokinId]
		if !exists {
			pokin = domain.PohonKinerja{
				Id:         pokinId,
				NamaPohon:  namaPohon,
				Parent:     parent,
				JenisPohon: jenisPohon,
				LevelPohon: levelPohon,
				KodeOpd:    kodeOpd,
				Keterangan: keterangan,
				Tahun:      tahunPokin,
				Status:     status,
				IsActive:   is_active,
			}
		}

		// Proses Pelaksana jika ada
		if pelaksanaId.Valid && pegawaiId.Valid {
			pelaksana := domain.PelaksanaPokin{
				Id:        pelaksanaId.String,
				PegawaiId: pegawaiId.String,
			}
			// Cek duplikasi pelaksana
			isDuplicate := false
			for _, p := range pokin.Pelaksana {
				if p.Id == pelaksana.Id {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				pokin.Pelaksana = append(pokin.Pelaksana, pelaksana)
			}
		}

		// Proses Indikator jika ada
		if indikatorId.Valid && namaIndikator.Valid {
			// Cek apakah indikator sudah diproses
			if !processedIndikators[indikatorId.String] {
				processedIndikators[indikatorId.String] = true

				indikator := domain.Indikator{
					Id:        indikatorId.String,
					PokinId:   fmt.Sprint(pokinId),
					Indikator: namaIndikator.String,
					Tahun:     tahunPokin,
				}

				// Gunakan map untuk melacak target yang unik
				processedTargets := make(map[string]bool)

				// Proses Target jika ada
				if targetId.Valid && targetValue.Valid && targetSatuan.Valid {
					if !processedTargets[targetId.String] {
						processedTargets[targetId.String] = true
						target := domain.Target{
							Id:          targetId.String,
							IndikatorId: indikatorId.String,
							Target:      targetValue.String,
							Satuan:      targetSatuan.String,
							Tahun:       tahunPokin,
						}
						indikator.Target = append(indikator.Target, target)
					}
				}

				pokin.Indikator = append(pokin.Indikator, indikator)
			}
		}

		pokinMap[pokinId] = pokin
	}

	// Konversi map ke slice
	var result []domain.PohonKinerja
	for _, pokin := range pokinMap {
		result = append(result, pokin)
	}

	// Urutkan berdasarkan level dan ID
	sort.Slice(result, func(i, j int) bool {
		if result[i].LevelPohon == result[j].LevelPohon {
			return result[i].Id < result[j].Id
		}
		return result[i].LevelPohon < result[j].LevelPohon
	})

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindIndikatorByPokinId(ctx context.Context, tx *sql.Tx, pokinId string) ([]domain.Indikator, error) {
	script := `
        SELECT i.id, i.pokin_id, i.indikator, i.tahun, i.clone_from,
               t.id, t.indikator_id, t.target, t.satuan, t.tahun, t.clone_from
        FROM tb_indikator i
        LEFT JOIN tb_target t ON i.id = t.indikator_id
        WHERE i.pokin_id = ?`

	rows, err := tx.QueryContext(ctx, script, pokinId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indikatorMap := make(map[string]*domain.Indikator)

	for rows.Next() {
		var indId, pokinId, indikator, indTahun, indCloneFrom string
		var targetId, indikatorId, target, satuan, targetTahun sql.NullString
		var targetCloneFrom sql.NullString

		err := rows.Scan(
			&indId, &pokinId, &indikator, &indTahun, &indCloneFrom,
			&targetId, &indikatorId, &target, &satuan, &targetTahun, &targetCloneFrom)
		if err != nil {
			return nil, err
		}

		// Proses Indikator
		ind, exists := indikatorMap[indId]
		if !exists {
			ind = &domain.Indikator{
				Id:        indId,
				Indikator: indikator,
				Tahun:     indTahun,
				CloneFrom: indCloneFrom,
				Target:    []domain.Target{},
			}
			indikatorMap[indId] = ind
		}

		// Proses Target jika ada
		if targetId.Valid && indikatorId.Valid {
			target := domain.Target{
				Id:          targetId.String,
				IndikatorId: indikatorId.String,
				Target:      target.String,
				Satuan:      satuan.String,
				Tahun:       targetTahun.String,
			}
			if targetCloneFrom.Valid {
				target.CloneFrom = targetCloneFrom.String
			}
			ind.Target = append(ind.Target, target)
		}
	}

	// Convert map to slice
	var result []domain.Indikator
	for _, ind := range indikatorMap {
		result = append(result, *ind)
	}

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindTargetByIndikatorId(ctx context.Context, tx *sql.Tx, indikatorId string) ([]domain.Target, error) {
	script := "SELECT id, indikator_id, target, satuan, tahun FROM tb_target WHERE indikator_id = ?"
	rows, err := tx.QueryContext(ctx, script, indikatorId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []domain.Target
	for rows.Next() {
		var target domain.Target
		err := rows.Scan(&target.Id, &target.IndikatorId, &target.Target, &target.Satuan, &target.Tahun)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinToClone(ctx context.Context, tx *sql.Tx, id int) (domain.PohonKinerja, error) {
	script := "SELECT id, nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status, is_active FROM tb_pohon_kinerja WHERE id = ?"
	rows, err := tx.QueryContext(ctx, script, id)
	if err != nil {
		return domain.PohonKinerja{}, fmt.Errorf("gagal memeriksa data yang akan di-clone: %v", err)
	}
	defer rows.Close()

	var existingPokin domain.PohonKinerja
	if rows.Next() {
		err := rows.Scan(
			&existingPokin.Id,
			&existingPokin.NamaPohon,
			&existingPokin.Parent,
			&existingPokin.JenisPohon,
			&existingPokin.LevelPohon,
			&existingPokin.KodeOpd,
			&existingPokin.Keterangan,
			&existingPokin.Tahun,
			&existingPokin.Status,
			&existingPokin.IsActive,
		)
		if err != nil {
			return domain.PohonKinerja{}, fmt.Errorf("gagal membaca data yang akan di-clone: %v", err)
		}
		return existingPokin, nil
	}
	return domain.PohonKinerja{}, fmt.Errorf("data dengan ID %d tidak ditemukan", id)
}

func (repository *PohonKinerjaRepositoryImpl) ValidateParentLevel(ctx context.Context, tx *sql.Tx, parentId int, levelPohon int) error {
	// Validasi dasar: level tidak boleh kurang dari 4
	if levelPohon < 4 {
		return fmt.Errorf("level pohon tidak boleh kurang dari 4")
	}

	// Untuk level 4, parent bisa memiliki level 0 hingga 3
	if levelPohon == 4 {
		if parentId == 0 {
			return nil
		}
		// Cek level parentnya
		script := "SELECT level_pohon FROM tb_pohon_kinerja WHERE id = ?"
		var parentLevel int
		err := tx.QueryRowContext(ctx, script, parentId).Scan(&parentLevel)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("parent dengan ID %d tidak ditemukan", parentId)
			}
			return fmt.Errorf("gagal memeriksa level parent: %v", err)
		}

		// Validasi level parent untuk level 4
		if parentLevel < 0 || parentLevel > 3 {
			return fmt.Errorf("level pohon 4 harus memiliki parent dengan level 0 hingga 3, bukan level %d", parentLevel)
		}
		return nil
	}

	// Untuk level > 4, parent tidak boleh 0
	if parentId == 0 {
		return fmt.Errorf("level pohon %d harus memiliki parent", levelPohon)
	}

	// Cek level parent untuk level > 4
	script := "SELECT level_pohon FROM tb_pohon_kinerja WHERE id = ?"
	var parentLevel int
	err := tx.QueryRowContext(ctx, script, parentId).Scan(&parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parent dengan ID %d tidak ditemukan", parentId)
		}
		return fmt.Errorf("gagal memeriksa level parent: %v", err)
	}

	// Validasi: level parent harus tepat 1 tingkat di atas level saat ini
	expectedParentLevel := levelPohon - 1
	if parentLevel != expectedParentLevel {
		return fmt.Errorf("level pohon %d harus memiliki parent dengan level %d, bukan level %d",
			levelPohon, expectedParentLevel, parentLevel)
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) FindIndikatorToClone(ctx context.Context, tx *sql.Tx, pokinId int) ([]domain.Indikator, error) {
	script := "SELECT id, indikator, tahun FROM tb_indikator WHERE pokin_id = ?"
	rows, err := tx.QueryContext(ctx, script, pokinId)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data indikator: %v", err)
	}
	defer rows.Close()

	var indikators []domain.Indikator
	for rows.Next() {
		var indikator domain.Indikator
		err := rows.Scan(&indikator.Id, &indikator.Indikator, &indikator.Tahun)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca data indikator: %v", err)
		}
		indikators = append(indikators, indikator)
	}
	return indikators, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindTargetToClone(ctx context.Context, tx *sql.Tx, indikatorId string) ([]domain.Target, error) {
	script := "SELECT id, target, satuan, tahun FROM tb_target WHERE indikator_id = ?"
	rows, err := tx.QueryContext(ctx, script, indikatorId)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data target: %v", err)
	}
	defer rows.Close()

	var targets []domain.Target
	for rows.Next() {
		var target domain.Target
		err := rows.Scan(&target.Id, &target.Target, &target.Satuan, &target.Tahun)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca data target: %v", err)
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func (repository *PohonKinerjaRepositoryImpl) InsertClonedPokin(ctx context.Context, tx *sql.Tx, pokin domain.PohonKinerja) (int64, error) {
	script := `INSERT INTO tb_pohon_kinerja 
        (nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status, clone_from, is_active) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, script,
		pokin.NamaPohon,
		pokin.Parent,
		pokin.JenisPohon,
		pokin.LevelPohon,
		pokin.KodeOpd,
		pokin.Keterangan,
		pokin.Tahun,
		pokin.Status,
		pokin.CloneFrom,
		pokin.IsActive,
	)
	if err != nil {
		return 0, fmt.Errorf("gagal menyimpan data pohon kinerja yang di-clone: %v", err)
	}
	return result.LastInsertId()
}

func (repository *PohonKinerjaRepositoryImpl) InsertClonedIndikator(ctx context.Context, tx *sql.Tx, indikatorId string, pokinId int64, indikator domain.Indikator) error {
	script := "INSERT INTO tb_indikator (id, pokin_id, indikator, tahun, clone_from) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, script,
		indikatorId,
		pokinId,
		indikator.Indikator,
		indikator.Tahun,
		indikator.Id, // Id indikator asli sebagai clone_from
	)
	if err != nil {
		return fmt.Errorf("gagal menyimpan indikator baru: %v", err)
	}
	return nil
}

func (repository *PohonKinerjaRepositoryImpl) InsertClonedTarget(ctx context.Context, tx *sql.Tx, targetId string, indikatorId string, target domain.Target) error {
	fmt.Printf("Inserting target with clone_from: %s\n", target.Id) // Log sementara
	script := "INSERT INTO tb_target (id, indikator_id, target, satuan, tahun, clone_from) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, script,
		targetId,
		indikatorId,
		target.Target,
		target.Satuan,
		target.Tahun,
		target.Id, // Id target asli sebagai clone_from
	)
	if err != nil {
		return fmt.Errorf("gagal menyimpan target baru: %v", err)
	}
	return nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinByJenisPohon(ctx context.Context, tx *sql.Tx, jenisPohon string, levelPohon int, tahun string, kodeOpd string, status string) ([]domain.PohonKinerja, error) {
	script := "SELECT id, nama_pohon, jenis_pohon, level_pohon, kode_opd, tahun, keterangan, status, is_active FROM tb_pohon_kinerja WHERE 1=1"
	parameters := []interface{}{}
	if jenisPohon != "" {
		script += " AND jenis_pohon = ?"
		parameters = append(parameters, jenisPohon)
	}
	if levelPohon != 0 {
		script += " AND level_pohon = ?"
		parameters = append(parameters, levelPohon)
	}
	if kodeOpd != "" {
		script += " AND kode_opd = ?"
		parameters = append(parameters, kodeOpd)
	}
	if tahun != "" {
		script += " AND tahun = ?"
		parameters = append(parameters, tahun)
	}
	if status != "" {
		script += " AND status = ?"
		parameters = append(parameters, status)
	}
	script += " ORDER BY nama_pohon asc"

	rows, err := tx.QueryContext(ctx, script, parameters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokins []domain.PohonKinerja
	for rows.Next() {
		var pokin domain.PohonKinerja
		err := rows.Scan(&pokin.Id, &pokin.NamaPohon, &pokin.JenisPohon, &pokin.LevelPohon, &pokin.KodeOpd, &pokin.Tahun, &pokin.Keterangan, &pokin.Status, &pokin.IsActive)
		if err != nil {
			return nil, err
		}
		pokins = append(pokins, pokin)
	}
	return pokins, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinByPelaksana(ctx context.Context, tx *sql.Tx, pegawaiId string, tahun string) ([]domain.PohonKinerja, error) {
	script := `
        SELECT DISTINCT
            pk.id,
            pk.nama_pohon,
            pk.parent,
            pk.jenis_pohon,
            pk.level_pohon,
            pk.kode_opd,
            pk.keterangan,
            pk.tahun,
            pk.created_at,
            pp.id as pelaksana_id,
            pp.pegawai_id
        FROM 
            tb_pohon_kinerja pk
        INNER JOIN 
            tb_pelaksana_pokin pp ON pk.id = pp.pohon_kinerja_id
        WHERE 
            pp.pegawai_id = ?
            AND pk.tahun = ?
        ORDER BY 
            pk.level_pohon, pk.id, pk.created_at ASC
    `

	rows, err := tx.QueryContext(ctx, script, pegawaiId, tahun)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data pohon kinerja: %v", err)
	}
	defer rows.Close()

	pokinMap := make(map[int]domain.PohonKinerja)

	for rows.Next() {
		var pokin domain.PohonKinerja
		var pelaksana domain.PelaksanaPokin

		err := rows.Scan(
			&pokin.Id,
			&pokin.NamaPohon,
			&pokin.Parent,
			&pokin.JenisPohon,
			&pokin.LevelPohon,
			&pokin.KodeOpd,
			&pokin.Keterangan,
			&pokin.Tahun,
			&pokin.CreatedAt,
			&pelaksana.Id,
			&pelaksana.PegawaiId,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan data pohon kinerja: %v", err)
		}

		// Cek apakah pohon kinerja sudah ada di map
		existingPokin, exists := pokinMap[pokin.Id]
		if exists {
			// Jika sudah ada, tambahkan pelaksana baru ke slice pelaksana yang ada
			existingPokin.Pelaksana = append(existingPokin.Pelaksana, pelaksana)
			pokinMap[pokin.Id] = existingPokin
		} else {
			// Jika belum ada, buat entry baru dengan pelaksana pertama
			pokin.Pelaksana = []domain.PelaksanaPokin{pelaksana}
			pokinMap[pokin.Id] = pokin
		}
	}

	// Konversi map ke slice untuk hasil akhir
	var result []domain.PohonKinerja
	for _, pokin := range pokinMap {
		result = append(result, pokin)
	}

	if len(result) == 0 {
		return nil, sql.ErrNoRows
	}

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinByStatus(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string, status string) ([]domain.PohonKinerja, error) {
	SQL := `SELECT id, nama_pohon, kode_opd, tahun, jenis_pohon, level_pohon, parent, status 
            FROM tb_pohon_kinerja 
            WHERE kode_opd = ? AND tahun = ? AND status = ?`

	rows, err := tx.QueryContext(ctx, SQL, kodeOpd, tahun, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokins []domain.PohonKinerja
	for rows.Next() {
		pokin := domain.PohonKinerja{}
		err := rows.Scan(
			&pokin.Id,
			&pokin.NamaPohon,
			&pokin.KodeOpd,
			&pokin.Tahun,
			&pokin.JenisPohon,
			&pokin.LevelPohon,
			&pokin.Parent,
			&pokin.Status,
		)
		if err != nil {
			return nil, err
		}
		pokins = append(pokins, pokin)
	}
	return pokins, nil
}

func (repository *PohonKinerjaRepositoryImpl) UpdatePokinStatus(ctx context.Context, tx *sql.Tx, id int, status string) error {
	script := "UPDATE tb_pohon_kinerja SET status = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, script, status, id)
	if err != nil {
		return fmt.Errorf("gagal mengupdate status: %v", err)
	}
	return nil
}

func (repository *PohonKinerjaRepositoryImpl) CheckPokinStatus(ctx context.Context, tx *sql.Tx, id int) (string, error) {
	script := "SELECT status FROM tb_pohon_kinerja WHERE id = ?"
	var status string
	err := tx.QueryRowContext(ctx, script, id).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("gagal mengecek status: %v", err)
	}
	return status, nil
}

func (repository *PohonKinerjaRepositoryImpl) InsertClonedPokinWithStatus(ctx context.Context, tx *sql.Tx, pokin domain.PohonKinerja) (int64, error) {
	script := `INSERT INTO tb_pohon_kinerja 
        (nama_pohon, parent, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status, clone_from) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, script,
		pokin.NamaPohon,
		pokin.Parent,
		pokin.JenisPohon,
		pokin.LevelPohon,
		pokin.KodeOpd,
		pokin.Keterangan,
		pokin.Tahun,
		pokin.Status,
		pokin.CloneFrom,
	)
	if err != nil {
		return 0, fmt.Errorf("gagal menyimpan data pohon kinerja yang di-clone: %v", err)
	}
	return result.LastInsertId()
}

func (repository *PohonKinerjaRepositoryImpl) UpdatePokinStatusTolak(ctx context.Context, tx *sql.Tx, id int, status string) error {
	script := "UPDATE tb_pohon_kinerja SET status = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, script, status, id)
	if err != nil {
		return fmt.Errorf("gagal mengupdate status dan alasan: %v", err)
	}
	return nil
}

func (repository *PohonKinerjaRepositoryImpl) CheckCloneFrom(ctx context.Context, tx *sql.Tx, id int) (int, error) {
	script := "SELECT COALESCE(clone_from, 0) FROM tb_pohon_kinerja WHERE id = ?"
	var cloneFrom int
	err := tx.QueryRowContext(ctx, script, id).Scan(&cloneFrom)
	if err != nil {
		return 0, fmt.Errorf("gagal mengecek clone_from: %v", err)
	}
	return cloneFrom, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinByCloneFrom(ctx context.Context, tx *sql.Tx, cloneFromId int) ([]domain.PohonKinerja, error) {
	script := "SELECT id, parent, nama_pohon, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status, clone_from FROM tb_pohon_kinerja WHERE clone_from = ?"
	rows, err := tx.QueryContext(ctx, script, cloneFromId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokins []domain.PohonKinerja
	for rows.Next() {
		var pokin domain.PohonKinerja
		err := rows.Scan(
			&pokin.Id,
			&pokin.Parent,
			&pokin.NamaPohon,
			&pokin.JenisPohon,
			&pokin.LevelPohon,
			&pokin.KodeOpd,
			&pokin.Keterangan,
			&pokin.Tahun,
			&pokin.Status,
			&pokin.CloneFrom,
		)
		if err != nil {
			return nil, err
		}
		pokins = append(pokins, pokin)
	}
	return pokins, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindIndikatorByCloneFrom(ctx context.Context, tx *sql.Tx, pokinId int, cloneFromId string) (domain.Indikator, error) {
	script := "SELECT id, indikator, tahun FROM tb_indikator WHERE pokin_id = ? AND clone_from = ?"
	var indikator domain.Indikator
	err := tx.QueryRowContext(ctx, script, pokinId, cloneFromId).Scan(
		&indikator.Id,
		&indikator.Indikator,
		&indikator.Tahun,
	)
	if err != nil {
		return domain.Indikator{}, err
	}
	return indikator, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindTargetByCloneFrom(ctx context.Context, tx *sql.Tx, indikatorId string, cloneFromId string) (domain.Target, error) {
	script := "SELECT id, target, satuan, tahun FROM tb_target WHERE indikator_id = ? AND clone_from = ?"
	var target domain.Target
	err := tx.QueryRowContext(ctx, script, indikatorId, cloneFromId).Scan(
		&target.Id,
		&target.Target,
		&target.Satuan,
		&target.Tahun,
	)
	if err != nil {
		return domain.Target{}, err
	}
	return target, nil
}

// Tambahkan method baru untuk FindPokinByCrosscuttingStatus
func (repository *PohonKinerjaRepositoryImpl) FindPokinByCrosscuttingStatus(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string) ([]domain.PohonKinerja, error) {
	script := `SELECT 
        id, nama_pohon, parent, jenis_pohon, level_pohon, 
        kode_opd, keterangan, tahun, status 
        FROM tb_pohon_kinerja 
        WHERE kode_opd = ? 
        AND tahun = ? 
        AND status IN ('crosscutting_menunggu','crosscutting_ditolak')
        ORDER BY level_pohon, id ASC`

	rows, err := tx.QueryContext(ctx, script, kodeOpd, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokins []domain.PohonKinerja
	for rows.Next() {
		var pokin domain.PohonKinerja
		err := rows.Scan(
			&pokin.Id, &pokin.NamaPohon, &pokin.Parent, &pokin.JenisPohon, &pokin.LevelPohon, &pokin.KodeOpd, &pokin.Keterangan, &pokin.Tahun, &pokin.Status,
		)
		if err != nil {
			return nil, err
		}
		pokins = append(pokins, pokin)
	}
	return pokins, nil
}

func (repository *PohonKinerjaRepositoryImpl) DeleteClonedPokinHierarchy(ctx context.Context, tx *sql.Tx, id int) error {
	// Query untuk mendapatkan hierarki dari data clone
	findIdsScript := `
        WITH RECURSIVE clone_hierarki AS (
            -- Base case: node clone yang akan dihapus
            SELECT id, parent, level_pohon, clone_from 
            FROM tb_pohon_kinerja 
            WHERE id = ?
            
            UNION ALL
            
            -- Recursive case: child nodes dari data clone
            SELECT pk.id, pk.parent, pk.level_pohon, pk.clone_from
            FROM tb_pohon_kinerja pk
            INNER JOIN clone_hierarki ch ON 
                pk.parent = ch.id
        )
        SELECT id FROM clone_hierarki;
    `

	rows, err := tx.QueryContext(ctx, findIdsScript, id)
	if err != nil {
		return fmt.Errorf("gagal mengambil hierarki clone: %v", err)
	}
	defer rows.Close()

	var idsToDelete []interface{}
	for rows.Next() {
		var idToDelete int
		if err := rows.Scan(&idToDelete); err != nil {
			return fmt.Errorf("gagal scan ID: %v", err)
		}
		idsToDelete = append(idsToDelete, idToDelete)
	}

	if len(idsToDelete) == 0 {
		return fmt.Errorf("tidak ada data yang akan dihapus")
	}

	// Buat placeholder untuk query IN clause
	placeholders := make([]string, len(idsToDelete))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	// Hapus target terlebih dahulu
	scriptDeleteTarget := fmt.Sprintf(`
        DELETE FROM tb_target 
        WHERE indikator_id IN (
            SELECT id FROM tb_indikator 
            WHERE pokin_id IN (%s)
        )`, inClause)
	_, err = tx.ExecContext(ctx, scriptDeleteTarget, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus target: %v", err)
	}

	// Hapus indikator
	scriptDeleteIndikator := fmt.Sprintf("DELETE FROM tb_indikator WHERE pokin_id IN (%s)", inClause)
	_, err = tx.ExecContext(ctx, scriptDeleteIndikator, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus indikator: %v", err)
	}

	// Hapus pelaksana
	scriptDeletePelaksana := fmt.Sprintf("DELETE FROM tb_pelaksana_pokin WHERE pohon_kinerja_id IN (%s)", inClause)
	_, err = tx.ExecContext(ctx, scriptDeletePelaksana, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus pelaksana: %v", err)
	}

	// Hapus pohon kinerja
	scriptDeletePokin := fmt.Sprintf("DELETE FROM tb_pohon_kinerja WHERE id IN (%s)", inClause)
	_, err = tx.ExecContext(ctx, scriptDeletePokin, idsToDelete...)
	if err != nil {
		return fmt.Errorf("gagal menghapus pohon kinerja: %v", err)
	}

	return nil
}

func (r *PohonKinerjaRepositoryImpl) FindChildPokins(ctx context.Context, tx *sql.Tx, parentId int64) ([]domain.PohonKinerja, error) {
	SQL := `SELECT id, parent, nama_pohon, jenis_pohon, level_pohon, kode_opd, keterangan, tahun, status, clone_from, is_active
            FROM tb_pohon_kinerja 
            WHERE parent = ?`

	rows, err := tx.QueryContext(ctx, SQL, parentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pokins []domain.PohonKinerja
	for rows.Next() {
		var pokin domain.PohonKinerja
		err := rows.Scan(
			&pokin.Id,
			&pokin.Parent,
			&pokin.NamaPohon,
			&pokin.JenisPohon,
			&pokin.LevelPohon,
			&pokin.KodeOpd,
			&pokin.Keterangan,
			&pokin.Tahun,
			&pokin.Status,
			&pokin.CloneFrom,
			&pokin.IsActive,
		)
		if err != nil {
			return nil, err
		}
		pokins = append(pokins, pokin)
	}
	return pokins, nil
}

func (repository *PohonKinerjaRepositoryImpl) InsertClonedPelaksana(ctx context.Context, tx *sql.Tx, newId string, pokinId int64, pelaksana domain.PelaksanaPokin) error {
	SQL := `INSERT INTO tb_pelaksana_pokin (id, pokin_id, pegawai_id)
            VALUES (?, ?, ?)`

	_, err := tx.ExecContext(ctx, SQL, newId, pokinId, pelaksana.PegawaiId)
	return err
}

func (repository *PohonKinerjaRepositoryImpl) UpdatePokinStatusFromApproved(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `
        UPDATE tb_pohon_kinerja 
        SET status = 'menunggu_disetujui' 
        WHERE id = ? 
        AND status = 'disetujui'
    `

	result, err := tx.ExecContext(ctx, SQL, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("tidak ada data yang diupdate untuk ID %d", id)
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) UpdateParent(ctx context.Context, tx *sql.Tx, pohonKinerja domain.PohonKinerja) (domain.PohonKinerja, error) {
	script := `UPDATE tb_pohon_kinerja SET parent = ? WHERE id = ?`
	_, err := tx.ExecContext(ctx, script, pohonKinerja.Parent, pohonKinerja.Id)
	if err != nil {
		return domain.PohonKinerja{}, err
	}
	return pohonKinerja, nil
}

func (repository *PohonKinerjaRepositoryImpl) FindidPokinWithAllTema(ctx context.Context, tx *sql.Tx, id int) ([]domain.PohonKinerja, error) {
	script := `
                 WITH RECURSIVE ancestor_tree AS (
            -- Base case: node yang dicari
            SELECT 
                pk.id, pk.nama_pohon, pk.parent, pk.jenis_pohon, 
                pk.level_pohon, pk.kode_opd, pk.keterangan, 
                pk.tahun, pk.status,
                i.id as indikator_id, i.indikator as nama_indikator,
                t.id as target_id, t.target as target_value, 
                t.satuan as target_satuan,
                pp.id as pelaksana_id, pp.pegawai_id
            FROM tb_pohon_kinerja pk
            LEFT JOIN tb_indikator i ON pk.id = i.pokin_id
            LEFT JOIN tb_target t ON i.id = t.indikator_id
            LEFT JOIN tb_pelaksana_pokin pp ON pk.id = pp.pohon_kinerja_id
            WHERE pk.id = ?
            
            UNION ALL
            
            -- Recursive case: ambil parent nodes
            SELECT 
                pk.id, pk.nama_pohon, pk.parent, pk.jenis_pohon, 
                pk.level_pohon, pk.kode_opd, pk.keterangan, 
                pk.tahun, pk.status,
                i.id as indikator_id, i.indikator as nama_indikator,
                t.id as target_id, t.target as target_value, 
                t.satuan as target_satuan,
                pp.id as pelaksana_id, pp.pegawai_id
            FROM tb_pohon_kinerja pk
            LEFT JOIN tb_indikator i ON pk.id = i.pokin_id
            LEFT JOIN tb_target t ON i.id = t.indikator_id
            LEFT JOIN tb_pelaksana_pokin pp ON pk.id = pp.pohon_kinerja_id
            INNER JOIN ancestor_tree at ON pk.id = at.parent
        )
        SELECT * FROM ancestor_tree
        ORDER BY level_pohon ASC`

	rows, err := tx.QueryContext(ctx, script, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pokinMap := make(map[int]*domain.PohonKinerja)
	processedIndikators := make(map[string]bool)

	for rows.Next() {
		var (
			pokinId                                    int
			namaPohon, jenisPohon, kodeOpd, keterangan string
			parent, levelPohon                         int
			tahun, status                              string
			indikatorId, namaIndikator                 sql.NullString
			targetId, targetValue, targetSatuan        sql.NullString
			pelaksanaId, pegawaiId                     sql.NullString
		)

		err := rows.Scan(
			&pokinId, &namaPohon, &parent, &jenisPohon,
			&levelPohon, &kodeOpd, &keterangan,
			&tahun, &status,
			&indikatorId, &namaIndikator,
			&targetId, &targetValue, &targetSatuan,
			&pelaksanaId, &pegawaiId,
		)
		if err != nil {
			return nil, err
		}

		// Inisialisasi atau ambil pohon kinerja dari map
		pokin, exists := pokinMap[pokinId]
		if !exists {
			pokin = &domain.PohonKinerja{
				Id:         pokinId,
				NamaPohon:  namaPohon,
				Parent:     parent,
				JenisPohon: jenisPohon,
				LevelPohon: levelPohon,
				KodeOpd:    kodeOpd,
				Keterangan: keterangan,
				Tahun:      tahun,
				Status:     status,
			}
			pokinMap[pokinId] = pokin
		}

		// Proses Indikator dan Target
		if indikatorId.Valid && namaIndikator.Valid {
			if !processedIndikators[indikatorId.String] {
				processedIndikators[indikatorId.String] = true
				indikator := domain.Indikator{
					Id:        indikatorId.String,
					PokinId:   fmt.Sprint(pokinId),
					Indikator: namaIndikator.String,
					Tahun:     tahun,
				}

				if targetId.Valid && targetValue.Valid && targetSatuan.Valid {
					target := domain.Target{
						Id:          targetId.String,
						IndikatorId: indikatorId.String,
						Target:      targetValue.String,
						Satuan:      targetSatuan.String,
						Tahun:       tahun,
					}
					indikator.Target = append(indikator.Target, target)
				}

				pokin.Indikator = append(pokin.Indikator, indikator)
			}
		}

		// Proses Pelaksana
		if pelaksanaId.Valid && pegawaiId.Valid {
			pelaksana := domain.PelaksanaPokin{
				Id:        pelaksanaId.String,
				PegawaiId: pegawaiId.String,
			}
			// Cek duplikasi
			isDuplicate := false
			for _, p := range pokin.Pelaksana {
				if p.Id == pelaksana.Id {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				pokin.Pelaksana = append(pokin.Pelaksana, pelaksana)
			}
		}
	}

	// Convert map to sorted slice
	var result []domain.PohonKinerja
	for _, pokin := range pokinMap {
		result = append(result, *pokin)
	}

	// Sort berdasarkan level_pohon
	sort.Slice(result, func(i, j int) bool {
		return result[i].LevelPohon < result[j].LevelPohon
	})

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) ValidatePokinId(ctx context.Context, tx *sql.Tx, pokinId int) error {
	script := "SELECT COUNT(*) FROM tb_pohon_kinerja WHERE id = ?"

	var count int
	err := tx.QueryRowContext(ctx, script, pokinId).Scan(&count)
	if err != nil {
		return fmt.Errorf("gagal melakukan validasi pohon kinerja: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("pohon kinerja dengan ID %d tidak ditemukan", pokinId)
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) ValidatePokinLevel(ctx context.Context, tx *sql.Tx, pokinId int, expectedLevel int, purpose string) error {
	script := "SELECT level_pohon FROM tb_pohon_kinerja WHERE id = ?"

	var levelPohon int
	err := tx.QueryRowContext(ctx, script, pokinId).Scan(&levelPohon)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("pohon kinerja dengan ID %d tidak ditemukan", pokinId)
		}
		return fmt.Errorf("gagal melakukan validasi pohon kinerja: %v", err)
	}

	if levelPohon != expectedLevel {
		return fmt.Errorf("%s hanya bisa dibuat dari pohon kinerja level %d, bukan level %d",
			purpose, expectedLevel, levelPohon)
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) FindPokinWithPeriode(ctx context.Context, tx *sql.Tx, pokinId int, jenisPeriode string) (domain.PohonKinerja, domain.Periode, error) {
	query := `
        SELECT 
            pk.id,
            pk.nama_pohon,
            pk.parent,
            pk.jenis_pohon,
            pk.level_pohon,
            pk.kode_opd,
            pk.keterangan,
            pk.tahun,
            pk.status,
            COALESCE(p.id, 0) as periode_id,
            COALESCE(p.tahun_awal, '') as tahun_awal,
            COALESCE(p.tahun_akhir, '') as tahun_akhir,
            i.id as indikator_id,
            i.indikator as indikator_text,
            i.rumus_perhitungan,
            i.sumber_data,
            t.id as target_id,
            t.target as target_value,
            t.satuan as target_satuan
        FROM 
            tb_pohon_kinerja pk
        LEFT JOIN 
            tb_periode p ON CAST(pk.tahun AS SIGNED) BETWEEN CAST(p.tahun_awal AS SIGNED) AND CAST(p.tahun_akhir AS SIGNED)
            AND p.jenis_periode = ?
        LEFT JOIN 
            tb_indikator i ON pk.id = i.pokin_id
        LEFT JOIN 
            tb_target t ON i.id = t.indikator_id
        WHERE 
            pk.id = ?
        ORDER BY 
            i.id`

	rows, err := tx.QueryContext(ctx, query, jenisPeriode, pokinId)
	if err != nil {
		return domain.PohonKinerja{}, domain.Periode{}, fmt.Errorf("error querying data: %v", err)
	}
	defer rows.Close()

	var pokin domain.PohonKinerja
	var periode domain.Periode
	indikatorMap := make(map[string]*domain.Indikator)
	firstRow := true

	for rows.Next() {
		var (
			periodeId                           int
			tahunAwal, tahunAkhir               string
			indikatorId, indikatorText          sql.NullString
			rumusPerhitungan, sumberData        sql.NullString
			targetId, targetValue, targetSatuan sql.NullString
		)

		err := rows.Scan(
			&pokin.Id,
			&pokin.NamaPohon,
			&pokin.Parent,
			&pokin.JenisPohon,
			&pokin.LevelPohon,
			&pokin.KodeOpd,
			&pokin.Keterangan,
			&pokin.Tahun,
			&pokin.Status,
			&periodeId,
			&tahunAwal,
			&tahunAkhir,
			&indikatorId,
			&indikatorText,
			&rumusPerhitungan,
			&sumberData,
			&targetId,
			&targetValue,
			&targetSatuan,
		)
		if err != nil {
			return domain.PohonKinerja{}, domain.Periode{}, fmt.Errorf("error scanning row: %v", err)
		}

		if firstRow {
			periode = domain.Periode{
				Id:         periodeId,
				TahunAwal:  tahunAwal,
				TahunAkhir: tahunAkhir,
			}
			firstRow = false
		}

		// Proses indikator jika ada
		if indikatorId.Valid && indikatorText.Valid {
			indikator, exists := indikatorMap[indikatorId.String]
			if !exists {
				indikator = &domain.Indikator{
					Id:               indikatorId.String,
					PokinId:          fmt.Sprint(pokin.Id),
					Indikator:        indikatorText.String,
					RumusPerhitungan: rumusPerhitungan,
					SumberData:       sumberData,
					Target:           []domain.Target{},
				}
				indikatorMap[indikatorId.String] = indikator

				// Buat target untuk setiap tahun dalam periode
				if periode.Id != 0 && periode.TahunAwal != "" && periode.TahunAkhir != "" {
					tahunAwalInt, _ := strconv.Atoi(periode.TahunAwal)
					tahunAkhirInt, _ := strconv.Atoi(periode.TahunAkhir)

					for tahun := tahunAwalInt; tahun <= tahunAkhirInt; tahun++ {
						tahunStr := strconv.Itoa(tahun)
						target := domain.Target{
							Id:          "-",
							IndikatorId: indikatorId.String,
							Target:      "",
							Satuan:      "",
							Tahun:       tahunStr,
						}
						indikator.Target = append(indikator.Target, target)
					}
				}
			}

			// Update target jika ada data
			if targetId.Valid && targetValue.Valid && targetSatuan.Valid {
				for i := range indikator.Target {
					if indikator.Target[i].Tahun == pokin.Tahun {
						indikator.Target[i] = domain.Target{
							Id:          targetId.String,
							IndikatorId: indikatorId.String,
							Target:      targetValue.String,
							Satuan:      targetSatuan.String,
							Tahun:       pokin.Tahun,
						}
						break
					}
				}
			}
		}
	}

	// Konversi map indikator ke slice
	for _, indikator := range indikatorMap {
		// Sort target berdasarkan tahun
		sort.Slice(indikator.Target, func(i, j int) bool {
			tahunI, _ := strconv.Atoi(indikator.Target[i].Tahun)
			tahunJ, _ := strconv.Atoi(indikator.Target[j].Tahun)
			return tahunI < tahunJ
		})
		pokin.Indikator = append(pokin.Indikator, *indikator)
	}

	// Sort indikator berdasarkan ID
	sort.Slice(pokin.Indikator, func(i, j int) bool {
		return pokin.Indikator[i].Id < pokin.Indikator[j].Id
	})

	if pokin.Id == 0 {
		return pokin, periode, fmt.Errorf("pohon kinerja tidak ditemukan")
	}

	return pokin, periode, nil
}

// aktif / nonaktif tematik
func (repository *PohonKinerjaRepositoryImpl) UpdateTematikStatus(ctx context.Context, tx *sql.Tx, id int, isActive bool) error {
	query := "UPDATE tb_pohon_kinerja SET is_active = ? WHERE id = ?"
	_, err := tx.ExecContext(ctx, query, isActive, id)
	return err
}

func (repository *PohonKinerjaRepositoryImpl) GetChildrenAndClones(ctx context.Context, tx *sql.Tx, parentId int, isActivating bool) ([]int, error) {
	var query string

	if isActivating {
		// Query untuk mengaktifkan: ambil semua yang terhubung tanpa memandang status is_active
		query = `
            WITH RECURSIVE tree AS (
                -- Base case: direct children and clones yang nonaktif
                SELECT id, parent, clone_from, level_pohon
                FROM tb_pohon_kinerja
                WHERE (parent = ? OR clone_from = ?) 
                AND is_active = false
                
                UNION ALL
                
                -- Recursive case: children dan clone yang nonaktif
                SELECT pk.id, pk.parent, pk.clone_from, pk.level_pohon
                FROM tb_pohon_kinerja pk
                INNER JOIN tree t ON (pk.parent = t.id OR pk.clone_from = t.id)
                WHERE pk.is_active = false
            )
            SELECT DISTINCT id FROM tree`
	} else {
		// Query untuk menonaktifkan: ambil semua yang terhubung dan masih aktif
		query = `
            WITH RECURSIVE tree AS (
                -- Base case: direct children and clones yang aktif
                SELECT id, parent, clone_from, level_pohon
                FROM tb_pohon_kinerja
                WHERE (parent = ? OR clone_from = ?)
                AND is_active = true
                
                UNION ALL
                
                -- Recursive case: children dan clone yang aktif
                SELECT pk.id, pk.parent, pk.clone_from, pk.level_pohon
                FROM tb_pohon_kinerja pk
                INNER JOIN tree t ON (pk.parent = t.id OR pk.clone_from = t.id)
                WHERE pk.is_active = true
            )
            SELECT DISTINCT id FROM tree`
	}

	rows, err := tx.QueryContext(ctx, query, parentId, parentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// clone pokin opd
func (repository *PohonKinerjaRepositoryImpl) IsExistsByTahun(ctx context.Context, tx *sql.Tx, kodeOpd string, tahun string) bool {
	script := "SELECT COUNT(*) FROM tb_pohon_kinerja WHERE kode_opd = ? AND tahun = ?"
	var count int
	err := tx.QueryRowContext(ctx, script, kodeOpd, tahun).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (repository *PohonKinerjaRepositoryImpl) ClonePokinOpd(ctx context.Context, tx *sql.Tx, kodeOpd string, sourceTahun string, targetTahun string) error {
	// 1. Dapatkan daftar ID yang valid (status kosong dan parent dengan status kosong)
	scriptValidIds := `
	  SELECT p1.id 
	  FROM tb_pohon_kinerja p1
	  LEFT JOIN tb_pohon_kinerja p2 ON p1.parent = p2.id 
	  WHERE p1.kode_opd = ? 
	  AND p1.tahun = ?
	  AND p1.status = ''
	  AND (p1.parent = 0 OR (p2.status = '' OR p2.status IS NULL))
  `
	validRows, err := tx.QueryContext(ctx, scriptValidIds, kodeOpd, sourceTahun)
	if err != nil {
		return err
	}
	defer validRows.Close()

	var validIds []int
	for validRows.Next() {
		var id int
		if err := validRows.Scan(&id); err != nil {
			return err
		}
		validIds = append(validIds, id)
	}

	// 2. Clone Pohon Kinerja yang valid
	for _, validId := range validIds {
		scriptPokin := `
		  INSERT INTO tb_pohon_kinerja (
			  nama_pohon, parent, jenis_pohon, level_pohon, 
			  kode_opd, keterangan, keterangan_crosscutting, 
			  tahun, status, is_active
		  )
		  SELECT 
			  nama_pohon, parent, jenis_pohon, level_pohon,
			  kode_opd, keterangan, keterangan_crosscutting,
			  ?, '', is_active
		  FROM tb_pohon_kinerja
		  WHERE id = ?
	  `
		_, err := tx.ExecContext(ctx, scriptPokin, targetTahun, validId)
		if err != nil {
			return err
		}
	}

	// 3. Dapatkan mapping ID lama ke ID baru
	scriptMapping := `
	  SELECT 
		  src.id as old_id,
		  dst.id as new_id
	  FROM tb_pohon_kinerja src
	  JOIN tb_pohon_kinerja dst ON 
		  src.nama_pohon = dst.nama_pohon AND
		  src.kode_opd = dst.kode_opd AND
		  src.level_pohon = dst.level_pohon
	  WHERE src.tahun = ? AND dst.tahun = ? 
	  AND src.kode_opd = ?
  `
	rows, err := tx.QueryContext(ctx, scriptMapping, sourceTahun, targetTahun, kodeOpd)
	if err != nil {
		return err
	}
	defer rows.Close()

	idMapping := make(map[int]int)
	for rows.Next() {
		var oldId, newId int
		if err := rows.Scan(&oldId, &newId); err != nil {
			return err
		}
		idMapping[oldId] = newId
	}

	// 4. Update parent IDs menggunakan mapping
	for oldId, newId := range idMapping {
		scriptUpdateParent := `
		  UPDATE tb_pohon_kinerja 
		  SET parent = ?
		  WHERE id = ? AND tahun = ?
	  `
		oldPokin := `SELECT parent FROM tb_pohon_kinerja WHERE id = ? AND tahun = ?`
		var oldParent int
		err := tx.QueryRowContext(ctx, oldPokin, oldId, sourceTahun).Scan(&oldParent)
		if err != nil {
			continue
		}

		if newParent, exists := idMapping[oldParent]; exists {
			_, err = tx.ExecContext(ctx, scriptUpdateParent, newParent, newId, targetTahun)
			if err != nil {
				return err
			}
		}
	}

	// 5. Clone Indikator untuk pohon kinerja yang valid
	for oldId, newId := range idMapping {
		// Clone indikator - Tambahkan kolom tahun
		scriptIndikator := `
			INSERT INTO tb_indikator (
				id, pokin_id, indikator, tahun
			)
			SELECT 
				CONCAT('IND-', UUID_SHORT()), 
				?, 
				indikator,
				''
			FROM tb_indikator
			WHERE pokin_id = ?
		`
		_, err := tx.ExecContext(ctx, scriptIndikator, newId, oldId)
		if err != nil {
			return err
		}

		// Dapatkan mapping ID indikator (menggunakan string untuk ID)
		scriptIndikatorMapping := `
            SELECT 
                src.id as old_indikator_id,
                dst.id as new_indikator_id
            FROM tb_indikator src
            JOIN tb_indikator dst ON 
                src.indikator = dst.indikator AND
                dst.pokin_id = ?
            WHERE src.pokin_id = ?
        `
		indikatorRows, err := tx.QueryContext(ctx, scriptIndikatorMapping, newId, oldId)
		if err != nil {
			return err
		}
		defer indikatorRows.Close()

		// Ubah tipe mapping menjadi string
		indikatorMapping := make(map[string]string)
		for indikatorRows.Next() {
			var oldIndikatorId, newIndikatorId string
			if err := indikatorRows.Scan(&oldIndikatorId, &newIndikatorId); err != nil {
				return err
			}
			indikatorMapping[oldIndikatorId] = newIndikatorId
		}

		// Clone target untuk setiap indikator
		for oldIndikatorId, newIndikatorId := range indikatorMapping {
			scriptTarget := `
				INSERT INTO tb_target (
					id, indikator_id, target, satuan, tahun
				)
				SELECT 
					CONCAT('TRG-', UUID_SHORT()), 
					?, 
					target, 
					satuan,
					? 
				FROM tb_target
				WHERE indikator_id = ?
			`
			_, err := tx.ExecContext(ctx, scriptTarget, newIndikatorId, targetTahun, oldIndikatorId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// count pokin pemda in opd
func (repository *PohonKinerjaRepositoryImpl) CountPokinPemdaByLevel(ctx context.Context, tx *sql.Tx, kodeOpd, tahun string) (map[int]int, error) {
	script := `
 WITH RECURSIVE pohon_all AS (
    SELECT 
        id,
        parent,
        level_pohon,
        status,
        CASE 
            WHEN level_pohon = 4 AND status = 'pokin dari pemda' AND parent = 0 THEN TRUE
            ELSE FALSE
        END as is_counted
    FROM tb_pohon_kinerja
    WHERE kode_opd = ? AND tahun = ?
),
valid_level_4 AS (
    SELECT id 
    FROM pohon_all 
    WHERE level_pohon = 4 
    AND status = 'pokin dari pemda' 
    AND parent = 0
),
pohon_hierarchy AS (
    SELECT 
        p.*,
        p.is_counted as should_count
    FROM pohon_all p
    WHERE p.level_pohon = 4

    UNION ALL

    SELECT 
        child.*,
        CASE
            WHEN child.level_pohon = 5 AND child.status = 'pokin dari pemda' THEN
                CASE
                    WHEN parent.status = 'pokin dari pemda' THEN
                        CASE WHEN (SELECT p2.parent FROM pohon_all p2 WHERE p2.id = parent.id) = 0 THEN TRUE
                        ELSE FALSE END
                    WHEN parent.status = '' THEN TRUE
                    WHEN EXISTS (
                        SELECT 1 FROM valid_level_4 
                        WHERE id = child.parent
                    ) THEN TRUE
                    ELSE FALSE
                END
            WHEN child.level_pohon >= 6 AND child.status = 'pokin dari pemda' THEN
                CASE
                    WHEN parent.status = 'pokin dari pemda' THEN
                        CASE
                            WHEN (SELECT p2.status FROM pohon_all p2 WHERE p2.id = parent.parent) = 'pokin dari pemda' THEN
                                CASE WHEN (SELECT p3.parent FROM pohon_all p3 WHERE p3.id = parent.parent) = 0 THEN TRUE
                                ELSE FALSE END
                            WHEN (SELECT p2.status FROM pohon_all p2 WHERE p2.id = parent.parent) = '' THEN TRUE
                            ELSE FALSE
                        END
                    WHEN parent.status = '' THEN TRUE
                    WHEN EXISTS (
                        WITH RECURSIVE ancestors AS (
                            SELECT p2.id, p2.parent, p2.level_pohon
                            FROM pohon_all p2
                            WHERE p2.id = child.parent
                            
                            UNION ALL
                            
                            SELECT p3.id, p3.parent, p3.level_pohon
                            FROM pohon_all p3
                            JOIN ancestors a ON p3.id = a.parent
                            WHERE p3.level_pohon >= 4
                        )
                        SELECT 1 
                        FROM ancestors a
                        JOIN valid_level_4 v ON v.id = a.id
                        WHERE a.level_pohon = 4
                    ) THEN TRUE
                    ELSE FALSE
                END
            ELSE FALSE
        END as should_count
    FROM pohon_all child
    JOIN pohon_hierarchy parent ON child.parent = parent.id
    WHERE child.level_pohon > 4
)
SELECT 
    level_pohon,
    COUNT(*) as jumlah
FROM pohon_hierarchy
WHERE should_count = TRUE
GROUP BY level_pohon
ORDER BY level_pohon;`

	rows, err := tx.QueryContext(ctx, script, kodeOpd, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var level, count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		result[level] = count
	}

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) CheckAsalPokin(ctx context.Context, tx *sql.Tx, id int) (int, error) {
	SQL := "SELECT clone_from FROM pohon_kinerja WHERE id = ?"
	var cloneFrom sql.NullInt64

	err := tx.QueryRowContext(ctx, SQL, id).Scan(&cloneFrom)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, err
		}
		return 0, err
	}

	if cloneFrom.Valid {
		return int(cloneFrom.Int64), nil
	}
	return 0, nil
}

func (repository *PohonKinerjaRepositoryImpl) DeletePokinWithIndikatorAndTarget(ctx context.Context, tx *sql.Tx, id int) error {
	// Hapus target terlebih dahulu
	scriptDeleteTarget := `
        DELETE FROM tb_target 
        WHERE indikator_id IN (
            SELECT id FROM tb_indikator 
            WHERE pokin_id = ?
        )`
	_, err := tx.ExecContext(ctx, scriptDeleteTarget, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus target: %v", err)
	}

	// Hapus indikator
	scriptDeleteIndikator := "DELETE FROM tb_indikator WHERE pokin_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeleteIndikator, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus indikator: %v", err)
	}

	// Hapus pelaksana
	scriptDeletePelaksana := "DELETE FROM tb_pelaksana_pokin WHERE pohon_kinerja_id = ?"
	_, err = tx.ExecContext(ctx, scriptDeletePelaksana, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus pelaksana: %v", err)
	}

	// Hapus pohon kinerja
	scriptDeletePokin := "DELETE FROM tb_pohon_kinerja WHERE id = ?"
	_, err = tx.ExecContext(ctx, scriptDeletePokin, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus pohon kinerja: %v", err)
	}

	return nil
}

func (repository *PohonKinerjaRepositoryImpl) FindListOpdAllTematik(ctx context.Context, tx *sql.Tx, tahun string) ([]domain.PohonKinerja, error) {
	script := `
        WITH RECURSIVE pohon_hierarki AS (
            -- Base case: pilih semua node level 0
            SELECT 
                id, 
                nama_pohon, 
                parent, 
                jenis_pohon, 
                level_pohon, 
                kode_opd, 
                keterangan, 
                tahun, 
                status, 
                is_active,
                id as root_id
            FROM tb_pohon_kinerja 
            WHERE level_pohon = 0 AND tahun = ?
            
            UNION ALL
            
            -- Recursive case: ambil semua child nodes
            SELECT 
                pk.id, 
                pk.nama_pohon, 
                pk.parent, 
                pk.jenis_pohon, 
                pk.level_pohon, 
                pk.kode_opd, 
                pk.keterangan, 
                pk.tahun, 
                pk.status, 
                pk.is_active,
                ph.root_id
            FROM tb_pohon_kinerja pk
            INNER JOIN pohon_hierarki ph ON pk.parent = ph.id
        )
        SELECT 
            ph.id,
            ph.nama_pohon,
            ph.parent,
            ph.jenis_pohon,
            ph.level_pohon,
            ph.kode_opd,
            ph.keterangan,
            ph.tahun,
            ph.status,
            ph.is_active,
            o.nama_opd,
            ph.root_id
        FROM 
            pohon_hierarki ph
        LEFT JOIN 
            tb_operasional_daerah o ON ph.kode_opd = o.kode_opd
        ORDER BY 
            ph.level_pohon, ph.id
    `

	rows, err := tx.QueryContext(ctx, script, tahun)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map untuk menyimpan pohon kinerja level 0
	pokinMap := make(map[int]domain.PohonKinerja)
	// Map untuk menyimpan OPD yang sudah diproses per pohon kinerja
	opdMap := make(map[int]map[string]string)

	for rows.Next() {
		var (
			pokinId, parent, levelPohon, rootId                            int
			namaPohon, jenisPohon, kodeOpd, keterangan, tahunPokin, status string
			is_active                                                      bool
			namaOpd                                                        sql.NullString
		)

		err := rows.Scan(
			&pokinId, &namaPohon, &parent, &jenisPohon, &levelPohon,
			&kodeOpd, &keterangan, &tahunPokin, &status, &is_active,
			&namaOpd, &rootId,
		)
		if err != nil {
			return nil, err
		}

		// Jika ini adalah level 0, simpan ke pokinMap
		if levelPohon == 0 {
			pokinMap[pokinId] = domain.PohonKinerja{
				Id:         pokinId,
				NamaPohon:  namaPohon,
				Parent:     parent,
				JenisPohon: jenisPohon,
				LevelPohon: levelPohon,
				KodeOpd:    kodeOpd,
				Keterangan: keterangan,
				Tahun:      tahunPokin,
				Status:     status,
				IsActive:   is_active,
				ListOpd:    []domain.OpdList{},
			}
			// Inisialisasi map OPD untuk level 0
			opdMap[pokinId] = make(map[string]string)
		}

		// Jika ini adalah child node (level > 0) dan memiliki kode_opd
		if levelPohon > 0 && kodeOpd != "" {
			// Gunakan rootId untuk menambahkan OPD ke parent level 0
			if _, exists := opdMap[rootId]; !exists {
				opdMap[rootId] = make(map[string]string)
			}
			// Tambahkan OPD ke map jika belum ada
			if _, exists := opdMap[rootId][kodeOpd]; !exists {
				opdMap[rootId][kodeOpd] = namaOpd.String
			}
		}
	}

	// Konversi map ke slice
	var result []domain.PohonKinerja
	for id, pokin := range pokinMap {
		// Tambahkan list OPD ke pohon kinerja
		for kodeOpd, namaOpd := range opdMap[id] {
			pokin.ListOpd = append(pokin.ListOpd, domain.OpdList{
				KodeOpd:         kodeOpd,
				PerangkatDaerah: namaOpd,
			})
		}
		result = append(result, pokin)
	}

	return result, nil
}

func (repository *PohonKinerjaRepositoryImpl) ValidateParentLevelTarikStrategiOpd(ctx context.Context, tx *sql.Tx, parentId int, childLevel int) error {
	// Jika tidak ada parent (parent = 0), tidak perlu validasi
	if parentId == 0 {
		return nil
	}

	// Ambil data parent
	query := "SELECT level_pohon FROM tb_pohon_kinerja WHERE id = ?"
	var parentLevel int
	err := tx.QueryRowContext(ctx, query, parentId).Scan(&parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parent dengan id %d tidak ditemukan", parentId)
		}
		return err
	}

	// Validasi khusus untuk level 4 (Strategic)
	if childLevel == 4 {
		// Level 4 bisa memiliki parent dari level 0,1,2,3
		if parentLevel < 0 || parentLevel > 3 {
			return fmt.Errorf("untuk level Strategic (4), parent harus memiliki level 0-3, parent level saat ini: %d", parentLevel)
		}
		return nil
	}

	// Untuk level lainnya, parent harus memiliki level tepat 1 di atasnya
	expectedParentLevel := childLevel - 1
	if parentLevel != expectedParentLevel {
		return fmt.Errorf("level parent (%d) tidak sesuai dengan yang diharapkan (%d) untuk child level %d",
			parentLevel, expectedParentLevel, childLevel)
	}

	return nil
}
