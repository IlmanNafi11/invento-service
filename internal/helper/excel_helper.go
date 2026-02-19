package helper

import (
	"github.com/xuri/excelize/v2"
)

type ExcelHelper struct{}

func NewExcelHelper() *ExcelHelper {
	return &ExcelHelper{}
}

func (h *ExcelHelper) GenerateImportTemplate() (*excelize.File, error) {
	f := excelize.NewFile()

	if err := h.createDataSheet(f); err != nil {
		return nil, err
	}

	if err := h.createGuideSheet(f); err != nil {
		return nil, err
	}

	f.SetActiveSheet(0)
	return f, nil
}

func (h *ExcelHelper) createDataSheet(f *excelize.File) error {
	sheet := "Data Import"
	idx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	f.SetActiveSheet(idx)

	if err := f.DeleteSheet("Sheet1"); err != nil {
		return err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D6EAF8"}, Pattern: 1},
	})
	if err != nil {
		return err
	}

	headers := []string{"Email", "Nama", "Password", "Jenis Kelamin", "Role"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	examples := [][]interface{}{
		{"dosen1@polije.ac.id", "Dr. Ahmad Fauzi", "", "Laki-laki", "Dosen"},
		{"mahasiswa1@student.polije.ac.id", "Siti Nurhaliza", "password123", "Perempuan", "Mahasiswa"},
		{"admin@polije.ac.id", "Budi Santoso", "", "Laki-laki", ""},
	}
	for rowIdx, row := range examples {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	widths := map[string]float64{"A": 30, "B": 25, "C": 20, "D": 18, "E": 15}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}

	dvGender := excelize.NewDataValidation(true)
	dvGender.Sqref = "D2:D1000"
	dvGender.SetDropList([]string{"Laki-laki", "Perempuan"})
	if err := f.AddDataValidation(sheet, dvGender); err != nil {
		return err
	}

	dvRole := excelize.NewDataValidation(true)
	dvRole.Sqref = "E2:E1000"
	dvRole.SetDropList([]string{"Admin", "Dosen", "Mahasiswa"})
	if err := f.AddDataValidation(sheet, dvRole); err != nil {
		return err
	}

	return nil
}

func (h *ExcelHelper) createGuideSheet(f *excelize.File) error {
	sheet := "Panduan"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	if err != nil {
		return err
	}

	f.SetCellValue(sheet, "A1", "Panduan Pengisian Template Import User")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	guideHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D6EAF8"}, Pattern: 1},
	})
	if err != nil {
		return err
	}

	guideHeaders := []string{"Kolom", "Wajib", "Keterangan"}
	for i, h := range guideHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, guideHeaderStyle)
	}

	fields := [][]string{
		{"Email", "Ya", "Alamat email valid. Mahasiswa harus @student.polije.ac.id"},
		{"Nama", "Ya", "Nama lengkap pengguna (2-100 karakter)"},
		{"Password", "Tidak", "Kosongkan untuk auto-generate. Min 8 karakter jika diisi"},
		{"Jenis Kelamin", "Tidak", "Pilih: Laki-laki atau Perempuan"},
		{"Role", "Tidak", "Pilih: Admin, Dosen, atau Mahasiswa. Kosongkan untuk default"},
	}
	for rowIdx, row := range fields {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+4)
			f.SetCellValue(sheet, cell, val)
		}
	}

	boldStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return err
	}

	f.SetCellValue(sheet, "A10", "Catatan:")
	f.SetCellStyle(sheet, "A10", "A10", boldStyle)
	f.SetCellValue(sheet, "A11", "1. Baris dengan Email atau Nama kosong akan dilewati")
	f.SetCellValue(sheet, "A12", "2. Email duplikat (sudah terdaftar atau duplikat dalam file) akan dilewati")
	f.SetCellValue(sheet, "A13", "3. Mahasiswa dengan email selain @student.polije.ac.id akan dilewati")
	f.SetCellValue(sheet, "A14", "4. Hapus baris contoh sebelum mengimpor")

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 8)
	f.SetColWidth(sheet, "C", "C", 60)

	return nil
}
