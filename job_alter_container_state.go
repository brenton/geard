package geard

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

type startedContainerStateJobRequest struct {
	JobResponse
	jobRequest
	GearId Identifier
	UserId string
}

func (j *startedContainerStateJobRequest) Execute() {
	status, err := StartAndEnableUnit(SystemdConnection(), j.GearId.UnitNameFor(), j.GearId.UnitPathFor(), "fail")

	switch {
	case IsNoSuchUnit(err):
		j.Failure(ErrGearNotFound)
		return
	case err != nil:
		log.Printf("job_alter_container_state: Gear did not start: %+v", err)
		j.Failure(ErrGearStartFailed)
		return
	case status != "done":
		log.Printf("job_alter_container_state: Unit did not return 'done': %v", err)
		j.Failure(ErrGearStartFailed)
		return
	}

	w := j.SuccessWithWrite(JobResponseAccepted, true)
	fmt.Fprintf(w, "Gear %s starting\n", j.GearId)
	// stdout, err := ProcessLogsFor(j.GearId)
	// if err != nil {
	// 	stdout = emptyReader
	// 	log.Printf("job_alter_container_state: Could not fetch journal logs: %+v", err)
	// }
	// ioerr := make(chan error)

	// go func() {
	// 	_, err := io.Copy(w, stdout)
	// 	ioerr <- err
	// }()

	// time.Sleep(1 * time.Second)
	// stdout.Close()

	// select {
	// case erri := <-ioerr:
	// 	log.Printf("job_alter_container_state: Error from IO on wait: %+v", erri)
	// case <-time.After(15 * time.Second):
	// 	log.Printf("job_alter_container_state: Timeout waiting for write to complete")
	// }
}

type stoppedContainerStateJobRequest struct {
	JobResponse
	jobRequest
	GearId Identifier
	UserId string
}

func (j *stoppedContainerStateJobRequest) Execute() {
	w := j.SuccessWithWrite(JobResponseAccepted, true)

	unitName := j.GearId.UnitNameFor()
	done := make(chan time.Time)

	ioerr := make(chan error)
	go func() {
		ioerr <- WriteLogsTo(w, unitName, done)
	}()

	joberr := make(chan error)
	go func() {
		status, err := SystemdConnection().StopUnit(unitName, "fail")
		if err == nil && status != "done" {
			err = errors.New(fmt.Sprintf("Job status 'done' != %s", status))
		}
		joberr <- err
	}()

	var err error
	select {
	case err = <-ioerr:
		log.Printf("job_alter_container_state: Client hung up")
	case err = <-joberr:
		log.Printf("job_alter_container_state: Stop job done")
	case <-time.After(15 * time.Second):
		log.Printf("job_alter_container_state: Timeout waiting for stop completion")
	}
	close(done)

	switch {
	case IsNoSuchUnit(err):
		if _, err := os.Stat(j.GearId.UnitPathFor()); err == nil {
			fmt.Fprintf(w, "Gear %s is stopped\n", j.GearId)
		} else {
			fmt.Fprintf(w, "No such gear %s\n", j.GearId)
		}
	case err != nil:
		fmt.Fprintf(w, "Could not start gear: %s\n", err.Error())
	default:
		fmt.Fprintf(w, "Gear %s is stopped\n", j.GearId)
	}
}
