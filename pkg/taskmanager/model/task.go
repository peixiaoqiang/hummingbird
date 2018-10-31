package model

type Task struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	DriverCPU int64  `json:"driver_cpu,omitempty"`
	// DriverMem represents driver memory which unit is Mi
	DriverMem   int64 `json:"driver_mem,omitempty"`
	ExecutorCPU int64 `json:"executor_cpu,omitempty"`
	// ExecutorMem represents executor memory which unit is Mi
	ExecutorMem  int64    `json:"executor_mem,omitempty"`
	ExecutorNum  int64    `json:"executor_num,omitempty"`
	Conf         []string `json:"conf,omitempty"`
	MasterIP     string   `json:"master_ip,omitempty"`
	Image        string   `json:"image,omitempty"`
	TaskArgs     []string `json:"task_args,omitempty"`
	Class        string   `json:"class,omitempty"`
	StartTime    string   `json:"start_time,omitempty"`
	EndTime      string   `json:"end_time,omitempty"`
	CreateTime   string   `json:"create_time,omitempty"`
	FailedReason string   `json:"failed_reason,omitempty"`
}

func (t *Task) TotalResource() (int64, int64) {
	return t.DriverCPU + t.ExecutorCPU*t.ExecutorNum, t.DriverMem + t.ExecutorMem*t.ExecutorNum
}
