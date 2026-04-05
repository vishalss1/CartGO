package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vishalss1/CartGO/services/inventory-service/internal/model"
	"github.com/vishalss1/CartGO/services/inventory-service/internal/repository"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock available")
	ErrNotFound          = errors.New("inventory record not found")
	ErrConflict          = errors.New("concurrent update conflict, please retry")
	ErrInternal          = errors.New("internal server error")
	ErrInvalidState      = errors.New("invalid reservation state for this operation")
)

type InventoryService interface {
	GetInventory(ctx context.Context, productID string) (*model.InventoryResponse, error)
	AdjustStock(ctx context.Context, productID string, totalStock int) error
	ReserveStock(ctx context.Context, productID string, orderID string, quantity int) error
	ReleaseStock(ctx context.Context, orderID string) error
	CommitStock(ctx context.Context, orderID string) error
}

type inventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) GetInventory(ctx context.Context, productID string) (*model.InventoryResponse, error) {
	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return inv.ToResponse(), nil
}

func (s *inventoryService) AdjustStock(ctx context.Context, productID string, totalStock int) error {
	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_, err = s.repo.Upsert(ctx, &model.Inventory{
				ProductID: productID,
				Stock:     totalStock,
				Reserved:  0,
			})
			if err != nil {
				return fmt.Errorf("%w: %v", ErrInternal, err)
			}
			return nil
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	err = s.repo.UpdateStock(ctx, productID, totalStock, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return ErrConflict
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}

func (s *inventoryService) ReserveStock(ctx context.Context, productID string, orderID string, quantity int) error {
	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	if (inv.Stock - inv.Reserved) < quantity {
		return ErrInsufficientStock
	}

	err = s.repo.Reserve(ctx, productID, orderID, quantity, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrIdempotent) {
			return nil // Already reserved, success
		}
		if errors.Is(err, repository.ErrConflict) {
			return ErrConflict
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}

func (s *inventoryService) ReleaseStock(ctx context.Context, orderID string) error {
	res, err := s.repo.GetReservation(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	inv, err := s.repo.GetByProductID(ctx, res.ProductID.String())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	err = s.repo.Release(ctx, orderID, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrIdempotent) {
			return nil
		}
		if errors.Is(err, repository.ErrConflict) {
			return ErrConflict
		}
		if errors.Is(err, repository.ErrInvalidState) {
			return ErrInvalidState
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}

func (s *inventoryService) CommitStock(ctx context.Context, orderID string) error {
	res, err := s.repo.GetReservation(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	inv, err := s.repo.GetByProductID(ctx, res.ProductID.String())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	err = s.repo.Commit(ctx, orderID, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrIdempotent) {
			return nil
		}
		if errors.Is(err, repository.ErrConflict) {
			return ErrConflict
		}
		if errors.Is(err, repository.ErrInvalidState) {
			return ErrInvalidState
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}
