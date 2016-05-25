package commands

import (
    "os"
    "fmt"
    "time"
    "errors"
    "strings"
    "os/exec"
    "os/signal"
    "math/rand"
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

const SECONDS_TIMING = 6
const RANDOM_RANGE = 3

var db *sql.DB
var scheduledCommands []*Scheduling

func InitScheduling() {
    // DEBUG
    os.Remove("./pioneer.db")
    
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

func ScheduleCommand(scheduling Scheduling) error {
    fmt.Println(time.Now().String() + " [SCHED] Starting to schedule something...")
    
    scheduling.EndDate = time.Date(scheduling.EndDate.Year(), scheduling.EndDate.Month(), scheduling.EndDate.Day(), 23, 59, 59, 999999999, time.Local)
    
    for _, sch := range scheduledCommands {
        if (scheduling.StartDate.Before(sch.StartDate) && scheduling.EndDate.After(sch.StartDate)) ||
           (scheduling.EndDate.After(sch.EndDate) && scheduling.StartDate.Before(sch.EndDate)) ||
           (scheduling.StartDate.Before(sch.StartDate) && scheduling.EndDate.After(sch.EndDate)) {
               // Gotta check time
                if (scheduling.StartTime.Before(sch.StartTime) && scheduling.EndTime.After(sch.StartTime)) ||
                   (scheduling.EndTime.After(sch.EndTime) && scheduling.StartTime.Before(sch.EndTime)) ||
                   (scheduling.StartTime.Before(sch.StartTime) && scheduling.EndTime.After(sch.EndTime)) {
                    // Oh noes!
                    return errors.New("This entry would collide with a different scheduled command! Scheduling not commited, please try again.")
                }
           }
    }
    
    scheduledCommands = append(scheduledCommands, &scheduling)
    saveExec("INSERT INTO scheduling (startDate, endDate, startTime, endTime, dynamic, commandOn, commandOnArgs, commandOff, commandOffArgs, executedOn, executedOff) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1)",
        scheduling.StartDate, scheduling.EndDate, scheduling.StartTime, scheduling.EndTime, scheduling.Dynamic, scheduling.CommandOn, schedulingArgsToString(scheduling.CommandOnArgs), scheduling.CommandOff, schedulingArgsToString(scheduling.CommandOffArgs))
    
    fmt.Println(time.Now().String() + " [SCHED] Scheduled something.")
    return nil
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
    ticker := time.NewTicker((time.Duration(SECONDS_TIMING) / 3) * time.Second)
    go func () {
        for {
            <-ticker.C
            now := time.Now()
            for _, sch := range scheduledCommands {
                if sch.StartDate.Before(now) && sch.EndDate.After(now) {
                    // Inside date range
                    timeComparerNow := time.Date(0, 0, 0, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
                    if sch.StartTime.Before(timeComparerNow) && sch.EndTime.Add(time.Duration(SECONDS_TIMING) * time.Second).After(timeComparerNow) {
                        // In general time range
                        if !sch.ExecutedOn && sch.StartTime.Before(timeComparerNow) && sch.StartTime.Add(time.Duration(SECONDS_TIMING) * time.Second).After(timeComparerNow) {
                            // Should start
                            execOn(sch)
                            sch.ExecutedOn = true
                            sch.ExecutedOff = false
                            saveExec("UPDATE scheduling SET executedOn=1, executedOff=0 WHERE id=?", sch.ID)
                        }
                        if !sch.ExecutedOff && sch.EndTime.Before(timeComparerNow) && sch.EndTime.Add(time.Duration(SECONDS_TIMING) * time.Second).After(timeComparerNow) {
                            // Should stop
                            execOff(sch)
                            sch.ExecutedOff = true
                            sch.ExecutedOn = false
                            saveExec("UPDATE scheduling SET executedOn=0, executedOff=1 WHERE id=?", sch.ID)
                        }
                        if sch.Dynamic {
                            // Should randomize
                            seed := rand.Int31n(RANDOM_RANGE)
                            if seed == (RANDOM_RANGE/2) {
                                if !sch.ExecutedOff && !sch.ExecutedOn {
                                    sch.ExecutedOn = true
                                    execOn(sch)
                                } else {
                                    sch.ExecutedOn = false
                                    execOff(sch)
                                }
                            }
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
    fmt.Println(time.Now().String() + " [SCHED] Deleting a scheduling...")
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
    fmt.Println(time.Now().String() + " [SCHED] Activating a scheduling...")
    _, _ = exec.Command(sch.CommandOn, sch.CommandOnArgs...).CombinedOutput()
}

func execOff(sch *Scheduling) {
    fmt.Println(time.Now().String() + " [SCHED] Deactivating a scheduling...")
    _, _ = exec.Command(sch.CommandOff, sch.CommandOffArgs...).CombinedOutput()
}

func saveExec(cmd string, args ...interface{}) sql.Result {
    retval, err := db.Exec(cmd, args...)
    if err != nil {
        panic(err)
    }
    return retval
}