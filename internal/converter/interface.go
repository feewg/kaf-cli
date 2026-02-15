package converter

import "github.com/feewg/kaf-cli/internal/model"

type Converter interface {
	Build(book model.Book) error
}
