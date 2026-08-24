// modules/resolution/dto/resolution_dto.go
package dto

type ResolutionResponse struct {
	Handle  string `json:"handle"`
	Chain   string `json:"chain"`
	Address string `json:"address"`
}
