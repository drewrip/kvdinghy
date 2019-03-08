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

func check(err error){
	if err != nil{
		log.Fatal(err)
	}
}

func main(){
	binary:="/home/drew/go/src/github.com/drewrip/kvdinghy/kvdinghy"
	for i:=3; i<53; i+=2{
		color.Green("Starting cluster n=%d", i)
		exec.Command(binary, "--test", "-s", strconv.Itoa(i)).Start()
		time.Sleep(2 * time.Second)
		cmd := exec.Command("/home/drew/go/src/github.com/drewrip/kvdinghy/test.sh")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		exec.Command("killall", "kvdinghy").Run()
		color.Red("Shutting down cluster")
	}
	
}