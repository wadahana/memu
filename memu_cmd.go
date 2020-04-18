package memu

import (
	"os/exec"
	"path"
	"strings"
	"strconv"
	"fmt"
	"os"

	"github.com/wadahana/memu/log"
)

const (

//	MEMUC_EXECUTOR  string = "memuc"
//	MEMU_EXECUTOR   string = "memu"
	MEMUC_EXECUTOR  string = "memuc.exe"
	MEMU_EXECUTOR   string = "memu.exe"

	SEPARATOR       string = "\\"
	CTRL            string = "\r\n"
) 

const (
	AndroidRomUnsupport = 0
	AndroidRomV44       = 44
	AndroidRomV51       = 51
	AndroidRomV71       = 71
)

type MEmuCmd struct {
	memucPath string
	memuPath  string
	dumpCommand bool
}

var Cmd MEmuCmd;

func initMEmuCmd(memuPath string) {
	Cmd.memucPath = path.Join(memuPath, MEMUC_EXECUTOR)
	_, err := os.Stat(Cmd.memucPath)
	if err != nil {
		panic(err)
	}
	
	Cmd.memuPath = path.Join(memuPath, MEMU_EXECUTOR)
	_, err = os.Stat(Cmd.memuPath)
	if err != nil {
		panic(err)
	}
	Cmd.dumpCommand = true;
	return;
}


func (cmd MEmuCmd)dump(format string, args ...interface{}) {
	if cmd.dumpCommand {
		log.Debugf(format, args...)
	}
}

func (cmd MEmuCmd)checkReuslt(lines []string) *MEmuError {
	if len(lines) > 0 {
		if strings.HasPrefix(lines[0], "SUCCESS") {
			return nil;
		} else if strings.HasPrefix(lines[0], "ERROR") {
			return NewError(RC_MemucError, strings.TrimSpace(lines[0]))
		}
	}
	return ErrorCommandResultMalformat;
}

func (cmd MEmuCmd)checkVirtualKey(key string) bool {
	if strings.EqualFold(key, "volumeup") || 
	   strings.EqualFold(key, "volumedown") || 
	   strings.EqualFold(key, "back") ||
	   strings.EqualFold(key, "home") ||
	   strings.EqualFold(key, "memu") {
	   	return true
	   }
	return false
}

func (cmd MEmuCmd)execMemuc(args ...string) *MEmuError {

	command := exec.Command(cmd.memucPath, args...)
	cmd.dump("command: %s", command)
	
	out, err := command.CombinedOutput()  
    if err != nil {  
    	log.Errorf("exec %s %s fail, error: %s", MEMUC_EXECUTOR, args[0], err)
    	return NewError(RC_SystemError, err.Error())
    }
    result := string(out)
    cmd.dump("output: %s", result)
    lines := strings.Split(result, CTRL)
 	return cmd.checkReuslt(lines)
}

func (cmd MEmuCmd)execMemu(args ...string) *MEmuError {

	command := exec.Command(cmd.memuPath, args...)
	cmd.dump("command: %s", command)
	
	err := command.Run()  
    if err != nil {  
    	if strings.HasPrefix(err.Error(), "exit status") {
    		return nil;
    	}
    	log.Errorf("exec %s fail, error: %s", MEMU_EXECUTOR, err)
    	return NewError(RC_SystemError, err.Error())
    }
    return nil
}

func (cmd MEmuCmd)LookupByName(name string, running bool) (*MEmuInfo, *MEmuError) {
	list, err := cmd.List(running)
	log.Debugf("LookupByName-> list: %v, err: %v", list, err);
	if err == nil {
		for _, info := range *list {
			if strings.Compare(name, info.Name) == 0{
				return &info, err
			}
		}
	}
	if err == nil {
		err = ErrorEmulatorNotFound
	}
	return nil, err
}

