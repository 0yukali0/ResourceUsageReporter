package nodemonitor

import (
	"net/http"
)

type APIMethod int

const (
	GET APIMethod = iota
	POST
	PUT
	PATCH
	DELETE
)

var APIMethodNames = map[APIMethod]string{
	GET:    "GET",
	POST:   "POST",
	PUT:    "PUT",
	PATCH:  "PATCH",
	DELETE: "DELETE",
}

func (am APIMethod) String() string {
	return APIMethodNames[am]
}

type route struct {
	Name        string
	Method      string
	API         string
	HandlerFunc http.HandlerFunc
}

type routes []route

var monitorRoutes = routes{
	route{
		Name:        "CPUs_information",
		Method:      GET.String(),
		API:         "/v1/CPUs",
		HandlerFunc: nil,
	},
	route{
		Name:        "CPUs_average_usage",
		Method:      GET.String(),
		API:         "/v1/CPUs/usage",
		HandlerFunc: nil,
	},
	route{
		Name:        "CPU_information",
		Method:      GET.String(),
		API:         "/v1/CPU/:index",
		HandlerFunc: nil,
	},
	route{
		Name:        "CPU_usage",
		Method:      GET.String(),
		API:         "/v1/CPU/:index/usage",
		HandlerFunc: nil,
	},
	route{
		Name:        "memory_information",
		Method:      GET.String(),
		API:         "/v1/memories",
		HandlerFunc: nil,
	},
	route{
		Name:        "memory_usage",
		Method:      GET.String(),
		API:         "/v1/memories/usage",
		HandlerFunc: nil,
	},
	route{
		Name:        "current bandwidth",
		Method:      GET.String(),
		API:         "/v1/net",
		HandlerFunc: nil,
	},
	route{
		Name:        "disks_information",
		Method:      GET.String(),
		API:         "/v1/disks",
		HandlerFunc: nil,
	},
	route{
		Name:        "disk_usage",
		Method:      GET.String(),
		API:         "/v1/disk/:index/usage",
		HandlerFunc: nil,
	},
	route{
		Name:        "GPUs_information",
		Method:      GET.String(),
		API:         "/v1/GPUs",
		HandlerFunc: nil,
	},
	route{
		Name:        "GPU_information",
		Method:      GET.String(),
		API:         "/v1/GPU/:index",
		HandlerFunc: nil,
	},
	route{
		Name:        "GPU_usage",
		Method:      GET.String(),
		API:         "/v1/GPU/:index/usage",
		HandlerFunc: nil,
	},
}
