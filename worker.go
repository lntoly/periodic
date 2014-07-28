package main

import (
    "log"
    "container/list"
    "github.com/docker/libchan/unix"
    "github.com/docker/libchan/data"
)

type Worker struct {
    jobs *list.List
    conn *unix.UnixConn
    sched *Sched
}


func NewWorker(sched *Sched, conn *unix.UnixConn) (worker *Worker) {
    worker = new(Worker)
    worker.conn = conn
    worker.jobs = list.New()
    worker.sched = sched
    return
}


func (worker *Worker) HandeNewConnection() {
    if err := worker.conn.Send(data.Empty().Set("type", "connection").Bytes(), nil); err != nil {
        worker.sched.die_worker <- worker
        log.Printf("Error: %s\n", err.Error())
        return
    }
    go worker.Handle()
}


func (worker *Worker) HandleDo(job string) {
    worker.jobs.PushBack(job)
    if err := worker.conn.Send(data.Empty().Set("job", job).Bytes(), nil); err != nil {
        worker.sched.die_worker <- worker
        log.Printf("Error: %s\n", err.Error())
        return
    }
    go worker.Handle()
}


func (worker *Worker) HandleDone(job string) {
    worker.sched.Done(job)
    for e := worker.jobs.Front(); e != nil; e = e.Next() {
        if e.Value.(string) == job {
            worker.jobs.Remove(e)
            break
        }
    }
    go worker.Handle()
}


func (worker *Worker) HandleFail(job string) {
    worker.sched.Fail(job)
    for e := worker.jobs.Front(); e != nil; e = e.Next() {
        if e.Value.(string) == job {
            worker.jobs.Remove(e)
            break
        }
    }
    go worker.Handle()
}


func (worker *Worker) HandleWaitForJob() {
    if err := worker.conn.Send(data.Empty().Set("msg", "wait_for_job").Bytes(), nil); err != nil {
        worker.sched.die_worker <- worker
        log.Printf("Error: %s\n", err.Error())
        return
    }
    go worker.Handle()
}


func (worker *Worker) HandleNoJob() {
    if err := worker.conn.Send(data.Empty().Set("msg", "no_job").Bytes(), nil); err != nil {
        worker.sched.die_worker <- worker
        log.Printf("Error: %s\n", err.Error())
        return
    }
    go worker.Handle()
}


func (worker *Worker) Handle() {
    var payload []byte
    var err error
    var conn = worker.conn
    payload, _, err = conn.Receive()
    if err != nil {
        log.Printf("Error: %s\n", err.Error())
        worker.sched.die_worker <- worker
        return
    }
    msg := data.Message(string(payload));
    cmd := msg.Get("cmd")
    switch cmd[0] {
    case "ask":
        worker.sched.ask_worker <- worker
        break
    case "done":
        go worker.HandleDone(msg.Get("job")[0])
        break
    case "fail":
        go worker.HandleFail(msg.Get("job")[0])
        break
    case "sleep":
        if err = conn.Send(data.Empty().Set("message", "nop").Bytes(), nil); err != nil {
            log.Printf("Error: %s\n", err.Error())
        }
        go worker.Handle()
        break
    case "ping":
        if err = conn.Send(data.Empty().Set("message", "pong").Bytes(), nil); err != nil {
            log.Printf("Error: %s\n", err.Error())
        }
        go worker.Handle()
        break
    default:
        if err = conn.Send(data.Empty().Set("error", "command: " + cmd[0] + " unknown").Bytes(), nil); err != nil {
            log.Printf("Error: %s\n", err.Error())
            worker.sched.die_worker <- worker
        }
        go worker.Handle()
        break
    }
}


func (worker *Worker) Close() {
    worker.conn.Close()
    for e := worker.jobs.Front(); e != nil; e = e.Next() {
        worker.sched.Fail(e.Value.(string))
    }
}
