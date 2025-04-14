package service

import (
	"fmt"

	"github.com/ST359/pvz-service/internal/api"
	"github.com/ST359/pvz-service/internal/repository"
)

type PVZService struct {
	repo repository.PVZ
}

func NewPVZService(repo repository.PVZ) *PVZService {
	return &PVZService{repo: repo}
}
func (p *PVZService) Create(pvz api.PVZ) (api.PVZ, error) {
	const op = "service.pvz.Create"

	res, err := p.repo.Create(pvz)
	if err != nil {
		return api.PVZ{}, fmt.Errorf("%s:%w", op, err)
	}
	return res, nil
}
func (p *PVZService) GetByDate(params api.GetPvzParams) ([]api.PVZInfo, error) {
	const op = "service.pvz.GetByDate"

	resp, err := p.repo.GetByDate(params)
	if err != nil {
		return []api.PVZInfo{}, fmt.Errorf("%s:%w", op, err)
	}
	return resp, nil
}
