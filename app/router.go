package app

import (
	"ekak_kabupaten_madiun/controller"

	"github.com/julienschmidt/httprouter"
)

type RouteController struct {
}

func NewRouter(
	rencanaKinerjaController controller.RencanaKinerjaController,
	rencanaAksiController controller.RencanaAksiController,
	pelaksanaanRencanaAksiController controller.PelaksanaanRencanaAksiController,
	usulanMusrebangController controller.UsulanMusrebangController,
	usulanMandatoriController controller.UsulanMandatoriController,
	usulanPokokPikiranController controller.UsulanPokokPikiranController,
	usulanInisiatifController controller.UsulanInisiatifController,
	usulanTerpilihController controller.UsulanTerpilihController,
	gambaranUmumController controller.GambaranUmumController,
	dasarHukumController controller.DasarHukumController,
	inovasiController controller.InovasiController,
	subKegiatanController controller.SubKegiatanController,
	subKegiatanTerpilihController controller.SubKegiatanTerpilihController,
	pohonKinerjaOpdController controller.PohonKinerjaOpdController,
	pegawaiController controller.PegawaiController,
	lembagaController controller.LembagaController,
	jabatanController controller.JabatanController,
	pohonKinerjaAdminController controller.PohonKinerjaAdminController,
	opdController controller.OpdController,
	programController controller.ProgramController,
	urusanController controller.UrusanController,
	bidangUrusanController controller.BidangUrusanController,
	kegiatanController controller.KegiatanController,
	userController controller.UserController,
	roleController controller.RoleController,
	tujuanOpdController controller.TujuanOpdController,
	crosscuttingOpdController controller.CrosscuttingOpdController,
	manualIKController controller.ManualIKController,
	reviewController controller.ReviewController,
	periodeController controller.PeriodeController,
	tujuanPemdaController controller.TujuanPemdaController,
	sasaranPemdaController controller.SasaranPemdaController,
	permasalahanRekinController controller.PermasalahanRekinController,
	ikuController controller.IkuController,
	sasaranOpdController controller.SasaranOpdController,
	visiPemdaController controller.VisiPemdaController,
	misiPemdaController controller.MisiPemdaController,
	matrixRenstraController controller.MatrixRenstraController,
	cascadingOpdController controller.CascadingOpdController,
	rincianBelanjaController controller.RincianBelanjaController,
	kelompokAnggaranController controller.KelompokAnggaranController,
) *httprouter.Router {
	router := httprouter.New()

	//rencana_kinerja
	router.POST("/rencana_kinerja/create", rencanaKinerjaController.Create)
	router.GET("/get_rencana_kinerja/pegawai/:pegawai_id", rencanaKinerjaController.FindAllRencanaKinerja)
	router.GET("/detail-rencana_kinerja/:rencana_kinerja_id", rencanaKinerjaController.FindById)
	router.PUT("/rencana_kinerja/update/:id", rencanaKinerjaController.Update)
	router.DELETE("/rencana_kinerja/delete/:id", rencanaKinerjaController.Delete)
	router.GET("/rencana_kinerja_pokin/pokin_by_pelaksana/:pegawai_id/:tahun", pohonKinerjaOpdController.FindPokinByPelaksana)
	router.POST("/rencana_kinerja/create_level1", rencanaKinerjaController.CreateRekinLevel1)
	router.PUT("/rencana_kinerja/update_level1/:id", rencanaKinerjaController.UpdateRekinLevel1)
	router.GET("/rencana_kinerja_level1/:id", rencanaKinerjaController.FindIdRekinLevel1)
	router.GET("/rencana_kinerja_level3/:kode_opd/:tahun", rencanaKinerjaController.FindRekinLevel3)
	// router.GET("/rencana_kinerja_sasaran_opd/pegawai_level1/:pegawai_id/tahun/:tahun", rencanaKinerjaController.FindRekinSasaranOpd)

	//rencana_aksi
	router.GET("/rencana_aksi/findall/:rencana_kinerja_id", rencanaAksiController.FindAll)
	// router.GET("/rencana_kinerja/:rekin_id/rincian_kak", rencanaAksiController.FindAll)
	router.GET("/detail-rencana_aksi/:rencanaaksiId", rencanaAksiController.FindById)
	router.POST("/rencana_aksi/create/rencanaaksi/:rekin_id", rencanaAksiController.Create)
	router.PUT("/rencana_aksi/update/rencanaaksi/:rencanaaksiId", rencanaAksiController.Update)
	router.DELETE("/rencana_aksi/delete/rencanaaksi/:rencanaaksiId", rencanaAksiController.Delete)

	//pelaksanaan_rencana_aksi
	router.POST("/pelaksanaan_rencana_aksi/create/:rencanaAksiId", pelaksanaanRencanaAksiController.Create)
	router.PUT("/pelaksanaan_rencana_aksi/update/:pelaksanaanRencanaAksiId", pelaksanaanRencanaAksiController.Update)
	router.GET("/pelaksanaan_rencana_aksi/detail/:id", pelaksanaanRencanaAksiController.FindById)
	router.DELETE("/pelaksanaan_rencana_aksi/delete/:id", pelaksanaanRencanaAksiController.Delete)

	//usulan musrebang
	router.POST("/usulan_musrebang/create", usulanMusrebangController.Create)
	router.PUT("/usulan_musrebang/update/:id", usulanMusrebangController.Update)
	router.PUT("/usulan_musrebang/update/:id/:pegawai_id", usulanMusrebangController.Update)
	router.GET("/usulan_musrebang/detail/:id", usulanMusrebangController.FindById)
	router.DELETE("/usulan_musrebang/delete/:id", usulanMusrebangController.Delete)
	router.GET("/usulan_musrebang/pilihan", usulanMusrebangController.FindAll)
	router.GET("/usulan_musrebang/findall", usulanMusrebangController.FindAll)
	router.GET("/usulan_musrebang/opd/:kode_opd", usulanMusrebangController.FindAll)
	router.POST("/usulan_musrebang/create_rekin/:rencana_kinerja_id", usulanMusrebangController.CreateRekin)
	router.DELETE("/usulan_musrebang/delete_usulan_terpilih/:id", usulanMusrebangController.DeleteUsulanTerpilih)

	//usulan mandatori
	router.POST("/usulan_mandatori/create", usulanMandatoriController.Create)
	router.POST("/usulan_mandatori/create/:pegawai_id", usulanMandatoriController.Create)
	router.PUT("/usulan_mandatori/update/:id", usulanMandatoriController.Update)
	router.PUT("/usulan_mandatori/update/:id/:pegawai_id", usulanMandatoriController.Update)
	router.GET("/usulan_mandatori/detail/:id", usulanMandatoriController.FindById)
	router.DELETE("/usulan_mandatori/delete/:id", usulanMandatoriController.Delete)
	router.GET("/usulan_mandatori/findall", usulanMandatoriController.FindAll)
	router.GET("/usulan_mandatori/pilihan", usulanMandatoriController.FindAll)
	router.GET("/usulan_mandatori/pegawai/:pegawai_id", usulanMandatoriController.FindAll)

	//usulan pokok pikiran
	router.POST("/usulan_pokok_pikiran/create", usulanPokokPikiranController.Create)
	router.POST("/usulan_pokok_pikiran/create/:pegawai_id", usulanPokokPikiranController.Create)
	router.PUT("/usulan_pokok_pikiran/update/:id", usulanPokokPikiranController.Update)
	router.PUT("/usulan_pokok_pikiran/update/:id/:pegawai_id", usulanPokokPikiranController.Update)
	router.GET("/usulan_pokok_pikiran/detail/:id", usulanPokokPikiranController.FindById)
	router.DELETE("/usulan_pokok_pikiran/delete/:id", usulanPokokPikiranController.Delete)
	router.GET("/usulan_pokok_pikiran/findall", usulanPokokPikiranController.FindAll)
	router.GET("/usulan_pokok_pikiran/pilihan", usulanPokokPikiranController.FindAll)
	router.GET("/usulan_pokok_pikiran/opd/:kode_opd", usulanPokokPikiranController.FindAll)
	router.POST("/usulan_pokok_pikiran/create_rekin/:rencana_kinerja_id", usulanPokokPikiranController.CreateRekin)
	router.DELETE("/usulan_pokok_pikiran/delete_usulan_terpilih/:id", usulanPokokPikiranController.DeleteUsulanTerpilih)

	//usulan inisiatif
	router.POST("/usulan_inisiatif/create", usulanInisiatifController.Create)
	router.POST("/usulan_inisiatif/create/:pegawai_id", usulanInisiatifController.Create)
	router.PUT("/usulan_inisiatif/update/:id", usulanInisiatifController.Update)
	router.PUT("/usulan_inisiatif/update/:id/:pegawai_id", usulanInisiatifController.Update)
	router.GET("/usulan_inisiatif/detail/:id", usulanInisiatifController.FindById)
	router.DELETE("/usulan_inisiatif/delete/:id", usulanInisiatifController.Delete)
	router.GET("/usulan_inisiatif/findall", usulanInisiatifController.FindAll)
	router.GET("/usulan_inisiatif/pilihan", usulanInisiatifController.FindAll)
	router.GET("/usulan_inisiatif/pegawai/:pegawai_id", usulanInisiatifController.FindAll)

	//gambaran umum
	router.POST("/gambaran_umum/create/:rencana_kinerja_id", gambaranUmumController.Create)
	router.GET("/gambaran_umum/findall/:rencana_kinerja_id", gambaranUmumController.FindAll)
	router.GET("/gambaran_umum/detail/:id", gambaranUmumController.FindById)
	router.PUT("/gambaran_umum/update/:id", gambaranUmumController.Update)
	router.DELETE("/gambaran_umum/delete/:id", gambaranUmumController.Delete)

	//dasar hukum
	router.POST("/dasar_hukum/create/:rencana_kinerja_id", dasarHukumController.Create)
	router.GET("/dasar_hukum/findall/:rencana_kinerja_id", dasarHukumController.FindAll)
	router.GET("/dasar_hukum/detail/:id", dasarHukumController.FindById)
	router.PUT("/dasar_hukum/update/:id", dasarHukumController.Update)
	router.DELETE("/dasar_hukum/delete/:id", dasarHukumController.Delete)

	//inovasi
	router.POST("/inovasi/create/:rencana_kinerja_id", inovasiController.Create)
	router.GET("/inovasi/findall/:rencana_kinerja_id", inovasiController.FindAll)
	router.GET("/inovasi/detail/:id", inovasiController.FindById)
	router.PUT("/inovasi/update/:id", inovasiController.Update)
	router.DELETE("/inovasi/delete/:id", inovasiController.Delete)

	//sub kegiatan
	router.POST("/sub_kegiatan/create", subKegiatanController.Create)
	router.PUT("/sub_kegiatan/update/:id", subKegiatanController.Update)
	router.GET("/sub_kegiatan/detail/:id", subKegiatanController.FindById)
	router.GET("/sub_kegiatan/findall", subKegiatanController.FindAll)
	router.GET("/sub_kegiatan/pilihan/:kode_opd", subKegiatanController.FindAll)
	router.GET("/sub_kegiatan/byrekinid/:rencana_kinerja_id", subKegiatanController.FindAll)
	router.DELETE("/sub_kegiatan/delete/:id", subKegiatanController.Delete)

	//sub kegiatan terpilih
	router.POST("/sub_kegiatan/create_rekin/:rencana_kinerja_id", subKegiatanTerpilihController.CreateRekin)
	router.DELETE("/sub_kegiatan/delete_subkegiatan_terpilih/:id", subKegiatanTerpilihController.DeleteSubKegiatanTerpilih)
	router.PUT("/subkegiatanterpilih/create/:rencana_kinerja_id", subKegiatanTerpilihController.Update)
	router.DELETE("/subkegiatanterpilih/delete/:rencana_kinerja_id/:kode_subkegiatan", subKegiatanTerpilihController.Delete)
	router.GET("/subkegiatanterpilih/findbykodesubkegiatan/:kode_subkegiatan", subKegiatanTerpilihController.FindByKodeSubKegiatan)

	//pohon kinerja opd
	router.POST("/pohon_kinerja_opd/create", pohonKinerjaOpdController.Create)
	router.PUT("/pohon_kinerja_opd/update/:id", pohonKinerjaOpdController.Update)
	router.GET("/pohon_kinerja_opd/detail/:id", pohonKinerjaOpdController.FindById)
	router.DELETE("/pohon_kinerja_opd/delete/:id", pohonKinerjaOpdController.Delete)
	router.GET("/pohon_kinerja_opd/findall/:kode_opd/:tahun", pohonKinerjaOpdController.FindAll)
	router.GET("/pohon_kinerja_opd/strategic_no_parent/:kode_opd/:tahun", pohonKinerjaOpdController.FindStrategicNoParent)
	router.DELETE("/pohon_kinerja_opd/delete_pelaksana/:id", pohonKinerjaOpdController.DeletePelaksana)
	router.DELETE("/pohon_kinerja_opd/delete_pokin_pemda/:id", pohonKinerjaOpdController.DeletePokinPemdaInOpd)
	router.PUT("/pohon_kinerja_opd/pindah_parent/:id", pohonKinerjaOpdController.UpdateParent)

	//pohon kinerja admin
	router.POST("/pohon_kinerja_admin/create", pohonKinerjaAdminController.Create)
	router.PUT("/pohon_kinerja_admin/update/:pohonKinerjaId", pohonKinerjaAdminController.Update)
	router.GET("/pohon_kinerja_admin/detail/:id", pohonKinerjaAdminController.FindById)
	router.DELETE("/pohon_kinerja_admin/delete/:pohonKinerjaId", pohonKinerjaAdminController.Delete)
	router.GET("/pohon_kinerja_admin/findall/:tahun", pohonKinerjaAdminController.FindAll)
	router.GET("/pohon_kinerja_admin/subtematik/:tahun", pohonKinerjaAdminController.FindSubTematik)
	router.GET("/pohon_kinerja_admin/tematik/:idPokin", pohonKinerjaAdminController.FindPokinAdminByIdHierarki)
	router.POST("/pohon_kinerja_admin/clone_strategic/create", pohonKinerjaAdminController.CreateStrategicAdmin)
	router.POST("/pohon_kinerja_admin/clone_pokin_pemda/create", pohonKinerjaAdminController.CloneStrategiFromPemda)
	router.PUT("/pohon_kinerja_admin/tolak_pokin/:pohonKinerjaId", pohonKinerjaAdminController.UpdatePokinStatusTolak)
	router.GET("/pohon_kinerja_admin/crosscutting/:kode_opd/:tahun", pohonKinerjaAdminController.FindPokinByCrosscuttingStatus)
	router.POST("/pokin/activation_tematik/:id", pohonKinerjaAdminController.AktiforNonAktifTematik)
	router.GET("/pokin_tematik/list_opd/:tahun", pohonKinerjaAdminController.FindListOpdAllTematik)
	// router.POST("/pohon_kinerja_admin/crosscutting/create", pohonKinerjaAdminController.CrosscuttingOpd)
	// router.PUT("/pohon_kinerja_admin/setujui_crosscutting/:pohonKinerjaId", pohonKinerjaAdminController.SetujuiCrosscutting)
	// router.PUT("/pohon_kinerja_admin/tolak_crosscutting/:pohonKinerjaId", pohonKinerjaAdminController.TolakCrosscutting)

	//pohon kinerja for dropdown
	router.GET("/pohon_kinerja/tematik/:tahun", pohonKinerjaAdminController.FindPokinByTematik)
	router.GET("/pohon_kinerja/strategic/:kode_opd/:tahun", pohonKinerjaAdminController.FindPokinByStrategic)
	router.GET("/pohon_kinerja/tactical/:kode_opd/:tahun", pohonKinerjaAdminController.FindPokinByTactical)
	router.GET("/pohon_kinerja/operational/:kode_opd/:tahun", pohonKinerjaAdminController.FindPokinByOperational)
	router.GET("/pohon_kinerja/status/:kode_opd/:tahun", pohonKinerjaAdminController.FindPokinByStatus)
	router.GET("/pohon_kinerja/pemda/:kode_opd/:tahun", pohonKinerjaAdminController.FindPokinFromPemda)
	router.GET("/pohon_kinerja/pilih_parent/:kode_opd/:tahun/:level_pohon", pohonKinerjaAdminController.FindPokinFromOpd)
	router.GET("/pohon_kinerja_opd/pokinpemda_review/:id", pohonKinerjaOpdController.FindidPokinWithAllTema)

	//DATA MASTER
	//pegawai
	router.POST("/pegawai/create", pegawaiController.Create)
	router.PUT("/pegawai/update/:id", pegawaiController.Update)
	router.GET("/pegawai/detail/:id", pegawaiController.FindById)
	router.DELETE("/pegawai/delete/:id", pegawaiController.Delete)
	router.GET("/pegawai/findall", pegawaiController.FindAll)

	//lembaga
	router.POST("/lembaga/create", lembagaController.Create)
	router.PUT("/lembaga/update/:id", lembagaController.Update)
	router.GET("/lembaga/detail/:id", lembagaController.FindById)
	router.DELETE("/lembaga/delete/:id", lembagaController.Delete)
	router.GET("/lembaga/findall", lembagaController.FindAll)

	//jabatan
	router.POST("/jabatan/create", jabatanController.Create)
	router.PUT("/jabatan/update/:id", jabatanController.Update)
	router.GET("/jabatan/detail/:id", jabatanController.FindById)
	router.DELETE("/jabatan/delete/:id", jabatanController.Delete)
	router.GET("/jabatan/findall/:kode_opd", jabatanController.FindAll)
	router.GET("/jabatan/findall/:kode_opd/:tahun", jabatanController.FindAll)

	//opd
	router.POST("/opd/create", opdController.Create)
	router.PUT("/opd/update/:opdId", opdController.Update)
	router.GET("/opd/detail/:opdId", opdController.FindById)
	router.DELETE("/opd/delete/:opdId", opdController.Delete)
	router.GET("/opd/findall", opdController.FindAll)

	//program
	router.POST("/program_kegiatan/create", programController.Create)
	router.PUT("/program_kegiatan/update/:programId", programController.Update)
	router.GET("/program_kegiatan/detail/:id", programController.FindById)
	router.DELETE("/program_kegiatan/delete/:id", programController.Delete)
	router.GET("/program_kegiatan/findall", programController.FindAll)

	//urusan
	router.POST("/urusan/create", urusanController.Create)
	router.PUT("/urusan/update/:id", urusanController.Update)
	router.GET("/urusan/detail/:id", urusanController.FindById)
	router.DELETE("/urusan/delete/:id", urusanController.Delete)
	router.GET("/urusan/findall", urusanController.FindAll)
	// router.GET("/urusan/findall/:kode_opd", urusanController.FindByKodeOpd)
	router.GET("/urusan/findall/:kode_opd/urusan_bidang", urusanController.FindUrusanAndBidangByKodeOpd)

	//bidang urusan
	router.POST("/bidang_urusan/create", bidangUrusanController.Create)
	router.PUT("/bidang_urusan/update/:id", bidangUrusanController.Update)
	router.GET("/bidang_urusan/detail/:id", bidangUrusanController.FindById)
	router.DELETE("/bidang_urusan/delete/:id", bidangUrusanController.Delete)
	router.GET("/bidang_urusan/findall", bidangUrusanController.FindAll)
	router.GET("/bidang_urusan/findall/:kode_opd", bidangUrusanController.FindByKodeOpd)

	//kegiatan
	router.POST("/kegiatan/create", kegiatanController.Create)
	router.PUT("/kegiatan/update/:id", kegiatanController.Update)
	router.GET("/kegiatan/detail/:id", kegiatanController.FindById)
	router.DELETE("/kegiatan/delete/:id", kegiatanController.Delete)
	router.GET("/kegiatan/findall", kegiatanController.FindAll)

	//rincian kak
	router.GET("/rencana_kinerja/:rencana_kinerja_id/pegawai/:pegawai_id/input_rincian_kak", rencanaKinerjaController.FindAllRincianKak)

	//role
	router.POST("/role/create", roleController.Create)
	router.PUT("/role/update/:id", roleController.Update)
	router.GET("/role/detail/:id", roleController.FindById)
	router.DELETE("/role/delete/:id", roleController.Delete)
	router.GET("/role/findall", roleController.FindAll)

	//user
	router.POST("/user/create", userController.Create)
	router.PUT("/user/update/:id", userController.Update)
	router.GET("/user/detail/:id", userController.FindById)
	router.DELETE("/user/delete/:id", userController.Delete)
	router.GET("/user/findall", userController.FindAll)
	router.POST("/user/login", userController.Login)
	router.GET("/user/findbykodeopdandrole", userController.FindByKodeOpdAndRole)

	//tujuan opd
	router.POST("/tujuan_opd/create", tujuanOpdController.Create)
	router.PUT("/tujuan_opd/update/:tujuanOpdId", tujuanOpdController.Update)
	router.GET("/tujuan_opd/detail/:tujuanOpdId", tujuanOpdController.FindById)
	router.DELETE("/tujuan_opd/delete/:tujuanOpdId", tujuanOpdController.Delete)
	router.GET("/tujuan_opd/findall/:kode_opd/tahunawal/:tahun_awal/tahunakhir/:tahun_akhir/jenisperiode/:jenis_periode", tujuanOpdController.FindAll)

	//crosscutting opd
	router.POST("/crosscutting_opd/create/:parentId", crosscuttingOpdController.Create)
	router.PUT("/crosscutting_opd/update/:crosscuttingId", crosscuttingOpdController.Update)
	router.DELETE("/crosscutting_opd/delete/:crosscuttingId/:nip_pegawai", crosscuttingOpdController.Delete)
	router.GET("/crosscutting_opd/findall/:parentId", crosscuttingOpdController.FindAll)
	router.POST("/crosscutting/:crosscuttingId/permission", crosscuttingOpdController.ApproveOrReject)
	router.DELETE("/crosscutting/:crosscuttingId/unused", crosscuttingOpdController.DeleteUnused)
	router.GET("/crosscutting_menunggu/:kode_opd/:tahun", crosscuttingOpdController.FindPokinByCrosscuttingStatus)
	router.GET("/crosscutting_opd/opd-from/:crosscuttingTo", crosscuttingOpdController.FindOPDCrosscuttingFrom)

	//manual ik
	router.POST("/manual_ik/create/:indikatorId", manualIKController.Create)
	router.PUT("/manual_ik/update/:indikatorId", manualIKController.Update)
	router.GET("/manual_ik/detail/:indikatorId", manualIKController.FindManualIKByIndikatorId)
	router.GET("/manual_ik/sasaran_opd/:indikatorId/:tahun", manualIKController.FindManualIKSasaranOpdByIndikatorId)

	//review
	router.POST("/review_pokin/create/:pokinId", reviewController.Create)
	router.PUT("/review_pokin/update/:id", reviewController.Update)
	router.DELETE("/review_pokin/delete/:id", reviewController.Delete)
	router.GET("/review_pokin/findall/:pokin_id", reviewController.FindAll)
	router.GET("/review_pokin/detail/:id", reviewController.FindById)
	router.GET("/review_pokin/tematik/:tahun", reviewController.FindAllReviewByTematik)
	router.GET("/review_pokin/opd/:kode_opd/:tahun", reviewController.FindAllReviewOpd)

	//periode
	router.POST("/periode/create", periodeController.Create)
	router.PUT("/periode/update/:id", periodeController.Update)
	router.GET("/periode/tahun/:tahun", periodeController.FindByTahun)
	router.GET("/periode/findall", periodeController.FindAll)
	router.GET("/periode/detail/:id", periodeController.FindById)
	router.DELETE("/periode/delete/:id", periodeController.Delete)

	//tujuan pemda
	router.POST("/tujuan_pemda/create", tujuanPemdaController.Create)
	router.PUT("/tujuan_pemda/update/:id", tujuanPemdaController.Update)
	router.DELETE("/tujuan_pemda/delete/:id", tujuanPemdaController.Delete)
	router.GET("/tujuan_pemda/detail/:id", tujuanPemdaController.FindById)
	router.GET("/tujuan_pemda/findall/:tahun/:jenis_periode", tujuanPemdaController.FindAll)
	router.PUT("/tujuan_pemda/update_periode/:id", tujuanPemdaController.UpdatePeriode)
	router.GET("/tujuan_pemda/findall_with_pokin/:tahun_awal/:tahun_akhir/:jenis_periode", tujuanPemdaController.FindAllWithPokin)
	router.GET("/pohon_kinerja/pokin_with_periode/:pokin_id/:jenis_periode", tujuanPemdaController.FindPokinWithPeriode)

	//sasaran pemda
	router.POST("/sasaran_pemda/create", sasaranPemdaController.Create)
	router.PUT("/sasaran_pemda/update/:id", sasaranPemdaController.Update)
	router.DELETE("/sasaran_pemda/delete/:id", sasaranPemdaController.Delete)
	router.GET("/sasaran_pemda/detail/:id", sasaranPemdaController.FindById)
	// router.GET("/sasaran_pemda/findall/:tahun", sasaranPemdaController.FindAll)
	router.GET("/sasaran_pemda/findall/tahun_awal/:tahun_awal/tahun_akhir/:tahun_akhir/jenis_periode/:jenis_periode", sasaranPemdaController.FindAllWithPokin)

	//permasalahan rekin
	router.POST("/permasalahan_rekin/create", permasalahanRekinController.Create)
	router.PUT("/permasalahan_rekin/update/:id", permasalahanRekinController.Update)
	router.GET("/permasalahan_rekin/findall/:rekinId", permasalahanRekinController.FindAll)
	router.GET("/permasalahan_rekin/detail/:id", permasalahanRekinController.FindById)
	router.DELETE("/permasalahan_rekin/delete/:id", permasalahanRekinController.Delete)

	//iku
	router.GET("/indikator_utama/periode/:tahun_awal/:tahun_akhir/:jenis_periode", ikuController.FindAll)
	router.GET("/indikator_utama/opd/:kode_opd/:tahun_awal/:tahun_akhir/:jenis_periode", ikuController.FindAllIkuOpd)

	//sasaran opd
	router.GET("/sasaran_opd/findall/:kode_opd/:tahun_awal/:tahun_akhir/:jenis_periode", sasaranOpdController.FindAll)
	router.GET("/sasaran_opd/detail/:id", sasaranOpdController.FindById)
	router.POST("/sasaran_opd/create", sasaranOpdController.Create)
	router.PUT("/sasaran_opd/update/:id", sasaranOpdController.Update)
	router.DELETE("/sasaran_opd/delete/:id", sasaranOpdController.Delete)
	router.GET("/sasaran_opd/pokin/:id_pokin/tahun/:tahun", sasaranOpdController.FindByIdPokin)

	//visi pemda
	router.POST("/visi_pemda/create", visiPemdaController.Create)
	router.PUT("/visi_pemda/update/:id", visiPemdaController.Update)
	router.DELETE("/visi_pemda/delete/:id", visiPemdaController.Delete)
	// router.GET("/visi_pemda/findall/tahunawal/:tahun_awal/tahunakhir/:tahun_akhir/jenisperiode/:jenis_periode", visiPemdaController.FindAll)
	router.GET("/visi_pemda/findall/tahun/:tahun_awal/jenisperiode/:jenis_periode", visiPemdaController.FindAll)
	router.GET("/visi_pemda/detail/:id", visiPemdaController.FindById)

	//misi pemda
	router.POST("/misi_pemda/create", misiPemdaController.Create)
	router.PUT("/misi_pemda/update/:id", misiPemdaController.Update)
	router.DELETE("/misi_pemda/delete/:id", misiPemdaController.Delete)
	// router.GET("/misi_pemda/findall/tahunawal/:tahun_awal/tahunakhir/:tahun_akhir/jenisperiode/:jenis_periode", misiPemdaController.FindAll)
	router.GET("/misi_pemda/findall/tahun/:tahun_awal/jenisperiode/:jenis_periode", misiPemdaController.FindAll)
	router.GET("/misi_pemda/detail/:id", misiPemdaController.FindById)
	router.GET("/misi_pemda/findbyvisi/:id_visi", misiPemdaController.FindByIdVisi)

	//subkegiatan opd
	router.POST("/subkegiatanopd/create", subKegiatanTerpilihController.CreateOpd)
	router.DELETE("/subkegiatanopd/delete/:id", subKegiatanTerpilihController.DeleteOpd)
	router.PUT("/subkegiatanopd/update/:id", subKegiatanTerpilihController.UpdateOpd)
	router.GET("/subkegiatanopd/findall/:kode_opd/:tahun", subKegiatanTerpilihController.FindAllOpd)
	router.GET("/subkegiatanopd/detail/:id", subKegiatanTerpilihController.FindById)
	router.GET("/subkegiatanopd/bidangurusan/:kode_opd", subKegiatanTerpilihController.FindAllSubkegiatanByBidangUrusanOpd)

	//matrix renstra
	router.GET("/matrix_renstra/opd/:kode_opd", matrixRenstraController.GetByKodeSubKegiatan)
	router.GET("/matrix_renstra/indikator/detail/:id", matrixRenstraController.FindIndikatorById)
	router.POST("/matrix_renstra/indikator/create_indikator", matrixRenstraController.CreateIndikator)
	router.PUT("/matrix_renstra/indikator/update_indikator/:id", matrixRenstraController.UpdateIndikator)
	router.DELETE("/matrix_renstra/indikator/delete/:id", matrixRenstraController.DeleteIndikator)

	//cascading opd
	router.GET("/cascading_opd/findall/:kode_opd/:tahun", cascadingOpdController.FindAll)

	//rincian belanja
	router.GET("/rincian_belanja/asn/:pegawai_id/:tahun", rincianBelanjaController.FindRincianBelanjaAsn)
	router.POST("/rincian_belanja/create", rincianBelanjaController.Create)
	router.PUT("/rincian_belanja/update/:renaksiId", rincianBelanjaController.Update)
	router.GET("/rincian_belanja/pegawai/:pegawai_id/:tahun", rincianBelanjaController.LaporanRincianBelanjaPegawai)

	router.GET("/rincian_belanja/laporan", rincianBelanjaController.LaporanRincianBelanjaOpd)

	//kelompok anggaran
	router.POST("/kelompok_anggaran/create", kelompokAnggaranController.Create)
	router.GET("/kelompok_anggaran/findall", kelompokAnggaranController.FindAll)
	router.GET("/kelompok_anggaran/detail/:id", kelompokAnggaranController.FindById)

	//clonning pohon kinerja opd
	router.POST("/pohon_kinerja_opd/clone", pohonKinerjaOpdController.Clone)
	router.GET("/pohon_kinerja_opd/check_pokin/:kode_opd/:tahun", pohonKinerjaOpdController.CheckPokinExistsByTahun)

	//count pokin pemda in opd
	router.GET("/pohon_kinerja_opd/count_pokin_pemda/:kode_opd/:tahun", pohonKinerjaOpdController.CountPokinPemda)

	//Api Internal Consume
	router.GET("/api/pokin_opd/findall/:kode_opd/:tahun", pohonKinerjaOpdController.FindAll)

	return router
}
