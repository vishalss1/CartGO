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
)

type InventoryService interface {
	GetInventory(ctx context.Context, productID string) (*model.InventoryResponse, error)
	AdjustStock(ctx context.Context, productID string, totalStock int) error
	ReserveStock(ctx context.Context, productID string, quantity int) error
	ReleaseStock(ctx context.Context, productID string, quantity int) error
	CommitStock(ctx context.Context, productID string, quantity int) error
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
			// Create initial record if it doesn't exist
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

func (s *inventoryService) ReserveStock(ctx context.Context, productID string, quantity int) error {
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

	err = s.repo.Reserve(ctx, productID, quantity, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			// This could also mean insufficient stock if the update failed due to the (stock - reserved) check,
			// but we return ErrConflict to trigger a retry which will then hit ErrInsufficientStock.
			return ErrConflict
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}

func (s *inventoryService) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	err = s.repo.Release(ctx, productID, quantity, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return ErrConflict
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}

func (s *inventoryService) CommitStock(ctx context.Context, productID string, quantity int) error {
	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	err = s.repo.Commit(ctx, productID, quantity, inv.Version)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return ErrConflict
		}
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return nil
}
