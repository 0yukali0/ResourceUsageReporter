package webservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shirou/gopsutil/v4/cpu"
)

func ListCPUs(w http.ResponseWriter, r *http.Request) {
	var info []cpu.InfoStat
	var err error
	if info, err = cpu.Info(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "%s", err.Error())
	}
	if err = json.NewEncoder(w).Encode(info); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "%s", err.Error())
	}
}
