package dto

type IDParam struct {
	ID string `uri:"id" validate:"required,uuid"`
}
