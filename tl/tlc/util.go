package tlc

import (
	"github.com/PROger4ever/telegramapi/tl/tlschema"
)

func IDConstName(comb *tlschema.Comb) string {
	return "Tag" + comb.CombName.GoName()
}
