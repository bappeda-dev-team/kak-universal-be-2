package iku

import "time"

type IkuResponse struct {
	IndikatorId      string           `json:"indikator_id"`
	Sumber           string           `json:"asal_iku"`
	IsActive         bool             `json:"is_active"`
	Indikator        string           `json:"indikator"`
	RumusPerhitungan string           `json:"rumus_perhitungan"`
	SumberData       string           `json:"sumber_data"`
	CreatedAt        time.Time        `json:"created_at"`
	TahunAwal        string           `json:"tahun_awal"`
	TahunAkhir       string           `json:"tahun_akhir"`
	JenisPeriode     string           `json:"jenis_periode"`
	Target           []TargetResponse `json:"target"`
}

type TargetResponse struct {
	Id     string `json:"id"`
	Target string `json:"target"`
	Satuan string `json:"satuan"`
	Tahun  string `json:"tahun"`
}

type IkuOpdResponse struct {
	IndikatorId      string              `json:"indikator_id"`
	AsalIku          string              `json:"asal_iku"`
	Indikator        string              `json:"indikator"`
	CreatedAt        time.Time           `json:"created_at"`
	RumusPerhitungan string              `json:"rumus_perhitungan"`
	SumberData       string              `json:"sumber_data"`
	TahunAwal        string              `json:"tahun_awal"`
	TahunAkhir       string              `json:"tahun_akhir"`
	JenisPeriode     string              `json:"jenis_periode"`
	Target           []TargetOpdResponse `json:"target"`
}

type TargetOpdResponse struct {
	Target string `json:"target"`
	Satuan string `json:"satuan"`
	Tahun  string `json:"tahun"`
}
