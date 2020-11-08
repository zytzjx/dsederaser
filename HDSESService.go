// HDSESService
package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"net/http"
	"regexp"

	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type writer struct {
	mu *sync.Mutex
	wl *os.File
}

func (w *writer) Write(bytes []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	//fmt.Printf("%s\n", string(bytes))
	w.wl.WriteString(string(bytes) + "\n")
	return len(bytes), nil
}

type processlabel struct {
	cmddict map[int]*exec.Cmd
	mu      *sync.Mutex
}

func (pl *processlabel) Add(label int, cmd *exec.Cmd) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if _, ok := pl.cmddict[label]; ok {
		delete(pl.cmddict, label)
	}
	//fmt.Printf("Add: %d  %v\n", label, cmd)
	pl.cmddict[label] = cmd
}

func (pl *processlabel) Remove(label int) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if cmd, ok := pl.cmddict[label]; ok && cmd != nil {
		fmt.Printf("Remove: %d ready kill\n", label)
		cmd.Process.Kill()
		delete(pl.cmddict, label)
	}
}

// IsSSD check is SSD
func IsSSD(devicename string) bool {
	isSSD := false
	ss, _ := exec.Command("smartctl", "-i", devicename).Output()
	if strings.Contains(string(ss), "Solid State Device") {
		isSSD = true
	}
	return isSSD
}

func divmod(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

// RunExeWipe run dskwipe and handle output to database
func RunExeWipe(logpath string, devicename string, patten string, label int) error {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dskwipe := path.Join(dir, "dskwipe")
	fmt.Printf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten)
	Set(label, "starttasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
	cmd := exec.Command(dskwipe, devicename, "-y", "-n", "8000", patten)

	processlist.Add(label, cmd)

	f, err := os.OpenFile(fmt.Sprintf("%s/logs/%s/log_%d.log", os.Getenv("HDSESHOME"), logpath, label), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten))

	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan bool)

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)
	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			//HandleLog(label, line)
			handlelogprogress(label, line)
		}

		// We're all done, unblock the channel
		done <- true

	}()

	// Start the command and check for errors
	err = cmd.Start()

	// Wait for all output to be processed
	<-done

	// Wait for the command to finish
	err = cmd.Wait()
	return err
}