func (cmd MEmuCmd)LookupByIndex(index int, running bool) (*MEmuInfo, *MEmuError) {
	list, err := cmd.List(running)
	log.Debugf("LookupByIndex-> list: %v, err: %v", list, err);
	if err == nil {
		for _, info := range *list {
			if info.Index == index {
				return &info, err
			}
		}
	}
	if err == nil {
		err = ErrorEmulatorNotFound
	}
	return nil, err
}

func (cmd MEmuCmd)StartMiracast(name string) *MEmuError {
	return cmd.execMemu(name, "startmiracast", "0","0","1")
}

func (cmd MEmuCmd)StopMiracast(name string) *MEmuError {
	return cmd.execMemu(name, "stopmiracast")
}


func (cmd MEmuCmd)Create(version int) (int, string, *MEmuError) {
	
	command := exec.Command(cmd.memucPath, "create", fmt.Sprintf("%d", version))
	cmd.dump("command: %s", cmd)

	out, err := command.CombinedOutput()  
    if err != nil {  
    	log.Error(err)
    	return -1, "", NewError(RC_SystemError, err.Error())
    }
    result := string(out)
    cmd.dump("output: %s", result)

 	lines := strings.Split(result, CTRL)
 	meErr := cmd.checkReuslt(lines)
 
 	if  meErr == nil {
 		if len(lines) >= 2 && strings.HasPrefix(lines[1], "index") {
 			index, err := strconv.Atoi(lines[1][6:])
 			if err != nil {
 				return -1, "", NewError(RC_SystemError, err.Error())
 			}
 			name := ""
 			if index == 0 {
 				name = "MEmu"
 			} else {
 				name = fmt.Sprintf("MEmu_%d", index)
 			}
 			return index, name, cmd.RenameById(index, name)
 		} else {
 			meErr = ErrorCommandResultMalformat;
 		}
 	}
 	return -1, "", meErr;
}

func (cmd MEmuCmd)List(running bool) (*[]MEmuInfo, *MEmuError) {
	var command *exec.Cmd = nil
	if running {
		command = exec.Command(cmd.memucPath, "listvms", "--running", "-s")
	} else {
		command = exec.Command(cmd.memucPath, "listvms", "-s")
	}

	cmd.dump("command: %s", command)
	out, err := command.CombinedOutput()  
    if err != nil {  
    	log.Error(err)
    	return nil, NewError(RC_SystemError, err.Error())
    }
    result := string(out)
    cmd.dump("output: %s", result)

    if strings.HasPrefix(result, "ERROR") {
    	return nil, NewError(RC_MemucError, strings.TrimSpace(result))
    }
 	lines := strings.Split(result, CTRL)
 	if len(lines) <= 0 {
 		return nil, ErrorNotEmulator
 	}
 	list := new([]MEmuInfo)
 	for _, line := range lines {
 		//line = strings.TrimSpace(line)
 		log.Debugf("line: [%s]", line)
 		token := strings.Split(line, ",")
 		if len(token) != 6 {
 			log.Warnf("malformat vminfo: %s", line)
 			continue;
 		}
 		info := MEmuInfo{};
 		info.Index, err = strconv.Atoi(token[0])
 		if err != nil {
 			log.Warnf("unknown vm'index: %s", line)
 			continue;
 		}
 		info.Name = token[1];
 		info.Storage, err = strconv.ParseInt(token[5], 10, 64);
 		log.Debugf("i: %d, storage: %d, err: %v", info.Index, info.Storage, err)
 		info.Running = strings.EqualFold(token[3], "1")

 		*list = append(*list, info)
 	}
 	if len(*list) > 0 {
 		return list, nil
 	}
	return nil, ErrorNotEmulator
}


func (cmd MEmuCmd)RemoveByName(name string) *MEmuError {
	if len(name) == 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("remove", "-n", name)
}

func (cmd MEmuCmd)RemoveById(id int) *MEmuError {
	if id < 0 {
		return ErrorInvalidArgument 
	}
	return cmd.execMemuc("remove", "-i", strconv.Itoa(id))
}

