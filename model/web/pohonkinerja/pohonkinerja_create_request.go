package pohonkinerja

type PohonKinerjaCreateRequest struct {
	Parent      int                      `json:"parent"`
	NamaPohon   string                   `json:"nama_pohon"`
	JenisPohon  string                   `json:"jenis_pohon"`
	LevelPohon  int                      `json:"level_pohon"`
	KodeOpd     string                   `json:"kode_opd"`
	Keterangan  string                   `json:"keterangan"`
	Tahun       string                   `json:"tahun"`
	Status      string                   `json:"status"`
	PelaksanaId []PelaksanaCreateRequest `json:"pelaksana"`
	Indikator   []IndikatorCreateRequest `json:"indikator"`
}

type PelaksanaCreateRequest struct {
	IdPelaksana string `json:"id_pelaksana"`
	PegawaiId   string `json:"pegawai_id"`
}

type PohonKinerjaAdminCreateRequest struct {
	Parent     int                      `json:"parent"`
	NamaPohon  string                   `json:"nama_pohon"`
	JenisPohon string                   `json:"jenis_pohon"`
	KodeOpd    string                   `json:"kode_opd,omitempty"`
	LevelPohon int                      `json:"level_pohon"`
	Keterangan string                   `json:"keterangan"`
	Tahun      string                   `json:"tahun"`
	Status     string                   `json:"status"`
	Pelaksana  []PelaksanaCreateRequest `json:"pelaksana"`
	Indikator  []IndikatorCreateRequest `json:"indikator"`
}

type PohonKinerjaAdminStrategicCreateRequest struct {
	IdToClone int `json:"id"`
	Parent    int `json:"parent"`
	// LevelPohon int `json:"level_pohon"`
	JenisPohon string `json:"jenis_pohon"`
	Turunan    bool   `json:"turunan"`
}

type IndikatorCreateRequest struct {
	PohonKinerjaId int                   `json:"pohon_id"`
	NamaIndikator  string                `json:"indikator"`
	Target         []TargetCreateRequest `json:"target"`
}

type TargetCreateRequest struct {
	IndikatorId int    `json:"indikator_id"`
	Target      string `json:"target"`
	Satuan      string `json:"satuan"`
}

type TematikStatusRequest struct {
	Id       int  `json:"id" validate:"required"`
	IsActive bool `json:"is_active"`
}

type PohonKinerjaCloneRequest struct {
	KodeOpd     string `json:"kode_opd" validate:"required"`
	TahunSumber string `json:"tahun_sumber" validate:"required"`
	TahunTujuan string `json:"tahun_tujuan" validate:"required"`
}
