package cmd

import (
	"os"
	"os/exec"
	"runtime"
	"time"
)

func TaskRunShellCommandInChannel() error {
	for {
		cmdPath := <-chanShell
		if cmdPath == "" {
			continue
		}

		DebugInfo("RunShellCommandInChannel", Int2Str(len(chanShell)), ": == shell command ==: ", cmdPath)

		cmdBash := exec.Command("bash", "-c", cmdPath)
		if runtime.GOOS == "windows" {
			cmdBash = exec.Command("cmd", "/Q", "/C", cmdPath)
		}
		if IsDebug {
			out, err := cmdBash.CombinedOutput()
			if err != nil {
				PrintError("RunShellCommandInChannel.10", err)
			}
			DebugInfo("RunShellCommandInChannel.20", string(out))
		} else {
			_, err := cmdBash.CombinedOutput()
			if err != nil {
				PrintError("RunShellCommandInChannel.30", err)
			}
		}

		errRm := os.Remove(cmdPath)
		if errRm != nil {
			FilesToBeRemoved[cmdPath] = 1
		}

	}
}

func TaskDeleteFilesInFilesToBeRemoved() {
	for {
		time.Sleep(5 * time.Second)

		if len(FilesToBeRemoved) == 0 {
			continue
		}

		for cmdPath, v := range FilesToBeRemoved {
			if cmdPath == "" {
				continue
			}

			if v != 0 {
				finfo, err := os.Stat(cmdPath)
				if err == nil {
					tNow := time.Now()
					fage := tNow.Sub(finfo.ModTime())
					if fage < 10*time.Second {
						continue
					}

					errRm := os.Remove(cmdPath)
					if errRm == nil {
						FilesToBeRemoved[cmdPath] = 0
					} else {
						DebugWarn("TaskDeleteFilesInChannel", cmdPath, ": ", errRm.Error())
					}
				}
			} else {
				delete(FilesToBeRemoved, cmdPath)
			}
		}

		if len(FilesToBeRemoved) != 0 {
			DebugWarn("TaskDeleteFilesInChannel", "Cannot Be Removed(", len(FilesToBeRemoved), "): ", FilesToBeRemoved)
		} else {
			DebugInfo("TaskDeleteFilesInChannel", "ALL DONE")
		}

	}

}
