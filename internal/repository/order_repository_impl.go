package repository

import (
	"database/sql"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/driver"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
)

type OrderRepositoryImpl struct {
	*driver.DB
}

func NewOrderRepository(db *driver.DB) OrderRepository {
	return &OrderRepositoryImpl{
		DB: db,
	}
}

func (r *OrderRepositoryImpl) Insert(order *entity.Order) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `INSERT INTO orders (id, price, status, user_id, version)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *`

	newOrder := new(entity.Order)
	if err := r.SQL.QueryRowContext(
		ctx,
		stmt,
		order.ID,
		order.Price,
		order.Status,
		order.UserID,
		order.Version,
	).Scan(
		&newOrder.ID,
		&newOrder.Price,
		&newOrder.Status,
		&newOrder.UserID,
		&newOrder.Version,
	); err != nil {
		return nil, &common.Error{Op: "OrderRepositoryImpl", Err: err}
	}

	return newOrder, nil
}

func (r *OrderRepositoryImpl) FindOne(orderID int64) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT * FROM orders
	WHERE id = $1`

	order := new(entity.Order)
	if err := r.SQL.QueryRowContext(ctx, stmt, orderID).Scan(
		&order.ID,
		&order.Price,
		&order.Status,
		&order.UserID,
		&order.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ENOTFOUND,
				Op:      "OrderRepositoryImpl",
				Message: "Order Not Found",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "OrderRepositoryImpl", Err: err}
	}

	return order, nil
}

func (r *OrderRepositoryImpl) Update(order *entity.Order) (*entity.Order, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `UPDATE orders
	SET status = $1, version = $2
	WHERE id = $3 AND version = $4
	RETURNING *`

	updatedOrder := new(entity.Order)
	if err := r.SQL.QueryRowContext(
		ctx,
		stmt,
		order.Status,
		order.Version,
		order.ID,
		order.Version-1,
	).Scan(
		&updatedOrder.ID,
		&updatedOrder.Price,
		&updatedOrder.Status,
		&updatedOrder.UserID,
		&updatedOrder.Version,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, &common.Error{
				Code:    common.ECONCLICT,
				Op:      "OrderRepositoryImpl.Update",
				Message: "Order is out of sync",
				Err:     err,
			}
		}
		return nil, &common.Error{Op: "OrderRepositoryImpl.Update", Err: err}
	}

	return updatedOrder, nil
}
