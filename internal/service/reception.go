package service

import (
	"errors"
	"fmt"

	"github.com/ST359/pvz-service/internal/api"
	errs "github.com/ST359/pvz-service/internal/app_errors"
	"github.com/ST359/pvz-service/internal/repository"
	"github.com/google/uuid"
)

type ReceptionService struct {
	repo repository.Reception
}

func NewReceptionService(repo repository.Reception) *ReceptionService {
	return &ReceptionService{repo: repo}
}

func (r *ReceptionService) Create(pvzID uuid.UUID) (api.Reception, error) {
	const op = "service.reception.Create"

	id, err := r.GetReceptionInProgress(pvzID)
	if err != nil {
		if !errors.Is(err, errs.ErrNoReceptionsInProgress) {
			return api.Reception{}, fmt.Errorf("%s:%w", op, err)
		}
	}
	if id != uuid.Nil {
		return api.Reception{}, errs.ErrReceptionNotClosed
	}

	rec, err := r.repo.Create(pvzID)
	if err != nil {
		return api.Reception{}, fmt.Errorf("%s:%w", op, err)
	}
	return rec, nil
}
func (r *ReceptionService) AddProduct(pvzID uuid.UUID, product api.ProductType) (api.Product, error) {
	const op = "service.reception.AddProduct"

	recID, err := r.GetReceptionInProgress(pvzID)
	if err != nil {
		if errors.Is(err, errs.ErrNoReceptionsInProgress) {
			return api.Product{}, err
		}
		return api.Product{}, fmt.Errorf("%s:%w", op, err)
	}
	if recID == uuid.Nil {
		return api.Product{}, errs.ErrNoReceptionsInProgress
	}

	prod, err := r.repo.AddProduct(recID, product)
	if err != nil {
		return api.Product{}, fmt.Errorf("%s:%w", op, err)
	}
	return prod, nil
}
func (r *ReceptionService) GetReceptionInProgress(pvzID uuid.UUID) (uuid.UUID, error) {
	const op = "service.reception.GetReceptionInProgress"

	id, err := r.repo.GetReceptionInProgress(pvzID)
	if err != nil {
		if errors.Is(err, errs.ErrNoReceptionsInProgress) {
			return id, err
		}
		return uuid.Nil, fmt.Errorf("%s:%w", op, err)
	}
	return id, nil
}
func (r *ReceptionService) DeleteLastProduct(pvzID uuid.UUID) error {
	const op = "service.reception.DeleteLastProduct"

	recID, err := r.repo.GetReceptionInProgress(pvzID)
	if err != nil {
		if errors.Is(err, errs.ErrNoReceptionsInProgress) {
			return err
		}
		return fmt.Errorf("%s:%w", op, err)
	}
	err = r.repo.DeleteLastProduct(recID)
	if err != nil {
		if errors.Is(err, errs.ErrNoProductsInReception) {
			return err
		}
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}
func (r *ReceptionService) CloseLastReception(pvzID uuid.UUID) (api.Reception, error) {
	const op = "service.reception.AddProduct"

	recID, err := r.GetReceptionInProgress(pvzID)
	if err != nil {
		if errors.Is(err, errs.ErrNoReceptionsInProgress) {
			return api.Reception{}, err
		}
		return api.Reception{}, fmt.Errorf("%s:%w", op, err)
	}
	if recID == uuid.Nil {
		return api.Reception{}, errs.ErrNoReceptionsInProgress
	}

	rec, err := r.repo.CloseLastReception(recID)
	if err != nil {
		return api.Reception{}, fmt.Errorf("%s:%w", op, err)
	}
	return rec, nil
}