func (cmd MEmuCmd)CloneByName(name string) *MEmuError {
	if len(name) == 0 {
		return ErrorInvalidArgument
	}
	return nil
}

func (cmd MEmuCmd)CloneById(id int) *MEmuError {
	if id < 0 {
		return ErrorInvalidArgument
	}
	return nil;
}

func (cmd MEmuCmd)RenameByName(oldName string, newName string) *MEmuError {
	if len(newName) == 0 || len(oldName) == 0{
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("-n", oldName, "rename", newName)
}

func (cmd MEmuCmd)RenameById(id int, newName string) *MEmuError {
	if id < 0 && len(newName) < 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("-i", strconv.Itoa(id), "rename", newName)
}

func (cmd MEmuCmd)CheckVMRunningByName(name string) (bool, *MEmuError) {
	if len(name) == 0 {
		return false, ErrorInvalidArgument
	}
	command := exec.Command(cmd.memucPath, "isvmrunning", "-n", name)
	cmd.dump("command: %s", cmd)
	out, err := command.CombinedOutput()  
    if err != nil {  
    	log.Error(err)
    	return false, NewError(RC_SystemError, err.Error())
    }
    result := string(out)
    cmd.dump("output: %s", result)

    if strings.HasPrefix(result, "Running") {
    	log.Debugf("running");
    	return true, nil
    } else if strings.HasPrefix(result, "Not Running") {
    	log.Debugf("not running");
    	return false, nil
    }
    return false, NewError(RC_MemucError, strings.TrimSpace(result))
}

func (cmd MEmuCmd)CheckVMRunningById(id int) (bool, *MEmuError) {
	if id < 0 || id > 40  {
		return false, ErrorInvalidArgument
	}
	command := exec.Command(cmd.memucPath, "isvmrunning", "-i", strconv.Itoa(id))
	cmd.dump("command: %s", cmd)
	out, err := command.CombinedOutput()  
    if err != nil {  
    	log.Error(err)
    	return false, NewError(RC_SystemError, err.Error())
    }
    result := string(out)
    cmd.dump("output: %s", result)

    if strings.HasPrefix(result, "Running") {
    	log.Debugf("running");
    	return true, nil
    } else if strings.HasPrefix(result, "Not Running") {
    	log.Debugf("not running");
    	return false, nil
    }
    return false, NewError(RC_MemucError, strings.TrimSpace(result))
}


func (cmd MEmuCmd)StartByName(name string) *MEmuError {
	if len(name) == 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("start", "-n", name)
}

func (cmd MEmuCmd)StartById(id int) *MEmuError {
	if id < 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("start", "-i", strconv.Itoa(id))
}

func (cmd MEmuCmd)StopByName(name string) *MEmuError {
	if len(name) == 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("stop", "-n", name)
}

func (cmd MEmuCmd)StopById(id int) *MEmuError {
	if id < 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("stop", "-i", strconv.Itoa(id))
}

func (cmd MEmuCmd)RebootByName(name string) *MEmuError {
	if len(name) == 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("reboot", "-n", name)
}

func (cmd MEmuCmd)RebootById(id int) *MEmuError {
	if id < 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("reboot", "-i", strconv.Itoa(id))
}


func (cmd MEmuCmd)SendKeyByName(name string, key string) *MEmuError {
	if len(name) == 0 || !cmd.checkVirtualKey(key) {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("-n", name, "sendkey", key)
}

func (cmd MEmuCmd)SendKeyById(id int, key string) *MEmuError {
	if id < 0 || !cmd.checkVirtualKey(key) {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("-i", strconv.Itoa(id), "sendkey", key)
}

func (cmd MEmuCmd)ShakeByName(name string) *MEmuError {
	if len(name) == 0  {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("-n", name, "shake")
}

func (cmd MEmuCmd)ShakeById(id int) *MEmuError {
	if id < 0 {
		return ErrorInvalidArgument
	}
	return cmd.execMemuc("-i", strconv.Itoa(id), "shake")
}

