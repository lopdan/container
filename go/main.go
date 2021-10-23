package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run(os.Args[2:]...)
	case "child":
		child(os.Args[2:]...)
	default:
		log.Fatal("Unknown command. Use run <command_name>, like `run /bin/bash` or `run echo hello`")
	}
}

func run(command ...string) {
	log.Println("Executing", command, "from run")
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, command[0:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func child(command ...string) {
	log.Println("Executing", command, "from child")

	cg()

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("./debian_fs"))
	must(os.Chdir("./debian_fs"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	// Temporary filesystem
	must(syscall.Mount("something", "mytemp", "tmpfs", 0, ""))

	must(cmd.Run())

	// Cleanup mount
	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("mytemp", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"

	mem := filepath.Join(cgroups, "memory")
	kontainer := filepath.Join(mem, "kontainer")
	os.Mkdir(kontainer, 0755)
	must(ioutil.WriteFile(filepath.Join(kontainer, "memory.limit_in_bytes"), []byte("999424"), 0700))
	must(ioutil.WriteFile(filepath.Join(kontainer, "notify_on_release"), []byte("1"), 0700))

	pid := strconv.Itoa(os.Getpid())
	must(ioutil.WriteFile(filepath.Join(kontainer, "cgroup.procs"), []byte(pid), 0700))
}

func must(err error){
	if err != nil {
		panic(err)
	}
}