// RunSecureErase Run Secure Erase
func RunSecureErase(logpath string, devicename string, label int) {
	f, err := os.OpenFile(fmt.Sprintf("%s/logs/%s/log_%d.log", os.Getenv("HDSESHOME"), logpath, label), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	tstart := time.Now()
	f.WriteString(fmt.Sprintf("Start Task local time and date: %s\n", tstart.Format("Mon Jan _2 15:04:05 2006")))
	Set(label, "starttasktime", tstart.Format("Mon Jan _2 15:04:05 2006"), 0)
	stime := tstart.Format("15:04:05")
	funReadData := func() (string, error) {
		// if sector size is 520, this code is not working.Must use sglib. but not find go sglib.
		f, err := syscall.Open(devicename, syscall.O_RDONLY, 0777)
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		defer syscall.Close(f)
		b1 := make([]byte, 512)
		_, err = syscall.Read(f, b1)
		if err != nil {
			return "", err
		}
		md5 := md5.Sum(b1)
		ss := fmt.Sprintf("%x", md5)
		return ss, nil
	}

	funWriteData := func() error {
		// if sector size is 520, this code is not working.Must use sglib. but not find go sglib.
		f, err := syscall.Open(devicename, syscall.O_WRONLY, 0777)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer syscall.Close(f)
		b1 := make([]byte, 512)
		for i := 0; i < 512; i++ {
			b1[i] = 65
		}
		_, err = syscall.Write(f, b1)
		if err != nil {
			return err
		}
		return nil
	}

	var errorcode int
	smd5, err := funReadData()
	if err != nil {
		errorcode = 10
	}

	bverify := false

	if IsSSD(devicename) {
		funWriteData()
		f.WriteString(fmt.Sprintf("hdparm --user-master u --security-set-pass PASSFD %s\n", devicename))
		exec.Command("hdparm", "--user-master", "u", "--security-set-pass", "PASSFD", devicename).Output()
		f.WriteString(fmt.Sprintf("hdparm --user-master u --security-erase PASSFD %s\n", devicename))
		exec.Command("hdparm", "--user-master", "u", "--security-erase", "PASSFD", devicename).Output()
		tend := (int64)(time.Now().Sub(tstart).Seconds())
		hours, remainder := divmod(tend, 3600)
		minutes, seconds := divmod(remainder, 60)
		send := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		//   1      1 0x00   6.737%   6.737% 00:02:30 00:02:30 17:00:51 00002226   134.73   134.73
		line := fmt.Sprintf("   1      1 0x00 100.000%% 100.000%% %s %s %s %08d     0.00     0.00\n", send, send, stime, tend)
		f.WriteString(line)
		handlelogprogress(label, line)
		f.WriteString(fmt.Sprintf("end Task local time and date: %s\n", time.Now().Format("Mon Jan _2 15:04:05 2006")))
		//f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", 0))
		Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
		smd51, err := funReadData()
		if err != nil {
			errorcode = 10
		}
		if errorcode == 0 {
			bverify = smd51 == smd5
		}

	} else {
		f.WriteString(fmt.Sprintf("hdparm --yes-i-know-what-i-am-doing --sanitize-crypto-scramble %s\n", devicename))
		exec.Command("hdparm", "--yes-i-know-what-i-am-doing", "--sanitize-crypto-scramble", devicename).Output()
		time.Sleep(2 * time.Second)
		f.WriteString(fmt.Sprintf("hdparm --sanitize-status %s\n", devicename))
		exec.Command("hdparm", "--sanitize-status", devicename).Output()
		time.Sleep(2 * time.Second)
		exec.Command("hdparm", "--sanitize-status", devicename).Output()

		tend := (int64)(time.Now().Sub(tstart).Seconds())
		hours, remainder := divmod(tend, 3600)
		minutes, seconds := divmod(remainder, 60)
		send := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		//   1      1 0x00   6.737%   6.737% 00:02:30 00:02:30 17:00:51 00002226   134.73   134.73
		line := fmt.Sprintf("   1      1 0x00 100.000%% 100.000%% %s %s %s %08d     0.00     0.00\n", send, send, stime, tend)
		f.WriteString(line)
		handlelogprogress(label, line)
		f.WriteString(fmt.Sprintf("end Task local time and date: %s\n", time.Now().Format("Mon Jan _2 15:04:05 2006")))
		//f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", 0))
		Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)

		smd51, err := funReadData()
		if err != nil {
			errorcode = 10
		}
		if errorcode == 0 {
			bverify = smd51 != smd5
		}
	}

	if bverify && errorcode == 0 {
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", 0))
		Set(label, "errorcode", 0, 0)
	} else {
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", errorcode))
		Set(label, "errorcode", errorcode, 0)
	}
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}

func handlelogprogress(label int, line string) {
	var validlog = regexp.MustCompile(`(\d*\.\d*)%.*?(\d*\.\d*)%.*?(\d*\.\d*)$`)
	if !validlog.MatchString(line) {
		return
	}
	sp := func(r rune) bool {
		return r == '\t' || r == ' '
	}
	infos := strings.FieldsFunc(line, sp)
	//write database
	infos[9] = strings.Split(infos[9], ".")[0] + " MB/s"
	infos[5] = infos[5][:len(infos[5])-3]

	secondtotime := func(s string) string {
		ret := "00:01"
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return ret
		}
		ti := time.Duration(i) * time.Second
		ret = fmtDuration(ti)
		return ret
	}
	infos[8] = secondtotime(infos[8])

	goprogress := func(s string) string {
		ret := "0.01%"
		i, err := strconv.ParseFloat(s[:len(s)-1], 64)
		if err != nil {
			return ret
		}
		if i < 1.0 {
			i = 1.0
		} else if i > 99.0 {
			i = 99.0
		}
		ret = fmt.Sprintf("%.02f%%", i)
		return ret
	}
	infos[4] = goprogress(infos[4])

	fmt.Println(infos)
	if setProgressbar(label, infos) != nil {
		//print log
	}
}

