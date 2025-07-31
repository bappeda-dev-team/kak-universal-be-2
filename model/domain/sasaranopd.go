package domain

type SasaranOpd struct {
	Id                int
	IdPohon           int
	NamaPohon         string
	JenisPohon        string
	LevelPohon        int
	TahunPohon        string
	TahunAwalPeriode  string
	TahunAkhirPeriode string
	JenisPeriode      string
	Pelaksana         []PelaksanaPokin
	SasaranOpd        []SasaranOpdDetail
}

type SasaranOpdDetail struct {
	Id             int
	IdPohon        int
	NamaSasaranOpd string
	TahunAwal      string
	TahunAkhir     string
	JenisPeriode   string
	Indikator      []Indikator
}

type SasaranOpdTahunan struct {
	Id           int
	IdPohon      int
	KodeOpd      string
	NamaOpd      string
	SasaranOpd   string
	TahunAwal    string
	TahunAkhir   string
	JenisPeriode string
	JenisPohon   string
	PohonAktif   bool
	Indikator    []Indikator
}
