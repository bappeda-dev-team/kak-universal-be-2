package pohonkinerja

import (
	"ekak_kabupaten_madiun/model/web/opdmaster"
)

type PohonKinerjaOpdResponse struct {
	Id                     int                    `json:"id"`
	Parent                 string                 `json:"parent"`
	NamaPohon              string                 `json:"nama_pohon"`
	JenisPohon             string                 `json:"jenis_pohon"`
	LevelPohon             int                    `json:"level_pohon"`
	KodeOpd                string                 `json:"kode_opd,omitempty"`
	NamaOpd                string                 `json:"nama_opd,omitempty"`
	Keterangan             string                 `json:"keterangan,omitempty"`
	Tahun                  string                 `json:"tahun,omitempty"`
	CountReview            int                    `json:"jumlah_review"`
	Status                 string                 `json:"status"`
	Pelaksana              []PelaksanaOpdResponse `json:"pelaksana"`
	Indikator              []IndikatorResponse    `json:"indikator"`
	KeteranganCrosscutting *string                `json:"keterangan_crosscutting"`
}

type PohonKinerjaOpdAllResponse struct {
	KodeOpd    string                 `json:"kode_opd"`
	NamaOpd    string                 `json:"nama_opd"`
	Tahun      string                 `json:"tahun"`
	TujuanOpd  []TujuanOpdResponse    `json:"tujuan_opd"`
	Strategics []StrategicOpdResponse `json:"childs"`
}

type StrategicOpdResponse struct {
	Id                     int                         `json:"id"`
	Parent                 *int                        `json:"parent"`
	Strategi               string                      `json:"nama_pohon"`
	JenisPohon             string                      `json:"jenis_pohon"`
	LevelPohon             int                         `json:"level_pohon"`
	Keterangan             string                      `json:"keterangan"`
	KeteranganCrosscutting *string                     `json:"keterangan_crosscutting"`
	Status                 string                      `json:"status"`
	CountReview            int                         `json:"jumlah_review"`
	KodeOpd                opdmaster.OpdResponseForAll `json:"perangkat_daerah"`
	IsActive               bool                        `json:"is_active"`
	Pelaksana              []PelaksanaOpdResponse      `json:"pelaksana"`
	Indikator              []IndikatorResponse         `json:"indikator"`
	Tacticals              []TacticalOpdResponse       `json:"childs,omitempty"`
	Crosscutting           []CrosscuttingOpdResponse   `json:"crosscutting,omitempty"`
	Review                 []ReviewResponse            `json:"review,omitempty"`
}

type TacticalOpdResponse struct {
	Id                     int                         `json:"id"`
	Parent                 int                         `json:"parent"`
	Strategi               string                      `json:"nama_pohon"`
	JenisPohon             string                      `json:"jenis_pohon"`
	LevelPohon             int                         `json:"level_pohon"`
	Keterangan             string                      `json:"keterangan"`
	KeteranganCrosscutting *string                     `json:"keterangan_crosscutting"`
	Status                 string                      `json:"status"`
	CountReview            int                         `json:"jumlah_review"`
	KodeOpd                opdmaster.OpdResponseForAll `json:"perangkat_daerah"`
	IsActive               bool                        `json:"is_active"`
	Pelaksana              []PelaksanaOpdResponse      `json:"pelaksana"`
	Indikator              []IndikatorResponse         `json:"indikator"`
	Operationals           []OperationalOpdResponse    `json:"childs,omitempty"`
	Crosscutting           []CrosscuttingOpdResponse   `json:"crosscutting,omitempty"`
	Review                 []ReviewResponse            `json:"review,omitempty"`
}

type OperationalOpdResponse struct {
	Id                     int                         `json:"id"`
	Parent                 int                         `json:"parent"`
	Strategi               string                      `json:"nama_pohon"`
	JenisPohon             string                      `json:"jenis_pohon"`
	LevelPohon             int                         `json:"level_pohon"`
	Keterangan             string                      `json:"keterangan"`
	KeteranganCrosscutting *string                     `json:"keterangan_crosscutting"`
	Status                 string                      `json:"status"`
	CountReview            int                         `json:"jumlah_review"`
	KodeOpd                opdmaster.OpdResponseForAll `json:"perangkat_daerah"`
	IsActive               bool                        `json:"is_active"`
	Pelaksana              []PelaksanaOpdResponse      `json:"pelaksana"`
	Indikator              []IndikatorResponse         `json:"indikator"`
	Childs                 []OperationalNOpdResponse   `json:"childs,omitempty"`
	Crosscutting           []CrosscuttingOpdResponse   `json:"crosscutting,omitempty"`
	Review                 []ReviewResponse            `json:"review,omitempty"`
}

type OperationalNOpdResponse struct {
	Id                     int                         `json:"id"`
	Parent                 int                         `json:"parent"`
	Strategi               string                      `json:"nama_pohon"`
	JenisPohon             string                      `json:"jenis_pohon"`
	LevelPohon             int                         `json:"level_pohon"`
	Keterangan             string                      `json:"keterangan"`
	KeteranganCrosscutting *string                     `json:"keterangan_crosscutting"`
	Status                 string                      `json:"status"`
	CountReview            int                         `json:"jumlah_review"`
	KodeOpd                opdmaster.OpdResponseForAll `json:"perangkat_daerah"`
	IsActive               bool                        `json:"is_active"`
	Pelaksana              []PelaksanaOpdResponse      `json:"pelaksana"`
	Indikator              []IndikatorResponse         `json:"indikator"`
	Childs                 []OperationalNOpdResponse   `json:"childs,omitempty"`
	Review                 []ReviewResponse            `json:"review,omitempty"`
}

type PelaksanaOpdResponse struct {
	Id             string `json:"id_pelaksana"`
	PohonKinerjaId string `json:"pohon_kinerja_id,omitempty"`
	PegawaiId      string `json:"pegawai_id"`
	Nip            string `json:"nip"`
	NamaPegawai    string `json:"nama_pegawai"`
}

type TujuanOpdResponse struct {
	Id      int    `json:"id"`
	KodeOpd string `json:"kode_opd"`
	Tujuan  string `json:"tujuan"`
	// Periode   PeriodeResponse           `json:"periode,omitempty"`
	Indikator []IndikatorTujuanResponse `json:"indikator"`
}

type IndikatorTujuanResponse struct {
	Indikator string                 `json:"indikator"`
	Target    []TargetTujuanResponse `json:"targets"`
}

type TargetTujuanResponse struct {
	Tahun  string `json:"tahun"`
	Target string `json:"target"`
	Satuan string `json:"satuan"`
}

type PeriodeResponse struct {
	Id         int    `json:"id"`
	TahunAwal  string `json:"tahun_awal"`
	TahunAkhir string `json:"tahun_akhir"`
}

type CountPokinPemdaResponse struct {
	KodeOpd     string        `json:"kode_opd"`
	NamaOpd     string        `json:"nama_opd"`
	Tahun       string        `json:"tahun"`
	TotalPemda  int           `json:"total_pemda"`
	DetailLevel []LevelDetail `json:"detail_level"`
}

type LevelDetail struct {
	Level       int    `json:"level"`
	JenisPohon  string `json:"jenis_pohon"`
	JumlahPemda int    `json:"jumlah_pemda"`
}
