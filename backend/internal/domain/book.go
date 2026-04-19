package domain

import (
	"database/sql"
	"time"
)

type Book struct {
	ID            int             `db:"id"`
	Title         string          `db:"title"`
	Author        string          `db:"author"`
	CreatedBy     int             `db:"created_by"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
	Description   sql.NullString  `db:"description"`
	ISBN          sql.NullString  `db:"isbn"`
	PublishedYear sql.NullInt32   `db:"published_year"`
	AverageRating sql.NullFloat64 `db:"average_rating"`
}

type BookResponse struct {
	ID            int         `json:"id"`
	Title         string      `json:"title"`
	Author        string      `json:"author"`
	CreatedBy     UserSummary `json:"created_by"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Description   *string     `json:"description"`
	ISBN          *string     `json:"isbn"`
	PublishedYear *int        `json:"published_year"`
	AverageRating *float64    `json:"average_rating"`
}

type CreateBookRequest struct {
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int    `json:"published_year"`
}

type UpdateBookRequest struct {
	Title         *string `json:"title"`
	Author        *string `json:"author"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int    `json:"published_year"`
}

type BookFilter struct {
	Search *string `json:"search"`
	Sort   *string `json:"sort"`
	Order  *string `json:"order"`
	Page   *string `json:"page"`
	Limit  *string `json:"limit"`
}

func (b *Book) ToResponse(creator *UserSummary) *BookResponse {

	var desc *string
	desc = nil

	if b.Description.Valid {
		desc = &b.Description.String
	}

	var isbn *string
	isbn = nil

	if b.ISBN.Valid {
		isbn = &b.ISBN.String
	}

	var publishedYear *int
	publishedYear = nil

	if b.PublishedYear.Valid {
		tmpPubY := int(b.PublishedYear.Int32)
		publishedYear = &tmpPubY
	}

	var averageRating *float64
	averageRating = nil

	if b.PublishedYear.Valid {
		averageRating = &b.AverageRating.Float64
	}

	return &BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		CreatedBy:     *creator,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
		Description:   desc,
		ISBN:          isbn,
		PublishedYear: publishedYear,
		AverageRating: averageRating,
	}
}
