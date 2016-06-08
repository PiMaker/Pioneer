package commands

import (
    "os"
    "fmt"
    "time"
    "errors"
	"strconv"
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
    CommandID int
    ExecutedOn bool
    ExecutedOff bool    
}

const SECONDS_TIMING = 6
const RANDOM_RANGE = 3

var db *sql.DB
var scheduledCommands []*Scheduling

func InitScheduling() {
    // DEBUG
    //os.Remove("./pioneer.db")
    
    var err error
    db, err = sql.Open("sqlite3", "./pioneer.db")
    if err != nil {
        panic("Database creation failed!")
    }
    
    saveExec("CREATE TABLE IF NOT EXISTS scheduling (id INTEGER PRIMARY KEY AUTOINCREMENT, startDate DATETIME, endDate DATETIME, startTime DATETIME, endTime DATETIME, dynamic INTEGER, commandId INTEGER, executedOn INTEGER, executedOff INTEGER)")
    
    rows, err2 := db.Query("SELECT id, startDate, endDate, startTime, endTime, dynamic, commandId, executedOn, executedOff FROM scheduling")
    if err2 != sql.ErrNoRows {
        if err2 != nil {
            panic("Error on querying database!")
        }
        
        scheduledCommands = make([]*Scheduling, 0)
        
        for rows.Next() {
            scheduling := &Scheduling{}
            err3 := rows.Scan(&scheduling.ID, &scheduling.StartDate, &scheduling.EndDate, &scheduling.StartTime, &scheduling.EndTime, &scheduling.Dynamic, &scheduling.CommandID, &scheduling.ExecutedOn, &scheduling.ExecutedOff)
            if err3 != nil {
                panic("Database select/read error!")
            }
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
    scheduling.ExecutedOn = false
    scheduling.ExecutedOff = true

    fmt.Println(scheduling.StartTime)
    
    for _, sch := range scheduledCommands {
        if (sch.StartDate.Before(scheduling.EndDate) || dateEquals(sch.StartDate, scheduling.EndDate)) &&
           (sch.EndDate.After(scheduling.StartDate) || dateEquals(sch.EndDate, scheduling.StartDate)) {
            // Gotta check time
            if (sch.StartTime.Before(scheduling.EndTime) || timeEquals(sch.StartTime, scheduling.EndTime)) &&
               (sch.EndTime.After(scheduling.StartTime) || timeEquals(sch.EndTime, scheduling.StartTime)) {
                // Oh noes!
                return errors.New("ERROR: This entry would collide with a different scheduled command (#" + strconv.Itoa(sch.ID) + ")! Scheduling not commited, please try again.")
            }
        }
    }
    
    id, _ := saveExec("INSERT INTO scheduling (startDate, endDate, startTime, endTime, dynamic, commandId, executedOn, executedOff) VALUES (?, ?, ?, ?, ?, ?, 0, 1)",
        scheduling.StartDate, scheduling.EndDate, scheduling.StartTime, scheduling.EndTime, scheduling.Dynamic, scheduling.CommandID).LastInsertId()
    scheduling.ID = int(id)
    scheduledCommands = append(scheduledCommands, &scheduling)

    fmt.Println(scheduling.StartTime)
    
    fmt.Println(time.Now().String() + " [SCHED] Scheduled something.")
    return nil
}

func dateEquals(one, two time.Time) bool {
    return one.Year() == two.Year() && one.Month() == two.Month() && one.Day() == two.Day()
}

func timeEquals(one, two time.Time) bool {
    return one.Hour() == two.Hour() && one.Minute() == two.Minute() && one.Second() == two.Second()
}

func GetSchedulings() []*Scheduling {
    return scheduledCommands
}

func GetSchedulingById(id int) *Scheduling {
    for _, sch := range scheduledCommands {
        if sch.ID == id {
            return sch
        }
    }

    return nil
}

func CancelScheduling(sch *Scheduling) {
    deleteScheduling(sch)
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
    cmd := CommandsAvailable[sch.CommandID]
    cmd.ExecutableCommand.Execute("on")
}

func execOff(sch *Scheduling) {
    fmt.Println(time.Now().String() + " [SCHED] Deactivating a scheduling...")
    cmd := CommandsAvailable[sch.CommandID]
    cmd.ExecutableCommand.Execute("off")
}

func saveExec(cmd string, args ...interface{}) sql.Result {
    retval, err := db.Exec(cmd, args...)
    if err != nil {
        panic(err)
    }
    return retval
}