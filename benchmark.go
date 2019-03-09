package main

import(
	//"fmt"
	//"net/http"
	"strconv"
	"os/exec"
	"os"
	"time"
	//"bytes"
	"github.com/fatih/color"
	"log"
	//"encoding/json"
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
	data := make([]Data, 0)
	binary:="./kvdinghy"
	for i:=3; i<53; i+=2{
            data = append(data, Data{Size: i, TotalTime: 0, Trials: 0})
    }

	for i:=3; i<53; i+=2{
		color.Green("Starting cluster n=%d", i)
		startCmd := exec.Command(binary, "--test", "-s", strconv.Itoa(i))
		startCmd.Stdout = os.Stdout
		startCmd.Stderr = os.Stderr
		startCmd.Start()
		time.Sleep(10 * time.Second)
		cmd := exec.Command("/home/drew/go/src/github.com/drewrip/kvdinghy/test.sh")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		exec.Command("killall", "kvdinghy").Run()
		color.Red("Shutting down cluster")
	}
	
	for i:=0; i<len(data); i++{
        fmt.Printf("%d\t\t%d\n", data[i].Size, data[i].TotalTime/int64(data[i].Trials))
    }

}