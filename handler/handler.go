package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/laohanlinux/riot/command"
	"github.com/laohanlinux/riot/fsm"

	"github.com/laohanlinux/go-logger/logger"
	"github.com/laohanlinux/mux"
)

// RiotHandler ...
type RiotHandler struct{}

// ServeHTTP .
func (rh *RiotHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var value []byte
	var err error
	var errType string
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		errType, value, err = getValue(w, r)
		if err != nil {
			logger.Error(err)
		}
	case "DELETE":
		errType, err = delValue(w, r)
		if err != nil {
			logger.Error(err)
		}
	case "POST":
		errType, err = setValue(w, r)
		if err != nil {
			logger.Error(errType, err)
		}
	default:
		errType = InvalidRequest
	}
	msg := msgErrCodeMap[errType]
	if msg.httpCode == 200 {
		w.Write(value)
		return
	}
	w.WriteHeader(msg.httpCode)
	w.Write(msg.toJsonBytes())
}

func getValue(w http.ResponseWriter, r *http.Request) (string, []byte, error) {
	vars := mux.Vars(r)
	key := vars["key"]
	bucket := vars["bucket"]

	cmd := command.Command{
		Op:     command.CmdGet,
		Key:    key,
		Bucket: bucket,
	}
	if len(cmd.Key) == 0 {
		return InvalidKey, nil, fmt.Errorf("The request key is Empty")
	}

	qs := command.QsRandom
	var err error
	//Query strategires
	qsValue := r.URL.Query().Get("qs")
	if qsValue == "" {
		qs, err = strconv.Atoi(qsValue)
		if err != nil {
			return QsInvalid, nil, err
		}
	}
	var value []byte
	value, err = cmd.DoGet(qs)
	if err != nil && err != fsm.ErrNotFound {
		return OpErr, value, err
	}
	if err == fsm.ErrNotFound {
		return NotFound, nil, nil
	}
	return Ok, value, nil
}

func setValue(w http.ResponseWriter, r *http.Request) (string, error) {
	vars := mux.Vars(r)
	key := vars["key"]
	bucket := vars["bucket"]
	value, err := ioutil.ReadAll(r.Body)
	if err != nil || value == nil || len(value) == 0 {
		return InvalidRequest, err
	}
	cmd := command.Command{
		Op:     command.CmdSet,
		Bucket: bucket,
		Key:    key,
		Value:  value,
	}
	err = cmd.DoSet()
	if err != nil {
		return InternalErr, err
	}
	return Ok, nil
}

func delValue(w http.ResponseWriter, r *http.Request) (string, error) {
	vars := mux.Vars(r)
	key := vars["key"]
	bucket := vars["bucket"]
	cmd := command.Command{
		Op:     command.CmdDel,
		Bucket: bucket,
		Key:    key,
	}

	err := cmd.DoDel()
	if err != nil && err != fsm.ErrNotFound {
		return OpErr, err
	}

	if err == fsm.ErrNotFound {
		return NotFound, nil
	}
	return Ok, nil
}
