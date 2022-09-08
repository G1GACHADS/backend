package backend

import (
	"context"
)

type WarehouseSearchItem struct {
	Attributes    Warehouse `json:"attributes"`
	Relationships struct {
		Address    Address  `json:"address"`
		Categories []string `json:"categories"`
	} `json:"relationships"`
}

type WarehouseSearchOutput struct {
	TotalItems int                   `json:"total_items"`
	Items      []WarehouseSearchItem `json:"items"`
}

func (b *backend) WarehouseSearch(ctx context.Context, searchQuery string, limit int, priceAscending bool) (WarehouseSearchOutput, error) {
	sql := b.makeSearchWarehouseQuery(priceAscending)

	rows, err := b.clients.DB.Query(ctx, sql, searchQuery, limit)
	if err != nil {
		return WarehouseSearchOutput{}, err
	}
	defer rows.Close()

	var warehouses []WarehouseSearchItem

	for rows.Next() {
		var row WarehouseSearchItem
		err := rows.Scan(
			&row.Attributes.ID,
			&row.Attributes.AddressID,
			&row.Attributes.Name,
			&row.Attributes.ImageURL,
			&row.Attributes.Description,
			&row.Attributes.BasePrice,
			&row.Attributes.Email,
			&row.Attributes.PhoneNumber,
			&row.Attributes.CreatedAt,
			&row.Relationships.Address.ID,
			&row.Relationships.Address.Province,
			&row.Relationships.Address.City,
			&row.Relationships.Address.StreetName,
			&row.Relationships.Address.ZipCode,
			&row.Relationships.Categories,
		)
		if err != nil {
			return WarehouseSearchOutput{}, err
		}
		warehouses = append(warehouses, row)
	}

	return WarehouseSearchOutput{
		TotalItems: len(warehouses),
		Items:      warehouses,
	}, nil
}

func (b *backend) makeSearchWarehouseQuery(priceAscending bool) string {
	mainQuery := `SELECT
		w.*,
		a.*,
		array_agg(c.name)
	FROM warehouses w
		LEFT JOIN addresses AS a ON w.address_id = a.id
		LEFT JOIN warehouse_categories AS wc ON w.id = wc.warehouse_id
		LEFT JOIN categories AS c ON wc.category_id = c.id
	WHERE
		w.name ILIKE '%' || $1 || '%' OR
		w.description ILIKE '%' || $1 || '%' OR
		w.email ILIKE '%' || $1 || '%' OR
		a.province ILIKE '%' || $1 || '%' OR
		a.city ILIKE '%' || $1 || '%' OR
		a.street_name ILIKE '%' || $1 || '%' OR
		c.name ILIKE '%' || $1 || '%'
	GROUP BY w.id,
		a.id
	`

	if priceAscending {
		mainQuery += "ORDER BY w.base_price ASC"
	} else {
		mainQuery += "ORDER BY w.base_price DESC"
	}

	mainQuery += " LIMIT $2"

	return mainQuery
}
