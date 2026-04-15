package main
import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)
type Req struct {
	ID uuid.UUID `validate:"required"`
}
func main() {
	v := validator.New()
	req := Req{ID: uuid.MustParse("d2f7e025-aaaa-bbbb-cccc-111122223333")}
	err := v.Struct(req)
	fmt.Println("Err:", err)
}
