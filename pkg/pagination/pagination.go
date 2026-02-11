package pagination

import (
	"net/http"
	"strconv"
)

type Params struct {
	Limit  int32
	Offset int32
}

func Parse(r *http.Request) Params {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return Params{Limit: int32(limit), Offset: int32(offset)}
}
