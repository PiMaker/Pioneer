package commands

import (
    "os"
    "fmt"
    "time"
    "errors"
    "strings"
    "os/exec"
    "os/signal"
    "database/sql"
    
    // SQLite init
	_ "github.com/mattn/go-sqlite3"
)

type Scheduling struct {
    ID int
    StartDate time.Time
    EndDate time.Time
    StartTime time.Time
    EndTime time.Time
    Dynamic bool
    CommandOn string
    CommandOnArgs []string
    CommandOff string
    CommandOffArgs []string
    ExecutedOn bool
    ExecutedOff bool    
}

const MINUTES_TO_RETRY = 2

var db *sql.DB
var scheduledCommands []*Scheduling

func InitScheduling() {
    var err error
    db, err = sql.Open("sqlite3", "./pioneer.db")
    if err != nil {
        panic("Database creation failed!")
    }
    
    saveExec("CREATE TABLE IF NOT EXISTS scheduling (id INTEGER PRIMARY KEY AUTOINCREMENT, startDate DATETIME, endDate DATETIME, startTime DATETIME, endTime DATETIME, dynamic INTEGER, commandOn TEXT, commandOnArgs TEXT, commandOff TEXT, commandOffArgs TEXT, executedOn INTEGER, executedOff INTEGER)")
    
    rows, err2 := db.Query("SELECT id, startDate, endDate, startTime, endTime, dynamic, commandOn, commandOnArgs, commandOff, commandOffArgs, executedOn, executedOff FROM scheduling")
    if err2 != sql.ErrNoRows {
        if err2 != nil {
            panic("Error on querying database!")
        }
        
        scheduledCommands = make([]*Scheduling, 0)
        
        for rows.Next() {
            scheduling := &Scheduling{}
            var cmdOnArgs string
            var cmdOffArgs string
            err3 := rows.Scan(&scheduling.ID, &scheduling.StartDate, &scheduling.EndDate, &scheduling.StartTime, &scheduling.EndTime, &scheduling.Dynamic, &scheduling.CommandOn, &cmdOnArgs, &scheduling.CommandOff, &cmdOffArgs, &scheduling.ExecutedOn, &scheduling.ExecutedOff)
            if err3 != nil {
                panic("Database select/read error!")
            }
            scheduling.CommandOnArgs = strings.Split(cmdOnArgs, " ")
            scheduling.CommandOffArgs = strings.Split(cmdOffArgs, " ")
            scheduledCommands = append(scheduledCommands, scheduling)
        }
        
        rows.Close()
    }
    
    closeScheduling()
    scheduleWorker()
}

func ScheduleCommand(scheduling *Scheduling) (success bool, err error) {
    // TODO: Check for collisions
    if false {
        return false, errors.New("Collision with XYZ!")
    }
    
    saveExec("INSERT INTO scheduling (startDate, endDate, startTime, endTime, dynamic, commandOn, commandOnArgs, commandOff, commandOffArgs, executedOn, executedOff) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1)",
        scheduling.StartDate, scheduling.EndDate, scheduling.StartTime, scheduling.EndTime, scheduling.Dynamic, scheduling.CommandOn, schedulingArgsToString(scheduling.CommandOnArgs), scheduling.CommandOff, schedulingArgsToString(scheduling.CommandOffArgs))
    
    return true, nil
}

func schedulingArgsToString(in []string) string {
    retval := ""
    for i, val := range in {
        if i > 0 {
            retval += " "
        }
        retval += val
    }
    return retval
}

func closeScheduling() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
        for _ = range c {
            fmt.Println(time.Now().String() + " [INFO] Closing database connection...")
            db.Close()
        }
    }()
}

func scheduleWorker() {
    ticker := time.NewTicker(time.Duration(5) * time.Second)
    go func () {
        for {
            <-ticker.C
            now := time.Now()
            for _, sch := range scheduledCommands {
                if sch.StartDate.Before(now) && sch.EndDate.After(now) {
                    // Inside date range
                    timeComparerNow := time.Date(0, 0, 0, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
                    if sch.StartTime.Before(timeComparerNow) && sch.EndTime.Add(time.Duration(MINUTES_TO_RETRY) * time.Minute).After(timeComparerNow) {
                        // In general time range
                        if !sch.ExecutedOn && sch.StartTime.Before(timeComparerNow) && sch.StartTime.Add(time.Duration(MINUTES_TO_RETRY) * time.Minute).After(timeComparerNow) {
                            // Should start
                            execOn(sch)
                            sch.ExecutedOn = true
                            sch.ExecutedOff = false
                            saveExec("UPDATE scheduling s SET s.executedOn=1, s.executedOff=0 WHERE s.id=?", sch.ID)
                        }
                        if !sch.ExecutedOff && sch.EndTime.Before(timeComparerNow) && sch.EndTime.Add(time.Duration(MINUTES_TO_RETRY) * time.Minute).After(timeComparerNow) {
                            // Should stop
                            execOff(sch)
                            sch.ExecutedOff = true
                            sch.ExecutedOn = false
                            saveExec("UPDATE scheduling s SET s.executedOn=0, s.executedOff=1 WHERE s.id=?", sch.ID)
                        }
                        if sch.Dynamic && sch.ExecutedOn {
                            // Should randomize
                        }
                    }
                } else {
                    // Outside date range
                    if now.After(sch.EndDate) {
                        // After EndDate
                        deleteScheduling(sch)
                    }
                }
            }
        }
    }()
}

func deleteScheduling(sch *Scheduling) {
    saveExec("DELETE FROM scheduling WHERE id=?", sch.ID)
    for i, value := range scheduledCommands {
        if value.ID == sch.ID {
            copy(scheduledCommands[i:], scheduledCommands[i+1:])
            scheduledCommands[len(scheduledCommands)-1] = nil
            scheduledCommands = scheduledCommands[:len(scheduledCommands)-1]
            break
        }
    }
}

func execOn(sch *Scheduling) {
    fmt.Println("Executing something on!")
    exec.Command(sch.CommandOn, sch.CommandOnArgs...).Start()
}

func execOff(sch *Scheduling) {
    exec.Command(sch.CommandOff, sch.CommandOffArgs...).Start()    
}

func saveExec(cmd string, args ...interface{}) sql.Result {
    retval, err := db.Exec(cmd, args...)
    if err != nil {
        panic(err)
    }
    return retval
}