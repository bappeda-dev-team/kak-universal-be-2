package sasaranpemda

type SasaranPemdaResponse struct {
	Id            int                 `json:"id"`
	SubtemaId     int                 `json:"subtema_id,omitempty"`
	NamaSubtema   string              `json:"nama_subtema,omitempty"`
	TujuanPemdaId int                 `json:"tujuan_pemda_id,omitempty"`
	TujuanPemda   string              `json:"tujuan_pemda,omitempty"`
	SasaranPemda  string              `json:"sasaran_pemda"`
	JenisPohon    string              `json:"jenis_pohon"`
	Periode       PeriodeResponse     `json:"periode"`
	Indikator     []IndikatorResponse `json:"indikator"`
}

type SasaranPemdaMinimalResponse struct {
	Id           int                 `json:"id"`
	SasaranPemda string              `json:"sasaran_pemda"`
	TahunAwal    string              `json:"tahun_awal"`
	TahunAkhir   string              `json:"tahun_akhir"`
	Indikator    []IndikatorResponse `json:"indikator"`
}

type IndikatorResponse struct {
	Id               string           `json:"id"`
	Indikator        string           `json:"indikator"`
	RumusPerhitungan string           `json:"rumus_perhitungan"`
	SumberData       string           `json:"sumber_data"`
	Target           []TargetResponse `json:"target"`
}

type TargetResponse struct {
	Id     string `json:"id"`
	Target string `json:"target"`
	Satuan string `json:"satuan"`
	Tahun  string `json:"tahun"`
}

type PeriodeResponse struct {
	Id           int    `json:"id"`
	TahunAwal    string `json:"tahun_awal"`
	TahunAkhir   string `json:"tahun_akhir"`
	JenisPeriode string `json:"jenis_periode"`
}

// type TematikSasaranPemdaResponse struct {
// 	TematikId    int                             `json:"tematik_id"`
// 	NamaTematik  string                          `json:"nama_tematik"`
// 	SasaranPemda []SasaranPemdaWithPokinResponse `json:"sasaran_pemda"`
// }

// type SasaranPemdaWithPokinResponse struct {
// 	SubtematikId        int                           `json:"subtematik_id"`
// 	NamaSubtematik      string                        `json:"nama_subtematik"`
// 	JenisPohon          string                        `json:"jenis_pohon"`
// 	LevelPohon          int                           `json:"level_pohon"`
// 	IdsasaranPemda      int                           `json:"id_sasaran"`
// 	SasaranPemda        string                        `json:"sasaranpemda"`
// 	Keterangan          string                        `json:"keterangan"`
// 	IndikatorSubtematik []IndikatorSubtematikResponse `json:"indikator"`
// }

type IndikatorSubtematikResponse struct {
	Id               string           `json:"id"`
	Indikator        string           `json:"indikator"`
	RumusPerhitungan string           `json:"rumus_perhitungan"`
	SumberData       string           `json:"sumber_data"`
	Target           []TargetResponse `json:"target"`
}

type TematikResponse struct {
	TematikId   int                  `json:"tematik_id"`
	NamaTematik string               `json:"nama_tematik"`
	Subtematik  []SubtematikResponse `json:"subtematik"`
}

type SubtematikResponse struct {
	SubtematikId   int                             `json:"subtematik_id"`
	NamaSubtematik string                          `json:"nama_subtematik"`
	JenisPohon     string                          `json:"jenis_pohon"`
	LevelPohon     int                             `json:"level_pohon"`
	IsActive       bool                            `json:"is_active"`
	SasaranPemda   []SasaranPemdaWithpokinResponse `json:"sasaran_pemda"`
}

type SasaranPemdaWithpokinResponse struct {
	IdSasaranPemda int                           `json:"id_sasaran_pemda"`
	SasaranPemda   string                        `json:"sasaran_pemda"`
	Periode        PeriodeResponse               `json:"periode"`
	Indikator      []IndikatorSubtematikResponse `json:"indikator"`
}
