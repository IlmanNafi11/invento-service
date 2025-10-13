package domain

type ProjectUpdateMetadataRequest struct {
	NamaProject string `json:"nama_project" validate:"required,min=3,max=255"`
	Kategori    string `json:"kategori" validate:"required,oneof=website mobile iot machine_learning deep_learning"`
	Semester    int    `json:"semester" validate:"required,min=1,max=8"`
}
