package model

import (
	"fmt"
	"time"
)

// Filename pattern for orders
const orderFilename = "order_%d.md"

// Markdown template for the receipt
// You can populate it with ID, created time, drink name, and amount
const markdownTemplate = `# Order: %d

| Created At       | Drink ID | Amount |
|-----------------|----------|--------|
| %s | %d        | %d     |
`

type Order struct {
	Base
	Amount uint64 `json:"amount"`
	// Relationships
	DrinkID uint  `json:"drink_id" gorm:"not null"`
	Drink   Drink `json:"drink"`
}

// ToMarkdown returns a formatted markdown receipt for the order
func (o *Order) ToMarkdown() string {
	return fmt.Sprintf(
		markdownTemplate,
		o.ID,
		o.CreatedAt.Format(time.RFC1123),
		o.DrinkID,
		o.Amount,
	)
}

// GetFilename returns the filename for the order's receipt
func (o *Order) GetFilename() string {
	return fmt.Sprintf(orderFilename, o.ID)
}
