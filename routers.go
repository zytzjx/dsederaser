package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/gorilla/mux"
)

//standard={name}
func startTaskHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	label, _ := strconv.Atoi(mux.Vars(r)["label"])
	name := r.FormValue("standard")

	Is512Sector := false

	folder := path.Join(os.ExpandEnv("HDSESHOME"), "logs")

	sdevname, err := GetString(label, "linuxname")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(msgError)
		return
	}
	if len(sdevname) > 0 {
		exec.Command("umount", sdevname).Output()
	}
	sgName, err := GetString(label, "sglibname")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(msgError)
		return
	}

	fmt.Printf("%v_%s_%s_%s_%s_%d\n", Is512Sector, name, folder, sdevname, sgName, label)
	if name == "SecureErase" {
		go RunSecureErase(folder, sdevname, label)
		return
	}
	if Is512Sector && len(sdevname) > 0 {
		profile, err := configxmldata.FindProfileByName(name)
		if err != nil {
			return
		}
		patten := profile.CreatePatten()
		go RunExeWipe(folder, sdevname, patten, label)
	} else {
		profile, err := configxmldata.FindProfileByName(name)
		if err != nil {
			return
		}
		patten := profile.CreatePatten()

		go RunExeWipe(folder, sgName, patten, label)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(msgOK)
	return
}

func stopTaskHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
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
