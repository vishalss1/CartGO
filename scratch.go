package main
import (
	"fmt"
	"hash/fnv"
)
func main() {
	h := fnv.New32a()
	h.Write([]byte("00000000-0000-0000-0000-000000000000"))
	fmt.Println(h.Sum32()%100)
}
