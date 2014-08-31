package sched


type Job struct {
    Id      int64  `json:"job_id"`
    Name    string `json:"name"`
    Func    string `json:"func"`
    Args    string `json:"workload"`
    Timeout int64  `json:"timeout"`
    SchedAt int64  `json:"sched_at"`
    RunAt   int64  `json:"run_at"`
    Status  string `json:"status"`
}


const (
    JOB_STATUS_READY = "ready"
    JOB_STATUS_PROC  = "doing"
)


type Storer interface {
    Save(Job) error
    Delete(jobId int64) error
    Get(jobId int64) (Job, error)
    Count() (int64, error)
    GetOne([]byte) (Job, error)
    NewIterator([]byte, []byte) JobIterator
}


type Iterator interface {
    Next() bool
}


type JobIterator interface {
    Iterator
    Value() Job
    Error() error
    Close()
}