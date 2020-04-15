package imports

import (
	"fmt"
	"strconv"

	"cloudeng.io/errors"
)

func Example3() {
	fmt.Println(strconv.Itoa(2))
}

func Example4() error {
	return errors.New("eg")
}
