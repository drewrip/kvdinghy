package main

import(
	"fmt"
	"strconv"
	"os/exec"
	"os"
	"time"
	"github.com/fatih/color"
	"log"
	"flag"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"

)

type Data struct{
        Size            int

        //In Milliseconds
        TotalTime       int64

        Trials          int
}


func check(err error){
	if err != nil{
		log.Fatal(err)
	}
}

func main(){
	binary:="../kvdinghy"
    // Flags
	var sqlServerAddr string
	var trials int
	flag.StringVar(&sqlServerAddr, "s", "", "address of the SQL database to pump results into. This assumes your database has a corresponding table for the test you wish to run, 'benchmark'. Should be provided in the format: user:password@tcp(127.0.0.1:3306)/dbname")
	flag.IntVar(&trials, "n", 1, "number of trials to run, one trial is testing all odd sizes 3-51")
	flag.Parse()
	tablename := "benchmark"
		
	db, _ := sql.Open("mysql", sqlServerAddr) // Assumes MySQL db

	data:=make([]Data,0)
	// Making Data points
	for i:=3; i<53; i+=2{
		data = append(data, Data{Size: i, TotalTime: 0, Trials: 0})
	}
	for x:=0; x<trials; x++{
		ind:=0
		temptimes:=make([]int64, 0)
		for i:=3; i<53; i+=2{
			color.Yellow("[TEST] Starting for n=%d\n", i)
			startCmd := exec.Command(binary, "--test", "-s", strconv.Itoa(i))
			startCmd.Stderr = os.Stderr
			startCmd.Start()
			time.Sleep(5 * time.Second)
			cmd := exec.Command("./test.sh")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			start:=time.Now()
			cmd.Run()
			totalEl:=int64(time.Since(start))/1e6
			exec.Command("killall", "kvdinghy").Run()
			color.Blue("Took %dms", totalEl)
			color.Red("Shutting down cluster")

			// Process Results
			temptimes = append(temptimes, totalEl)
			data[ind].TotalTime += totalEl
			data[ind].Trials++
			ind++
			time.Sleep(500 * time.Millisecond)
		}

		// Pushing results to db if necessary after a full run of each size
		if sqlServerAddr != ""{
			for i:=0; i<len(data); i++{
				insert, err := db.Query(fmt.Sprintf("UPDATE %s SET time = time + %d, trials = trials + 1 WHERE size = %d", tablename, temptimes[i], data[i].Size))
				check(err)
				color.Cyan("WROTE RESULTS TO DATABASE")
				insert.Close()
			}
		}
	}

	fmt.Printf("Cluster Size:\tTest Time (ms):\n")
	for i:=0; i<len(data); i++{
		fmt.Printf("%d\t\t%d\n", data[i].Size, data[i].TotalTime/int64(data[i].Trials))
	}


}