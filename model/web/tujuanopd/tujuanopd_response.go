package tujuanopd

type TujuanOpdResponse struct {
	Id               int                 `json:"id_tujuan_opd"`
	KodeBidangUrusan string              `json:"kode_bidang_urusan,omitempty"`
	NamaBidangUrusan string              `json:"nama_bidang_urusan,omitempty"`
	KodeOpd          string              `json:"kode_opd,omitempty"`
	NamaOpd          string              `json:"nama_opd,omitempty"`
	Tujuan           string              `json:"tujuan,omitempty"`
	RumusPerhitungan string              `json:"rumus_perhitungan,omitempty"`
	SumberData       string              `json:"sumber_data,omitempty"`
	TahunAwal        string              `json:"tahun_awal,omitempty"`
	TahunAkhir       string              `json:"tahun_akhir,omitempty"`
	JenisPeriode     string              `json:"jenis_periode,omitempty"`
	Periode          *PeriodeResponse     `json:"periode,omitempty"`
	Indikator        []IndikatorResponse `json:"indikator"`
}

type IndikatorResponse struct {
	Id               string           `json:"id"`
	IdTujuanOpd      int              `json:"id_tujuan_opd"`
	NamaIndikator    string           `json:"indikator"`
	RumusPerhitungan string           `json:"rumus_perhitungan"`
	SumberData       string           `json:"sumber_data"`
	Target           []TargetResponse `json:"target"`
}

type TargetResponse struct {
	Id              string `json:"id"`
	IndikatorId     string `json:"indikator_id"`
	Tahun           string `json:"tahun"`
	TargetIndikator string `json:"target"`
	SatuanIndikator string `json:"satuan"`
}

type PeriodeResponse struct {
	Id           int    `json:"id"`
	TahunAwal    string `json:"tahun_awal"`
	TahunAkhir   string `json:"tahun_akhir"`
	JenisPeriode string `json:"jenis_periode"`
}

type TujuanOpdwithBidangUrusanResponse struct {
	KodeUrusan       string              `json:"kode_urusan"`
	Urusan           string              `json:"urusan"`
	KodeBidangUrusan string              `json:"kode_bidang_urusan"`
	NamaBidangUrusan string              `json:"nama_bidang_urusan"`
	KodeOpd          string              `json:"kode_opd"`
	NamaOpd          string              `json:"nama_opd"`
	TujuanOpd        []TujuanOpdResponse `json:"tujuan_opd"`
}