// RunWipe call dskwipe
func RunWipe(logpath string, devicename string, patten string, label int) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dskwipe := path.Join(dir, "dskwipe")
	fmt.Printf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten)
	Set(label, "starttasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
	cmd := exec.Command(dskwipe, devicename, "-y", "-n", "8000", patten)

	processlist.Add(label, cmd)

	f, err := os.OpenFile(fmt.Sprintf("%s/logs/%s/log_%d.log", os.Getenv("HDSESHOME"), logpath, label), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s %s %s %s %s %s\n", dskwipe, devicename, "-y", "-n", "8000", patten))

	var mu sync.Mutex

	cmd.Stderr = &writer{
		mu: &mu,
		wl: f,
	}
	cmd.Stdout = &writer{
		mu: &mu,
		wl: f,
	}
	/*
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		err = cmd.Wait()
		if err != nil {
			log.Fatal(err)
		}
	*/
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
			f.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("WipeExitCode=%d\n", waitStatus.ExitStatus())
			f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", waitStatus.ExitStatus()))
			Set(label, "errorcode", waitStatus.ExitStatus(), 0)
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("WipeExitCode=%d\n", waitStatus.ExitStatus())
		f.WriteString(fmt.Sprintf("WipeExitCode=%d\n", waitStatus.ExitStatus()))
		Set(label, "errorcode", waitStatus.ExitStatus(), 0)
	}
	Set(label, "endtasktime", time.Now().Format("Mon Jan _2 15:04:05 2006"), 0)
}

var mu sync.Mutex
var processlist *processlabel

var configxmldata *configs

func main() {
	fmt.Println("hdsesserver version: 20.10.25.0, auther:Jeffery Zhang")
	runtime.GOMAXPROCS(4)

	processlist = &processlabel{
		cmddict: make(map[int]*exec.Cmd),
		mu:      &mu,
	}

	LoadConfigXML()
	StartTCPServer()
	return
	/*
		r := mux.NewRouter()
		r.HandleFunc("/start/{label:[0-9]+}", handlerStartByLabel).Methods("POST") //.Queries("name", "{name}", "index", "{index:[0-9]+}", "folder", "{folder}")
		r.HandleFunc("/stop/{label:[0-9]+}", handlerStopByLabel)

		srv := &http.Server{
			Handler: r,
			Addr:    ":8000",
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		fmt.Println(srv)
		srv.ListenAndServe()
	*/
}

//name={name}&&index={index}&&folder={folder}
func handlerStartByLabel(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	decoder := json.NewDecoder(r.Body)
	var cmdinfo map[string]interface{}
	err := decoder.Decode(&cmdinfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	w.WriteHeader(http.StatusOK)

	var name, folder, sdevname string
	var index, label int
	Is512Sector := false
	if value, ok := cmdinfo["s512"]; ok {
		Is512Sector = value.(bool)
	}
	if value, ok := cmdinfo["name"]; ok {
		name = value.(string)
	}
	if value, ok := cmdinfo["folder"]; ok {
		folder = value.(string)
	}
	if value, ok := cmdinfo["device"]; ok {
		sdevname = value.(string)
	}
	if value, ok := cmdinfo["index"]; ok {
		index = int(value.(float64))
	}
	if value, ok := cmdinfo["label"]; ok {
		label = int(value.(float64))
	}

	vars := mux.Vars(r)

	if value, ok := vars["label"]; ok {
		label, _ = strconv.Atoi(value)
	}

	//fmt.Printf("%v_%s_%s_%s_%d_%d\n", Is512Sector, name, folder, sdevname, index, label)
	if Is512Sector {
		profile, err := configxmldata.FindProfileByName(name)
		if err != nil {
			w.Write(msgError)
			return
		}
		patten := profile.CreatePatten()
		go RunWipe(folder, sdevname, patten, label)
	} else {
		profile, err := configxmldata.FindProfileByName(name)
		if err != nil {
			w.Write(msgError)
			return
		}
		patten := profile.CreatePatten()
		sdevname = fmt.Sprintf("/dev/sg%d", index)
		go RunWipe(folder, sdevname, patten, label)
	}
	w.Write(msgOK)
	return
}

func handlerStopByLabel(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	vars := mux.Vars(r)
	var label int

	if value, ok := vars["label"]; ok {
		label, _ = strconv.Atoi(value)
	}
	go func() {
		processlist.Remove(label)
	}()

	w.Write(msgOK)

	return
}
